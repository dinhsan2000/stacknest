package logs

import (
	"runtime"
	"testing"
)

// TestDefaultLogRoot_Platform verifies OS-specific log root paths.
func TestDefaultLogRoot_Platform(t *testing.T) {
	root := defaultLogRoot()

	if root == "" {
		t.Fatal("defaultLogRoot() returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		if root != `C:\laragon\logs` {
			t.Errorf("Windows: defaultLogRoot() = %q, want C:\\laragon\\logs", root)
		}
	case "darwin":
		if root != "/usr/local/var/log" {
			t.Errorf("macOS: defaultLogRoot() = %q, want /usr/local/var/log", root)
		}
	default:
		if root != "/var/log" {
			t.Errorf("Linux: defaultLogRoot() = %q, want /var/log", root)
		}
	}
}

// TestLogPaths_AllServicesPresent verifies all expected services have log entries.
func TestLogPaths_AllServicesPresent(t *testing.T) {
	paths := LogPaths("")

	expected := []string{"apache", "nginx", "mysql", "php"}
	for _, svc := range expected {
		t.Run(svc, func(t *testing.T) {
			logs, ok := paths[svc]
			if !ok {
				t.Errorf("service %q missing from LogPaths", svc)
				return
			}
			if len(logs) == 0 {
				t.Errorf("service %q has empty log path list", svc)
			}
		})
	}
}
