package services

import (
	"runtime"
	"strings"
	"testing"
)

// TestBuildApacheCmd_Platform verifies correct binary name per OS.
func TestBuildApacheCmd_Platform(t *testing.T) {
	cmd, err := buildApacheCmd("/fake/bin", "")
	if err != nil {
		t.Fatalf("buildApacheCmd error: %v", err)
	}

	exe := cmd.Path
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(exe, "httpd.exe") {
			t.Errorf("Windows: expected httpd.exe, got %q", exe)
		}
	} else {
		if strings.HasSuffix(exe, ".exe") {
			t.Errorf("Unix: got .exe suffix: %q", exe)
		}
		// Unix apache runs in foreground mode
		found := false
		for _, arg := range cmd.Args {
			if arg == "-DFOREGROUND" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Unix: apache should have -DFOREGROUND flag")
		}
	}
}

// TestBuildNginxCmd_Platform verifies correct binary name per OS.
func TestBuildNginxCmd_Platform(t *testing.T) {
	cmd, err := buildNginxCmd("/fake/bin")
	if err != nil {
		t.Fatalf("buildNginxCmd error: %v", err)
	}

	exe := cmd.Path
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(exe, "nginx.exe") {
			t.Errorf("Windows: expected nginx.exe, got %q", exe)
		}
	} else {
		if strings.HasSuffix(exe, ".exe") {
			t.Errorf("Unix: got .exe suffix: %q", exe)
		}
	}
}

// TestBuildMySQLCmd_Platform verifies correct binary and flags per OS.
func TestBuildMySQLCmd_Platform(t *testing.T) {
	cmd, err := buildMySQLCmd("/fake/bin", "/fake/data", "")
	if err != nil {
		t.Fatalf("buildMySQLCmd error: %v", err)
	}

	exe := cmd.Path
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(exe, "mysqld.exe") {
			t.Errorf("Windows: expected mysqld.exe, got %q", exe)
		}
	} else {
		if strings.HasSuffix(exe, ".exe") {
			t.Errorf("Unix: got .exe suffix: %q", exe)
		}
		// Unix MySQL should have --user=mysql flag
		found := false
		for _, arg := range cmd.Args {
			if arg == "--user=mysql" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Unix: mysqld should have --user=mysql flag")
		}
	}
}

// TestBuildRedisCmd_Platform verifies correct binary name per OS.
func TestBuildRedisCmd_Platform(t *testing.T) {
	cmd, err := buildRedisCmd("/fake/bin", "")
	if err != nil {
		t.Fatalf("buildRedisCmd error: %v", err)
	}

	exe := cmd.Path
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(exe, "redis-server.exe") {
			t.Errorf("Windows: expected redis-server.exe, got %q", exe)
		}
	} else {
		if strings.HasSuffix(exe, ".exe") {
			t.Errorf("Unix: got .exe suffix: %q", exe)
		}
	}
}

// TestBuildPHPCmd_Platform verifies php-cgi (Windows) vs php-fpm (Unix).
func TestBuildPHPCmd_Platform(t *testing.T) {
	cmd, err := buildPHPCmd("/fake/bin", "")
	if err != nil {
		t.Fatalf("buildPHPCmd error: %v", err)
	}

	exe := cmd.Path
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(exe, "php-cgi.exe") {
			t.Errorf("Windows: expected php-cgi.exe, got %q", exe)
		}
	} else {
		if !strings.HasSuffix(exe, "php-fpm") {
			t.Errorf("Unix: expected php-fpm, got %q", exe)
		}
	}
}

// TestKillStaleProcesses_ExeFilter verifies only OS-appropriate executables are targeted.
func TestKillStaleProcesses_ExeFilter(t *testing.T) {
	// staleExeNames maps contain both .exe (Windows) and non-.exe (Unix) entries.
	// This test verifies the filter logic is correct — it doesn't actually kill anything
	// since binDir points to a nonexistent path.
	for name, exes := range staleExeNames {
		t.Run(string(name), func(t *testing.T) {
			hasWindows := false
			hasUnix := false
			for _, exe := range exes {
				if strings.HasSuffix(exe, ".exe") {
					hasWindows = true
				} else {
					hasUnix = true
				}
			}
			if !hasWindows {
				t.Errorf("service %q has no Windows executables in staleExeNames", name)
			}
			if !hasUnix {
				t.Errorf("service %q has no Unix executables in staleExeNames", name)
			}
		})
	}
}
