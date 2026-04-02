package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMySQLDataDir(t *testing.T) {
	c := &Config{DataPath: "/data"}

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"empty version uses default", "", filepath.Join("/data", "mysql", "default")},
		{"explicit version", "8.0.41", filepath.Join("/data", "mysql", "8.0.41")},
		{"minor version", "5.7", filepath.Join("/data", "mysql", "5.7")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := c.MySQLDataDir(tc.version)
			if got != tc.want {
				t.Errorf("MySQLDataDir(%q) = %q, want %q", tc.version, got, tc.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("root path not empty", func(t *testing.T) {
		if cfg.RootPath == "" {
			t.Error("RootPath must not be empty")
		}
	})

	t.Run("default ports", func(t *testing.T) {
		cases := []struct {
			label string
			got   int
			want  int
		}{
			{"Apache", cfg.Apache.Port, 80},
			{"Nginx", cfg.Nginx.Port, 8080},
			{"MySQL", cfg.MySQL.Port, 3306},
			{"PHP", cfg.PHP.Port, 9000},
			{"Redis", cfg.Redis.Port, 6379},
		}
		for _, c := range cases {
			if c.got != c.want {
				t.Errorf("%s port = %d, want %d", c.label, c.got, c.want)
			}
		}
	})

	t.Run("theme dark", func(t *testing.T) {
		if cfg.Theme != "dark" {
			t.Errorf("Theme = %q, want %q", cfg.Theme, "dark")
		}
	})

	t.Run("auto start false", func(t *testing.T) {
		if cfg.AutoStart {
			t.Error("AutoStart must be false by default")
		}
	})

	t.Run("apache enabled", func(t *testing.T) {
		if !cfg.Apache.Enabled {
			t.Error("Apache.Enabled must be true")
		}
	})

	t.Run("nginx disabled", func(t *testing.T) {
		if cfg.Nginx.Enabled {
			t.Error("Nginx.Enabled must be false")
		}
	})

	t.Run("mysql enabled", func(t *testing.T) {
		if !cfg.MySQL.Enabled {
			t.Error("MySQL.Enabled must be true")
		}
	})
}

func TestConfigJSONRoundTrip(t *testing.T) {
	original := &Config{
		RootPath: "/some/root",
		BinPath:  "/some/root/bin",
		DataPath: "/some/root/data",
		WWWPath:  "/some/root/www",
		LogPath:  "/some/root/logs",
		Theme:    "dark",
		Apache:   ServiceConfig{Enabled: true, Port: 80, Version: "2.4"},
		Nginx:    ServiceConfig{Enabled: false, Port: 8080, Version: "1.25"},
		MySQL:    ServiceConfig{Enabled: true, Port: 3306, Version: "8.0"},
		PHP:      ServiceConfig{Enabled: true, Port: 9000, Version: "8.2"},
		Redis:    ServiceConfig{Enabled: false, Port: 6379, Version: "7.0"},
		AutoStart: false,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored Config
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	checks := []struct {
		label string
		got   interface{}
		want  interface{}
	}{
		{"RootPath", restored.RootPath, original.RootPath},
		{"BinPath", restored.BinPath, original.BinPath},
		{"DataPath", restored.DataPath, original.DataPath},
		{"Theme", restored.Theme, original.Theme},
		{"AutoStart", restored.AutoStart, original.AutoStart},
		{"Apache.Port", restored.Apache.Port, original.Apache.Port},
		{"Apache.Enabled", restored.Apache.Enabled, original.Apache.Enabled},
		{"MySQL.Port", restored.MySQL.Port, original.MySQL.Port},
		{"Redis.Port", restored.Redis.Port, original.Redis.Port},
		{"Nginx.Enabled", restored.Nginx.Enabled, original.Nginx.Enabled},
	}

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("got %v, want %v", c.got, c.want)
			}
		})
	}
}

// TestEnsureDirs verifies all required directories are created.
func TestEnsureDirs(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	cfg := &Config{
		RootPath: root,
		BinPath:  bin,
		DataPath: filepath.Join(root, "data"),
		WWWPath:  filepath.Join(root, "www"),
		LogPath:  filepath.Join(root, "logs"),
		Apache:   ServiceConfig{Path: filepath.Join(bin, "apache", "bin")},
		Nginx:    ServiceConfig{Path: filepath.Join(bin, "nginx")},
		MySQL:    ServiceConfig{Path: filepath.Join(bin, "mysql", "bin")},
		Postgres: ServiceConfig{Path: filepath.Join(bin, "postgres", "bin")},
		MongoDB:  ServiceConfig{Path: filepath.Join(bin, "mongodb", "bin")},
		PHP:      ServiceConfig{Path: filepath.Join(bin, "php")},
		Redis:    ServiceConfig{Path: filepath.Join(bin, "redis")},
	}

	cfg.EnsureDirs()

	expected := []string{
		"bin",
		"data",
		"www",
		"logs",
		filepath.Join("logs", "apache"),
		filepath.Join("logs", "nginx"),
		filepath.Join("logs", "mysql"),
		filepath.Join("logs", "postgres"),
		filepath.Join("logs", "mongodb"),
		filepath.Join("logs", "php"),
		filepath.Join("logs", "redis"),
		"ssl",
		"vhosts",
		filepath.Join("vhosts", "nginx"),
		".config_backups",
		"etc",
		filepath.Join("bin", "apache", "bin"),
		filepath.Join("bin", "nginx"),
		filepath.Join("bin", "mysql", "bin"),
		filepath.Join("bin", "postgres", "bin"),
		filepath.Join("bin", "mongodb", "bin"),
		filepath.Join("bin", "php"),
		filepath.Join("bin", "redis"),
	}

	for _, dir := range expected {
		t.Run(dir, func(t *testing.T) {
			full := filepath.Join(root, dir)
			info, err := os.Stat(full)
			if err != nil {
				t.Errorf("directory %q not created: %v", dir, err)
				return
			}
			if !info.IsDir() {
				t.Errorf("%q exists but is not a directory", dir)
			}
		})
	}
}
