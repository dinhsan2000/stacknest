package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CertInfo thông tin về một SSL certificate
type CertInfo struct {
	Domain    string `json:"domain"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
	ExpiresAt string `json:"expires_at"`
}

// Manager quản lý SSL certificates
type Manager struct {
	rootPath string
}

func NewManager(rootPath string) *Manager {
	return &Manager{rootPath: rootPath}
}

// sslDir trả về thư mục lưu CA
func (m *Manager) sslDir() string {
	return filepath.Join(m.rootPath, "ssl")
}

// vhostsDir trả về thư mục vhosts (nơi domain certs được lưu)
func (m *Manager) vhostsDir() string {
	return filepath.Join(m.rootPath, "vhosts")
}

// CACertPath trả về đường dẫn CA certificate
func (m *Manager) CACertPath() string {
	return filepath.Join(m.sslDir(), "ca.crt")
}

func (m *Manager) caKeyPath() string {
	return filepath.Join(m.sslDir(), "ca.key")
}

// EnsureCA tạo CA nếu chưa tồn tại
func (m *Manager) EnsureCA() error {
	if _, err := os.Stat(m.CACertPath()); err == nil {
		return nil // CA đã tồn tại
	}

	os.MkdirAll(m.sslDir(), 0755)

	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate CA key: %w", err)
	}

	// CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Stacknest Local CA",
			Organization: []string{"Stacknest"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	// Self-sign CA cert
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create CA cert: %w", err)
	}

	// Write CA cert
	if err := writePEM(m.CACertPath(), "CERTIFICATE", caDER); err != nil {
		return err
	}

	// Write CA key
	keyDER, err := x509.MarshalPKCS8PrivateKey(caKey)
	if err != nil {
		return fmt.Errorf("marshal CA key: %w", err)
	}
	return writePEM(m.caKeyPath(), "PRIVATE KEY", keyDER)
}

// GenerateCert tạo domain certificate được ký bởi CA
func (m *Manager) GenerateCert(domain string) (certPath, keyPath string, err error) {
	if err = m.EnsureCA(); err != nil {
		return
	}

	// Load CA
	caCert, caKey, err := m.loadCA()
	if err != nil {
		return
	}

	// Generate domain key
	domainKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("generate domain key: %w", err)
		return
	}

	// Domain cert template
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{"Stacknest"},
		},
		DNSNames:              []string{domain, "*." + domain},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &domainKey.PublicKey, caKey)
	if err != nil {
		err = fmt.Errorf("create domain cert: %w", err)
		return
	}

	// Write to vhosts dir (where vhost manager expects them)
	os.MkdirAll(m.vhostsDir(), 0755)
	certPath = filepath.Join(m.vhostsDir(), domain+".crt")
	keyPath = filepath.Join(m.vhostsDir(), domain+".key")

	if err = writePEM(certPath, "CERTIFICATE", certDER); err != nil {
		return
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(domainKey)
	if err != nil {
		err = fmt.Errorf("marshal domain key: %w", err)
		return
	}
	err = writePEM(keyPath, "PRIVATE KEY", keyDER)
	return
}

// GetCerts trả về danh sách domain certs trong vhosts dir
func (m *Manager) GetCerts() []CertInfo {
	entries, err := os.ReadDir(m.vhostsDir())
	if err != nil {
		return nil
	}

	var certs []CertInfo
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".crt") {
			continue
		}
		domain := strings.TrimSuffix(e.Name(), ".crt")
		certPath := filepath.Join(m.vhostsDir(), e.Name())

		expiry := certExpiry(certPath)
		certs = append(certs, CertInfo{
			Domain:    domain,
			CertPath:  certPath,
			KeyPath:   filepath.Join(m.vhostsDir(), domain+".key"),
			ExpiresAt: expiry,
		})
	}
	return certs
}

// RemoveCert xóa cert và key của một domain
func (m *Manager) RemoveCert(domain string) error {
	certPath := filepath.Join(m.vhostsDir(), domain+".crt")
	keyPath := filepath.Join(m.vhostsDir(), domain+".key")
	os.Remove(certPath)
	os.Remove(keyPath)
	return nil
}

// IsCAInstalled kiểm tra CA có đang được trust trong OS không
func (m *Manager) IsCAInstalled() bool {
	caCertPath := m.CACertPath()
	if _, err := os.Stat(caCertPath); err != nil {
		return false // CA chưa được tạo
	}

	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("certutil", "-store", "-user", "Root").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "Stacknest Local CA")

	case "darwin":
		out, err := exec.Command("security", "find-certificate", "-c", "Stacknest Local CA", "-a").Output()
		if err != nil {
			return false
		}
		return len(out) > 0

	default: // Linux
		dest := "/usr/local/share/ca-certificates/stacknest-ca.crt"
		_, err := os.Stat(dest)
		return err == nil
	}
}

// TrustCA cài CA vào system trust store
func (m *Manager) TrustCA() error {
	if err := m.EnsureCA(); err != nil {
		return err
	}
	caCertPath := m.CACertPath()

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("certutil", "-addstore", "-user", "Root", caCertPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("certutil failed: %s", string(out))
		}
		return nil

	case "darwin":
		cmd := exec.Command("security", "add-trusted-cert",
			"-d", "-r", "trustRoot",
			"-k", os.ExpandEnv("$HOME/Library/Keychains/login.keychain"),
			caCertPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("security add-trusted-cert failed: %s", string(out))
		}
		return nil

	default: // Linux
		dest := "/usr/local/share/ca-certificates/stacknest-ca.crt"
		data, err := os.ReadFile(caCertPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("need sudo to install CA (copy to %s): %w", dest, err)
		}
		out, err := exec.Command("update-ca-certificates").CombinedOutput()
		if err != nil {
			return fmt.Errorf("update-ca-certificates failed: %s", string(out))
		}
		return nil
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (m *Manager) loadCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(m.CACertPath())
	if err != nil {
		return nil, nil, fmt.Errorf("read CA cert: %w", err)
	}
	keyPEM, err := os.ReadFile(m.caKeyPath())
	if err != nil {
		return nil, nil, fmt.Errorf("read CA key: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA cert: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	keyIface, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA key: %w", err)
	}
	rsaKey, ok := keyIface.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("CA key is not RSA")
	}

	return cert, rsaKey, nil
}

func writePEM(path, pemType string, derBytes []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: pemType, Bytes: derBytes})
}

func certExpiry(certPath string) string {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return ""
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return ""
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ""
	}
	return cert.NotAfter.Format("2006-01-02")
}
