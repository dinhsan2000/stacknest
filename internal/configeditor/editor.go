package configeditor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// ConfigFile đại diện cho một file cấu hình
type ConfigFile struct {
	Service  string `json:"service"`  // "apache", "mysql", "php", "nginx"
	Label    string `json:"label"`    // "httpd.conf", "php.ini"
	Path     string `json:"path"`     // đường dẫn đầy đủ
	Lang     string `json:"lang"`     // "apache", "ini", "nginx" — dùng cho syntax highlight
	Writable bool   `json:"writable"` // có thể ghi không
}

// BackupInfo thông tin về một bản backup
type BackupInfo struct {
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
	SizeBytes int64  `json:"size_bytes"`
}

// Manager quản lý config files
type Manager struct {
	rootPath string
}

func NewManager(rootPath string) *Manager {
	return &Manager{rootPath: rootPath}
}

// GetConfigFiles trả về danh sách config files của một service
func (m *Manager) GetConfigFiles(service string) []ConfigFile {
	switch strings.ToLower(service) {
	case "apache":
		return m.apacheConfigs()
	case "nginx":
		return m.nginxConfigs()
	case "mysql":
		return m.mysqlConfigs()
	case "php":
		return m.phpConfigs()
	default:
		return nil
	}
}

// ReadFile đọc nội dung file
func (m *Manager) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}
	return string(data), nil
}

// SaveFile lưu nội dung file (tự tạo backup trước)
func (m *Manager) SaveFile(path, content string) error {
	// Tạo backup trước khi ghi
	if _, err := os.Stat(path); err == nil {
		_ = m.createBackup(path)
	}

	// Ghi file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	return nil
}

// GetBackups trả về danh sách backups của một file
func (m *Manager) GetBackups(path string) []BackupInfo {
	backupDir := m.backupDir(path)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil
	}

	base := filepath.Base(path)
	var backups []BackupInfo
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), base+".bak.") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Path:      filepath.Join(backupDir, e.Name()),
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			SizeBytes: info.Size(),
		})
	}

	// Mới nhất trước
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})
	return backups
}

// RestoreBackup khôi phục từ backup
func (m *Manager) RestoreBackup(backupPath, targetPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	return os.WriteFile(targetPath, data, 0644)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (m *Manager) createBackup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	backupDir := m.backupDir(path)
	os.MkdirAll(backupDir, 0755)

	ts := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, filepath.Base(path)+".bak."+ts)
	return os.WriteFile(backupPath, data, 0644)
}

func (m *Manager) backupDir(path string) string {
	return filepath.Join(m.rootPath, ".config_backups", filepath.Base(filepath.Dir(path)))
}

func (m *Manager) isWritable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func (m *Manager) configFile(service, label, path, lang string) ConfigFile {
	return ConfigFile{
		Service:  service,
		Label:    label,
		Path:     path,
		Lang:     lang,
		Writable: m.isWritable(path),
	}
}

// ─── Service config paths ─────────────────────────────────────────────────────

// scanVersionedDirs quét bin/{service}/{version}/{relPath} cho mỗi version đã cài.
// Label = filename + " (version)".
func (m *Manager) scanVersionedDirs(service, relPath, lang string) []ConfigFile {
	base := filepath.Join(m.rootPath, "bin", service)
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	var configs []ConfigFile
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(base, e.Name(), relPath)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		label := filepath.Base(p) + " (" + e.Name() + ")"
		configs = append(configs, m.configFile(service, label, p, lang))
	}
	return configs
}

