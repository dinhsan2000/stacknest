package config

import (
	"runtime"
	"strings"
	"testing"
)

// TestConfigFilePath_Platform verifies ConfigFilePath uses correct OS-specific directory.
func TestConfigFilePath_Platform(t *testing.T) {
	cfg := DefaultConfig()
	path := cfg.ConfigFilePath()

	if path == "" {
		t.Fatal("ConfigFilePath() returned empty string")
	}

	if !strings.HasSuffix(path, "config.json") {
		t.Errorf("ConfigFilePath() = %q, want suffix config.json", path)
	}

	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(path, "Stacknest") {
			t.Errorf("Windows: ConfigFilePath() = %q, want path containing 'Stacknest'", path)
		}
	case "darwin":
		if !strings.Contains(path, "Library/Application Support/Stacknest") {
			t.Errorf("macOS: ConfigFilePath() = %q, want Library/Application Support/Stacknest", path)
		}
	default:
		if !strings.Contains(path, ".config/stacknest") {
			t.Errorf("Linux: ConfigFilePath() = %q, want .config/stacknest", path)
		}
	}
}

// TestDefaultConfig_PathSeparators verifies paths use correct separators for the current OS.
func TestDefaultConfig_PathSeparators(t *testing.T) {
	cfg := DefaultConfig()

	paths := []struct {
		name string
		path string
	}{
		{"BinPath", cfg.BinPath},
		{"DataPath", cfg.DataPath},
		{"WWWPath", cfg.WWWPath},
		{"LogPath", cfg.LogPath},
		{"Apache.Path", cfg.Apache.Path},
		{"MySQL.Path", cfg.MySQL.Path},
	}

	for _, p := range paths {
		t.Run(p.name, func(t *testing.T) {
			if p.path == "" {
				t.Skip("path is empty")
			}
			// On Windows, filepath.Join produces backslash; on Unix, forward slash.
			// Verify no mixed separators exist.
			if runtime.GOOS == "windows" {
				if strings.Contains(p.path, "/") {
					t.Errorf("Windows path %q contains forward slash", p.path)
				}
			} else {
				if strings.Contains(p.path, `\`) {
					t.Errorf("Unix path %q contains backslash", p.path)
				}
			}
		})
	}
}
