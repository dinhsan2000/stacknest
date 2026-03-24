// Package testutil provides shared test helpers for Stacknest Go tests.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SetupRootDir creates a temp directory mimicking Stacknest's root structure.
// Returns the root path. Cleanup is automatic via t.TempDir().
func SetupRootDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{
		"bin", "data", "www", "logs",
		"ssl", "vhosts", "etc",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatalf("SetupRootDir: mkdir %s: %v", dir, err)
		}
	}
	return root
}

// WriteFile creates a file at path with the given content, creating parent dirs as needed.
func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("WriteFile: mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: write %s: %v", path, err)
	}
}

// AssertFileExists fails the test if the file does not exist.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists fails the test if the file exists.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to NOT exist: %s", path)
	}
}

// AssertFileContains fails if the file does not contain the substring.
func AssertFileContains(t *testing.T, path, substring string) {
	t.Helper()
	content := ReadFileString(t, path)
	if !strings.Contains(content, substring) {
		t.Errorf("file %s does not contain %q", path, substring)
	}
}

// ReadFileString reads a file and returns its content as string.
func ReadFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFileString: %v", err)
	}
	return string(data)
}
