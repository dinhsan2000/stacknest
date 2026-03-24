package vhost

import (
	"runtime"
	"testing"
)

// TestHostsFilePath_Platform verifies correct hosts file location per OS.
func TestHostsFilePath_Platform(t *testing.T) {
	m := &Manager{}
	path := m.hostsFilePath()

	switch runtime.GOOS {
	case "windows":
		want := `C:\Windows\System32\drivers\etc\hosts`
		if path != want {
			t.Errorf("Windows: hostsFilePath() = %q, want %q", path, want)
		}
	default:
		if path != "/etc/hosts" {
			t.Errorf("Unix: hostsFilePath() = %q, want /etc/hosts", path)
		}
	}
}
