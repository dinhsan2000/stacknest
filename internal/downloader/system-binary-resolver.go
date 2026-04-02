package downloader

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// systemExeNames maps service name → executable names to search on the system.
// Used as fallback when no downloaded binary exists (macOS/Linux).
var systemExeNames = map[string][]string{
	"apache":   {"httpd", "apache2"},
	"nginx":    {"nginx"},
	"mysql":    {"mysqld"},
	"postgres": {"postgres"},
	"mongodb":  {"mongod"},
	"php":      {"php-fpm", "php-cgi", "php"},
	"redis":    {"redis-server"},
}

// systemSearchPaths returns OS-specific directories to scan for service binaries.
func systemSearchPaths(service string) []string {
	switch runtime.GOOS {
	case "darwin":
		return darwinSearchPaths(service)
	case "linux":
		return linuxSearchPaths(service)
	default:
		return nil // Windows uses downloaded binaries only
	}
}

func darwinSearchPaths(service string) []string {
	// Homebrew (Apple Silicon + Intel) and system paths
	paths := []string{
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}

	// Service-specific Homebrew opt paths
	switch service {
	case "apache":
		paths = append(paths,
			"/opt/homebrew/opt/httpd/bin",
			"/usr/local/opt/httpd/bin",
		)
	case "mysql":
		paths = append(paths,
			"/opt/homebrew/opt/mysql/bin",
			"/usr/local/opt/mysql/bin",
			"/opt/homebrew/opt/mysql@8.0/bin",
			"/usr/local/opt/mysql@8.0/bin",
		)
	case "postgres":
		paths = append(paths,
			"/opt/homebrew/opt/postgresql@17/bin",
			"/opt/homebrew/opt/postgresql@16/bin",
			"/opt/homebrew/opt/postgresql/bin",
			"/usr/local/opt/postgresql@17/bin",
			"/usr/local/opt/postgresql@16/bin",
			"/usr/local/opt/postgresql/bin",
		)
	case "mongodb":
		paths = append(paths,
			"/opt/homebrew/opt/mongodb-community/bin",
			"/usr/local/opt/mongodb-community/bin",
		)
	case "nginx":
		paths = append(paths,
			"/opt/homebrew/opt/nginx/bin",
			"/usr/local/opt/nginx/bin",
		)
	case "php":
		paths = append(paths,
			"/opt/homebrew/opt/php/sbin",
			"/opt/homebrew/opt/php@8.3/sbin",
			"/opt/homebrew/opt/php@8.2/sbin",
			"/usr/local/opt/php/sbin",
			"/usr/local/opt/php@8.3/sbin",
			"/usr/local/opt/php@8.2/sbin",
		)
	case "redis":
		paths = append(paths,
			"/opt/homebrew/opt/redis/bin",
			"/usr/local/opt/redis/bin",
		)
	}

	return paths
}

func linuxSearchPaths(service string) []string {
	paths := []string{
		"/usr/bin",
		"/usr/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}

	switch service {
	case "mysql":
		paths = append(paths,
			"/usr/local/mysql/bin",
			"/usr/lib/mysql/bin",
		)
	case "postgres":
		paths = append(paths,
			"/usr/lib/postgresql/17/bin",
			"/usr/lib/postgresql/16/bin",
			"/usr/lib/postgresql/15/bin",
		)
	case "mongodb":
		paths = append(paths,
			"/usr/lib/mongodb/bin",
		)
	}

	return paths
}

// FindSystemBinary searches for a service binary installed on the system.
// Returns the directory containing the executable, or empty string if not found.
// On Windows, always returns empty (Windows uses downloaded binaries).
func FindSystemBinary(service string) string {
	if runtime.GOOS == "windows" {
		return ""
	}

	exeNames, ok := systemExeNames[service]
	if !ok {
		return ""
	}

	// First: check OS-specific search paths
	for _, dir := range systemSearchPaths(service) {
		for _, exe := range exeNames {
			fullPath := filepath.Join(dir, exe)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return dir
			}
		}
	}

	// Second: fallback to PATH lookup
	for _, exe := range exeNames {
		if p, err := exec.LookPath(exe); err == nil {
			return filepath.Dir(p)
		}
	}

	return ""
}

// DetectSystemVersion attempts to detect the version of a system-installed binary.
// Returns a version string or "system" as fallback.
func DetectSystemVersion(service, binDir string) string {
	type versionCmd struct {
		exe  string
		args []string
	}

	cmds := map[string]versionCmd{
		"apache":   {exe: "httpd", args: []string{"-v"}},
		"nginx":    {exe: "nginx", args: []string{"-v"}},
		"mysql":    {exe: "mysqld", args: []string{"--version"}},
		"postgres": {exe: "postgres", args: []string{"--version"}},
		"mongodb":  {exe: "mongod", args: []string{"--version"}},
		"php":      {exe: "php", args: []string{"-v"}},
		"redis":    {exe: "redis-server", args: []string{"--version"}},
	}

	vc, ok := cmds[service]
	if !ok {
		return "system"
	}

	exePath := filepath.Join(binDir, vc.exe)
	if _, err := os.Stat(exePath); err != nil {
		// Try alternative exe names
		if alts, ok := systemExeNames[service]; ok {
			for _, alt := range alts {
				altPath := filepath.Join(binDir, alt)
				if _, err := os.Stat(altPath); err == nil {
					exePath = altPath
					break
				}
			}
		}
	}

	out, err := exec.Command(exePath, vc.args...).CombinedOutput()
	if err != nil {
		return "system"
	}

	return parseVersionFromOutput(string(out))
}

// parseVersionFromOutput extracts a version number (X.Y.Z pattern) from command output.
func parseVersionFromOutput(output string) string {
	// Look for common version patterns: X.Y.Z, X.Y
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Find first digit sequence that looks like a version
		for i, ch := range line {
			if ch >= '0' && ch <= '9' {
				// Extract version starting here
				end := i
				dots := 0
				for end < len(line) {
					c := line[end]
					if c >= '0' && c <= '9' {
						end++
					} else if c == '.' && dots < 2 {
						dots++
						end++
					} else {
						break
					}
				}
				ver := line[i:end]
				// Must have at least X.Y format
				if strings.Contains(ver, ".") && len(ver) >= 3 {
					return strings.TrimSuffix(ver, ".")
				}
			}
		}
	}
	return "system"
}
