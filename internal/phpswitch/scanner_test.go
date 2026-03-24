package phpswitch

import (
	"runtime"
	"testing"
)

func TestMajorVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"three parts", "8.2.10", "8.2"},
		{"three parts 7.4", "7.4.33", "7.4"},
		{"single part", "8", "8"},
		{"empty string", "", ""},
		{"two parts", "8.2", "8.2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := majorVersion(tc.input)
			if got != tc.want {
				t.Errorf("majorVersion(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestVersionRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{"full version", "PHP 8.2.10", true},
		{"version 7.4", "PHP 7.4.33", true},
		{"version 8.0.0", "PHP 8.0.0", true},
		{"no match text", "not php", false},
		{"empty string", "", false},
		{"partial match", "PHP 8.2", false},
		{"version without PHP prefix", "8.2.10", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := versionRegex.MatchString(tc.input)
			if got != tc.matches {
				t.Errorf("versionRegex.MatchString(%q) = %v, want %v", tc.input, got, tc.matches)
			}
		})
	}
}

func TestPhpExe(t *testing.T) {
	dir := "somedir"

	t.Run("windows uses php.exe suffix", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows-only")
		}
		got := phpExe(dir)
		if len(got) < 7 || got[len(got)-7:] != "php.exe" {
			t.Errorf("phpExe(%q) on windows = %q, want suffix php.exe", dir, got)
		}
	})

	t.Run("non-windows uses php suffix without .exe", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("non-windows only")
		}
		got := phpExe(dir)
		if len(got) < 3 || got[len(got)-3:] != "php" {
			t.Errorf("phpExe(%q) on unix = %q, want suffix php", dir, got)
		}
	})

	t.Run("result contains dir as prefix", func(t *testing.T) {
		got := phpExe(dir)
		if len(got) <= len(dir) {
			t.Errorf("phpExe(%q) = %q too short", dir, got)
		}
	})
}
