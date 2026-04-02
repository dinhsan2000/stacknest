package downloader

import (
	"runtime"
	"testing"
)

func TestParseVersionFromOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
	}{
		{"apache", "Server version: Apache/2.4.63 (Unix)", "2.4.63"},
		{"nginx", "nginx version: nginx/1.26.3", "1.26.3"},
		{"mysql", "mysqld  Ver 8.0.41 for Linux on x86_64", "8.0.41"},
		{"postgres", "postgres (PostgreSQL) 17.4", "17.4"},
		{"mongod", "db version v8.0.4", "8.0.4"},
		{"redis", "Redis server v=7.0.15 sha=00000000:0 malloc=jemalloc", "7.0.15"},
		{"php", "PHP 8.3.16 (cli) (built: Feb 2026)", "8.3.16"},
		{"empty", "", "system"},
		{"no version", "some random output", "system"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseVersionFromOutput(tc.input)
			if got != tc.want {
				t.Errorf("parseVersionFromOutput(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSystemExeNames_AllServicesPresent(t *testing.T) {
	expected := []string{"apache", "nginx", "mysql", "postgres", "mongodb", "php", "redis"}
	for _, svc := range expected {
		t.Run(svc, func(t *testing.T) {
			exes, ok := systemExeNames[svc]
			if !ok || len(exes) == 0 {
				t.Errorf("systemExeNames missing or empty for %q", svc)
			}
		})
	}
}

func TestFindSystemBinary_WindowsReturnsEmpty(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only")
	}
	// On Windows, FindSystemBinary should always return empty
	for svc := range systemExeNames {
		got := FindSystemBinary(svc)
		if got != "" {
			t.Errorf("FindSystemBinary(%q) on Windows = %q, want empty", svc, got)
		}
	}
}

func TestSystemSearchPaths_WindowsReturnsNil(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only")
	}
	for svc := range systemExeNames {
		paths := systemSearchPaths(svc)
		if paths != nil {
			t.Errorf("systemSearchPaths(%q) on Windows should return nil", svc)
		}
	}
}
