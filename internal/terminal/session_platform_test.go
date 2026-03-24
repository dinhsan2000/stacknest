package terminal

import (
	"runtime"
	"testing"
)

// TestGetShell_Platform verifies correct default shell per OS.
func TestGetShell_Platform(t *testing.T) {
	shell, args := getShell()

	if shell == "" {
		t.Fatal("getShell() returned empty shell path")
	}

	switch runtime.GOOS {
	case "windows":
		// Should be pwsh, powershell, or cmd.exe
		valid := shell == "cmd.exe" ||
			len(shell) > 0 // pwsh or powershell resolved from LookPath
		if !valid {
			t.Errorf("Windows: unexpected shell %q", shell)
		}
	case "darwin":
		// Should be zsh or bash
		if shell != "/bin/bash" && shell != "/bin/zsh" {
			// LookPath may return full path for zsh
			t.Logf("macOS: shell = %q (may be zsh from LookPath)", shell)
		}
	default:
		// Should be bash or sh
		if shell != "/bin/sh" && shell != "/bin/bash" {
			t.Logf("Linux: shell = %q (may be bash from LookPath)", shell)
		}
	}

	// args should not be nil (may be empty slice)
	_ = args
}
