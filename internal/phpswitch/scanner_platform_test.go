package phpswitch

import (
	"runtime"
	"testing"
)

// TestPhpSearchPaths_Platform verifies OS-specific PHP search paths are returned.
func TestPhpSearchPaths_Platform(t *testing.T) {
	paths := phpSearchPaths()

	if len(paths) == 0 {
		t.Fatal("phpSearchPaths() returned no paths")
	}

	switch runtime.GOOS {
	case "windows":
		// Should include Laragon-style paths
		found := false
		for _, p := range paths {
			if p == `C:\php` || p == `C:\xampp\php` {
				found = true
				break
			}
		}
		if !found {
			t.Error("Windows: expected common Windows PHP paths (C:\\php or C:\\xampp\\php)")
		}
	case "darwin":
		// Should include Homebrew paths
		found := false
		for _, p := range paths {
			if p == "/usr/local/bin" || p == "/usr/bin" {
				found = true
				break
			}
		}
		if !found {
			t.Error("macOS: expected /usr/local/bin or /usr/bin in search paths")
		}
	default:
		// Linux should include common version paths
		found := false
		for _, p := range paths {
			if p == "/usr/bin" || p == "/usr/local/bin" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Linux: expected /usr/bin or /usr/local/bin in search paths")
		}
	}
}

// TestPhpExe_Platform verifies correct executable suffix for current OS.
func TestPhpExe_Platform(t *testing.T) {
	exe := phpExe("/test/dir")

	switch runtime.GOOS {
	case "windows":
		if exe != `\test\dir\php.exe` && exe != `/test/dir/php.exe` {
			// filepath.Join normalizes, just check suffix
			if len(exe) < 7 || exe[len(exe)-7:] != "php.exe" {
				t.Errorf("Windows: phpExe() = %q, want suffix php.exe", exe)
			}
		}
	default:
		if len(exe) < 3 || exe[len(exe)-3:] != "php" {
			t.Errorf("Unix: phpExe() = %q, want suffix php (not php.exe)", exe)
		}
	}
}
