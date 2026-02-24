package phpswitch

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"os"
)

// PHPInstall thông tin một bản cài PHP
type PHPInstall struct {
	Version string `json:"version"` // "8.2.10"
	Major   string `json:"major"`   // "8.2"
	Path    string `json:"path"`    // đường dẫn tới php executable
	Active  bool   `json:"active"`
}

var versionRegex = regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)

// Scan tìm tất cả PHP installations trên máy
func Scan(extraDirs []string) []PHPInstall {
	candidates := phpSearchPaths()
	candidates = append(candidates, extraDirs...)

	seen := map[string]bool{}
	var installs []PHPInstall

	for _, dir := range candidates {
		exe := phpExe(dir)
		abs, err := filepath.Abs(exe)
		if err != nil {
			continue
		}
		if seen[abs] {
			continue
		}
		if _, err := os.Stat(abs); err != nil {
			continue
		}

		ver := getVersion(abs)
		if ver == "" {
			continue
		}

		seen[abs] = true
		installs = append(installs, PHPInstall{
			Version: ver,
			Major:   majorVersion(ver),
			Path:    abs,
		})
	}

	return installs
}

// GetVersion trả về version string của một PHP executable
func GetVersion(path string) string {
	return getVersion(path)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func getVersion(exe string) string {
	out, err := exec.Command(exe, "-r", "echo PHP_VERSION;").Output()
	if err != nil {
		return ""
	}
	ver := strings.TrimSpace(string(out))
	if versionRegex.MatchString(ver) || regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(ver) {
		return ver
	}
	return ""
}

func majorVersion(ver string) string {
	parts := strings.Split(ver, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return ver
}

func phpExe(dir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(dir, "php.exe")
	}
	return filepath.Join(dir, "php")
}

// phpSearchPaths trả về danh sách thư mục cần scan theo OS
func phpSearchPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return windowsPHPPaths()
	case "darwin":
		return macPHPPaths()
	default:
		return linuxPHPPaths()
	}
}

func windowsPHPPaths() []string {
	var paths []string

	// Laragon style: C:\laragon\bin\php\php8.2.*
	laragonRoots := []string{
		`C:\laragon\bin\php`,
		`D:\laragon\bin\php`,
	}
	for _, root := range laragonRoots {
		if entries, err := os.ReadDir(root); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					paths = append(paths, filepath.Join(root, e.Name()))
				}
			}
		}
	}

	// XAMPP, WampServer, direct installs
	paths = append(paths,
		`C:\php`,
		`C:\php8`,
		`C:\xampp\php`,
		`C:\wamp64\bin\php\php8.2.0`,
	)

	// PATH-based discovery
	if p, err := exec.LookPath("php"); err == nil {
		paths = append(paths, filepath.Dir(p))
	}
	if p, err := exec.LookPath("php.exe"); err == nil {
		paths = append(paths, filepath.Dir(p))
	}

	return paths
}

func macPHPPaths() []string {
	var paths []string

	// Homebrew multi-version: /opt/homebrew/opt/php@8.2/bin
	brew := []string{
		"/opt/homebrew/opt",
		"/usr/local/opt",
	}
	for _, base := range brew {
		if entries, err := os.ReadDir(base); err == nil {
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "php@") || e.Name() == "php" {
					paths = append(paths, filepath.Join(base, e.Name(), "bin"))
				}
			}
		}
	}

	paths = append(paths, "/usr/bin", "/usr/local/bin")

	if p, err := exec.LookPath("php"); err == nil {
		paths = append(paths, filepath.Dir(p))
	}
	return paths
}

func linuxPHPPaths() []string {
	var paths []string

	// phpenv, ondrej PPA style: /usr/bin/php8.2
	for _, ver := range []string{"8.3", "8.2", "8.1", "8.0", "7.4"} {
		paths = append(paths,
			"/usr/bin/php"+ver,
			"/usr/lib/php/"+ver,
		)
	}

	paths = append(paths, "/usr/bin", "/usr/local/bin")

	if p, err := exec.LookPath("php"); err == nil {
		paths = append(paths, filepath.Dir(p))
	}
	return paths
}
