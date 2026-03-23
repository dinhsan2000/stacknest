package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ServiceConfig struct {
	Enabled     bool   `json:"enabled"`
	Port        int    `json:"port"`
	Path        string `json:"path"`
	Version     string `json:"version"`
	AutoRecover bool   `json:"auto_recover"`
}

type Config struct {
	RootPath string `json:"root_path"`
	BinPath  string `json:"bin_path"`  // thư mục gốc chứa tất cả service binaries
	DataPath string `json:"data_path"` // thư mục lưu dữ liệu service (MySQL data, v.v.)
	WWWPath  string `json:"www_path"`
	LogPath  string `json:"log_path"`

	Apache ServiceConfig `json:"apache"`
	Nginx  ServiceConfig `json:"nginx"`
	MySQL  ServiceConfig `json:"mysql"`
	PHP    ServiceConfig `json:"php"`
	Redis  ServiceConfig `json:"redis"`

	AutoStart bool   `json:"auto_start"`
	Theme     string `json:"theme"` // "light" | "dark"
}

// MySQLDataDir trả về thư mục data riêng cho từng phiên bản MySQL.
// Ví dụ: version="8.0.41" → {DataPath}/mysql/8.0.41/
// Nếu version rỗng, dùng "default" để tránh conflict.
func (c *Config) MySQLDataDir(version string) string {
	if version == "" {
		version = "default"
	}
	return filepath.Join(c.DataPath, "mysql", version)
}

func DefaultConfig() *Config {
	root := getRootPath()
	bin := filepath.Join(root, "bin")
	return &Config{
		RootPath: root,
		BinPath:  bin,
		DataPath: filepath.Join(root, "data"),
		WWWPath:  filepath.Join(root, "www"),
		LogPath:  filepath.Join(root, "logs"),
		Apache: ServiceConfig{
			Enabled:     true,
			Port:        80,
			Path:        filepath.Join(bin, "apache", "bin"),
			Version:     "2.4",
			AutoRecover: true,
		},
		Nginx: ServiceConfig{
			Enabled: false,
			Port:    8080,
			Path:    filepath.Join(bin, "nginx"),
			Version: "1.25",
		},
		MySQL: ServiceConfig{
			Enabled:     true,
			Port:        3306,
			Path:        filepath.Join(bin, "mysql", "bin"),
			Version:     "8.0",
			AutoRecover: true,
		},
		PHP: ServiceConfig{
			Enabled: true,
			Port:    9000,
			Path:    filepath.Join(bin, "php"),
			Version: "8.2",
		},
		Redis: ServiceConfig{
			Enabled: false,
			Port:    6379,
			Path:    filepath.Join(bin, "redis"),
			Version: "7.0",
		},
		AutoStart: false,
		Theme:     "dark",
	}
}

func getRootPath() string {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)

	// Handle go run trên Linux/Mac (exe nằm trong /tmp)
	if dir == "/tmp" || dir == os.TempDir() {
		dir, _ = os.Getwd()
		return dir
	}

	// Handle wails dev mode: exe được compile vào build/bin/
	// → dùng working directory (thư mục project gốc) thay thế
	normalized := filepath.ToSlash(dir)
	if strings.HasSuffix(normalized, "/build/bin") {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}

	return dir
}

func (c *Config) ConfigFilePath() string {
	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "Stacknest")
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Stacknest")
	default:
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "stacknest")
	}
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

// EnsureDirs tạo tất cả thư mục cần thiết của app nếu chưa tồn tại.
func (c *Config) EnsureDirs() {
	for _, d := range []string{
		c.BinPath,
		c.DataPath,
		c.WWWPath,
		c.LogPath,
		c.Apache.Path,
		c.Nginx.Path,
		c.MySQL.Path,
		c.PHP.Path,
		c.Redis.Path,
	} {
		if d != "" {
			os.MkdirAll(d, 0755)
		}
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()
	path := cfg.ConfigFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.EnsureDirs()
			return cfg, cfg.Save()
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), nil
	}

	// Migrate: root_path sai do wails dev mode (exe nằm trong build/bin/)
	if strings.HasSuffix(filepath.ToSlash(cfg.RootPath), "/build/bin") {
		root := getRootPath()
		cfg.RootPath = root
		cfg.BinPath = filepath.Join(root, "bin")
		cfg.DataPath = filepath.Join(root, "data")
		cfg.WWWPath = filepath.Join(root, "www")
		cfg.LogPath = filepath.Join(root, "logs")
	}

	// Migrate: config cũ chưa có BinPath
	if cfg.BinPath == "" {
		cfg.BinPath = filepath.Join(cfg.RootPath, "bin")
	}
	// Migrate: config cũ chưa có DataPath
	if cfg.DataPath == "" {
		cfg.DataPath = filepath.Join(cfg.RootPath, "data")
	}
	// Migrate: Apache và MySQL path cũ chưa có subdir "bin/"
	if cfg.Apache.Path == filepath.Join(cfg.BinPath, "apache") {
		cfg.Apache.Path = filepath.Join(cfg.BinPath, "apache", "bin")
	}
	if cfg.MySQL.Path == filepath.Join(cfg.BinPath, "mysql") {
		cfg.MySQL.Path = filepath.Join(cfg.BinPath, "mysql", "bin")
	}

	cfg.EnsureDirs()
	return cfg, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.ConfigFilePath(), data, 0644)
}