func (m *Manager) apacheConfigs() []ConfigFile {
	// Quét bin/apache/{version}/conf/httpd.conf
	configs := m.scanVersionedDirs("apache", filepath.Join("conf", "httpd.conf"), "apache")

	// Fallback system paths (non-Windows)
	if runtime.GOOS != "windows" {
		for _, p := range []string{
			"/usr/local/etc/httpd/httpd.conf",
			"/opt/homebrew/etc/httpd/httpd.conf",
			"/etc/apache2/apache2.conf",
			"/etc/httpd/conf/httpd.conf",
		} {
			if _, err := os.Stat(p); err == nil {
				configs = append(configs, m.configFile("apache", filepath.Base(p), p, "apache"))
			}
		}
	}

	// Virtual hosts dir
	vhostsDir := filepath.Join(m.rootPath, "vhosts")
	if entries, err := os.ReadDir(vhostsDir); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".conf") {
				p := filepath.Join(vhostsDir, e.Name())
				configs = append(configs, m.configFile("apache", e.Name(), p, "apache"))
			}
		}
	}

	return configs
}

func (m *Manager) nginxConfigs() []ConfigFile {
	// Quét bin/nginx/{version}/conf/nginx.conf
	configs := m.scanVersionedDirs("nginx", filepath.Join("conf", "nginx.conf"), "nginx")

	// Fallback system paths (non-Windows)
	if runtime.GOOS != "windows" {
		for _, p := range []string{
			"/etc/nginx/nginx.conf",
			"/usr/local/etc/nginx/nginx.conf",
		} {
			if _, err := os.Stat(p); err == nil {
				configs = append(configs, m.configFile("nginx", filepath.Base(p), p, "nginx"))
			}
		}
	}
	return configs
}

func (m *Manager) mysqlConfigs() []ConfigFile {
	// Quét bin/mysql/{version}/my.ini và my.cnf
	var configs []ConfigFile
	base := filepath.Join(m.rootPath, "bin", "mysql")
	if entries, err := os.ReadDir(base); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			for _, name := range []string{"my.ini", "my.cnf"} {
				p := filepath.Join(base, e.Name(), name)
				if _, err := os.Stat(p); err == nil {
					configs = append(configs, m.configFile("mysql", name+" ("+e.Name()+")", p, "ini"))
				}
			}
		}
	}

	// Fallback system paths (non-Windows)
	if runtime.GOOS != "windows" {
		for _, p := range []string{
			"/etc/mysql/mysql.conf.d/mysqld.cnf",
			"/etc/my.cnf",
			"/etc/mysql/my.cnf",
			"/usr/local/etc/my.cnf",
			"/opt/homebrew/etc/my.cnf",
		} {
			if _, err := os.Stat(p); err == nil {
				configs = append(configs, m.configFile("mysql", filepath.Base(p), p, "ini"))
			}
		}
	}
	return configs
}

func (m *Manager) phpConfigs() []ConfigFile {
	// Quét bin/php/{version}/php.ini.
	// Nếu php.ini chưa tồn tại nhưng php.ini-production có, tự tạo từ template.
	base := filepath.Join(m.rootPath, "bin", "php")
	if entries, err := os.ReadDir(base); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			ini := filepath.Join(base, e.Name(), "php.ini")
			if _, err := os.Stat(ini); err != nil {
				// php.ini chưa có — copy từ php.ini-production nếu tồn tại
				if src, err := os.ReadFile(filepath.Join(base, e.Name(), "php.ini-production")); err == nil {
					os.WriteFile(ini, src, 0644) //nolint:errcheck
				}
			}
		}
	}

	configs := m.scanVersionedDirs("php", "php.ini", "ini")

	// Fallback system paths (non-Windows)
	if runtime.GOOS != "windows" {
		for _, p := range []string{
			"/usr/local/etc/php/8.2/php.ini",
			"/opt/homebrew/etc/php/8.2/php.ini",
			"/etc/php/8.2/fpm/php.ini",
			"/etc/php/8.2/cli/php.ini",
		} {
			if _, err := os.Stat(p); err == nil {
				dir := filepath.Base(filepath.Dir(p))
				configs = append(configs, m.configFile("php", dir+"/php.ini", p, "ini"))
			}
		}
	}
	return configs
}
