package vhost

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// VirtualHost thông tin virtual host
type VirtualHost struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Root   string `json:"root"`
	SSL    bool   `json:"ssl"`
	Active bool   `json:"active"`
}

// Manager quản lý virtual hosts
type Manager struct {
	configPath string
	hosts      []VirtualHost
}

func NewManager(configPath string) *Manager {
	m := &Manager{configPath: configPath}
	m.load()
	return m
}

func (m *Manager) GetAll() []VirtualHost {
	return m.hosts
}

func (m *Manager) Add(name, domain, root string, ssl bool) error {
	for _, h := range m.hosts {
		if h.Domain == domain {
			return fmt.Errorf("domain %s already exists", domain)
		}
	}

	host := VirtualHost{
		Name:   name,
		Domain: domain,
		Root:   root,
		SSL:    ssl,
		Active: true,
	}

	if err := m.writeApacheConfig(host); err != nil {
		return err
	}

	if err := m.addToHostsFile(domain); err != nil {
		return err
	}

	m.hosts = append(m.hosts, host)
	return m.save()
}

func (m *Manager) Remove(domain string) error {
	for i, h := range m.hosts {
		if h.Domain == domain {
			m.removeApacheConfig(h)
			m.removeFromHostsFile(domain)
			m.hosts = append(m.hosts[:i], m.hosts[i+1:]...)
			return m.save()
		}
	}
	return fmt.Errorf("domain %s not found", domain)
}

func (m *Manager) writeApacheConfig(h VirtualHost) error {
	confDir := filepath.Join(m.configPath, "vhosts")
	os.MkdirAll(confDir, 0755)

	confFile := filepath.Join(confDir, h.Domain+".conf")
	content := fmt.Sprintf(`<VirtualHost *:80>
    ServerName %s
    DocumentRoot "%s"
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>
`, h.Domain, h.Root, h.Root)

	if h.SSL {
		content += fmt.Sprintf(`
<VirtualHost *:443>
    ServerName %s
    DocumentRoot "%s"
    SSLEngine on
    SSLCertificateFile    "%s/%s.crt"
    SSLCertificateKeyFile "%s/%s.key"
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>
`, h.Domain, h.Root,
			confDir, h.Domain,
			confDir, h.Domain,
			h.Root)
	}

	return os.WriteFile(confFile, []byte(content), 0644)
}

func (m *Manager) removeApacheConfig(h VirtualHost) {
	confDir := filepath.Join(m.configPath, "vhosts")
	os.Remove(filepath.Join(confDir, h.Domain+".conf"))
}

func (m *Manager) hostsFilePath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

func (m *Manager) addToHostsFile(domain string) error {
	path := m.hostsFilePath()
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	entry := fmt.Sprintf("127.0.0.1\t%s", domain)
	if strings.Contains(string(content), entry) {
		return nil // already exists
	}

	// Try direct write first (works when app runs as admin)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		_, err = fmt.Fprintf(f, "\n%s\n", entry)
		return err
	}

	// On Windows, re-try with a UAC-elevated PowerShell process
	if runtime.GOOS == "windows" {
		script := fmt.Sprintf(
			"Add-Content -Path 'C:\\Windows\\System32\\drivers\\etc\\hosts' -Value \"`r`n127.0.0.1`t%s\"\r\n",
			domain,
		)
		return m.runElevated(script)
	}
	return fmt.Errorf("need admin privileges to modify hosts file: %w", err)
}

func (m *Manager) removeFromHostsFile(domain string) {
	path := m.hostsFilePath()
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var filtered []string
	for _, line := range lines {
		if !strings.Contains(line, domain) {
			filtered = append(filtered, line)
		}
	}
	newContent := strings.Join(filtered, "\n")

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil && runtime.GOOS == "windows" {
		// Rewrite via elevation
		script := fmt.Sprintf(
			"$h='C:\\Windows\\System32\\drivers\\etc\\hosts'; (Get-Content $h) | Where-Object { $_ -notlike '*%s*' } | Set-Content $h\r\n",
			domain,
		)
		m.runElevated(script) //nolint:errcheck
	}
}

// runElevated writes psScript to a temp .ps1 file and executes it in an
// elevated PowerShell process via UAC (Start-Process -Verb RunAs).
func (m *Manager) runElevated(psScript string) error {
	tmp, err := os.CreateTemp("", "stacknest_*.ps1")
	if err != nil {
		return fmt.Errorf("cannot create temp script: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString(psScript); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	// Start an outer non-elevated PowerShell that launches an elevated inner one
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf(`Start-Process powershell -Verb RunAs -ArgumentList '-NoProfile -ExecutionPolicy Bypass -NonInteractive -File \"%s\"' -Wait`, tmpPath),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("hosts file update failed (admin required): %v — %s", err, out)
	}
	return nil
}

func (m *Manager) load() {
	data, err := os.ReadFile(filepath.Join(m.configPath, "vhosts.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, &m.hosts)
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.hosts, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(m.configPath, 0755)
	return os.WriteFile(filepath.Join(m.configPath, "vhosts.json"), data, 0644)
}
