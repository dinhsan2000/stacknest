package database

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Server quản lý Adminer PHP built-in server
type Server struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	port    int
	phpExe  string
	admPath string
	rootPath string
}

func NewServer(rootPath string) *Server {
	s := &Server{rootPath: rootPath}
	s.phpExe = s.FindPHPExe()
	s.admPath = s.FindAdminerPath()
	return s
}

// FindAdminerPath tìm file adminer/phpMyAdmin trên máy
func (s *Server) FindAdminerPath() string {
	candidates := []string{
		`C:\laragon\etc\apps\adminer\index.php`,
		`D:\laragon\etc\apps\adminer\index.php`,
		`C:\xampp\phpMyAdmin\index.php`,
		`C:\laragon\www\phpmyadmin\index.php`,
		`C:\laragon\www\adminer\index.php`,
		filepath.Join(s.rootPath, "etc", "apps", "adminer", "index.php"),
		filepath.Join(s.rootPath, "www", "phpmyadmin", "index.php"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// FindPHPExe tìm PHP executable — ưu tiên từ phpswitch state file
func (s *Server) FindPHPExe() string {
	// Thử đọc từ phpswitch state file
	statePath := filepath.Join(s.rootPath, "php_versions.json")
	if data, err := os.ReadFile(statePath); err == nil {
		var state struct {
			ActivePath string `json:"active_path"`
		}
		if json.Unmarshal(data, &state) == nil && state.ActivePath != "" {
			if _, err := os.Stat(state.ActivePath); err == nil {
				return state.ActivePath
			}
		}
	}

	// Fallback: tìm trong PATH
	if p, err := exec.LookPath("php"); err == nil {
		return p
	}
	if p, err := exec.LookPath("php.exe"); err == nil {
		return p
	}

	// Laragon PHP dirs fallback
	laragonPhpBase := `C:\laragon\bin\php`
	if entries, err := os.ReadDir(laragonPhpBase); err == nil {
		for i := len(entries) - 1; i >= 0; i-- {
			e := entries[i]
			if e.IsDir() {
				exe := filepath.Join(laragonPhpBase, e.Name(), "php.exe")
				if _, err := os.Stat(exe); err == nil {
					return exe
				}
			}
		}
	}
	return ""
}

// FindHeidiSQL tìm HeidiSQL executable
func FindHeidiSQL() string {
	candidates := []string{
		`C:\laragon\bin\heidisql\heidisql.exe`,
		`D:\laragon\bin\heidisql\heidisql.exe`,
		`C:\Program Files\HeidiSQL\heidisql.exe`,
		`C:\Program Files (x86)\HeidiSQL\heidisql.exe`,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("heidisql.exe"); err == nil {
			return p
		}
	}
	return ""
}

// Start khởi động PHP built-in server để chạy Adminer
// Trả về URL để mở trong browser
func (s *Server) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		return s.URL(), nil // đã chạy rồi
	}

	if s.phpExe == "" {
		return "", fmt.Errorf("PHP not found — install PHP or configure it in PHP Versions")
	}
	if s.admPath == "" {
		return "", fmt.Errorf("Adminer not found — expected at C:\\laragon\\etc\\apps\\adminer\\index.php")
	}

	port, err := freePort(8484)
	if err != nil {
		return "", fmt.Errorf("no free port available: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	admDir := filepath.Dir(s.admPath)
	admFile := filepath.Base(s.admPath)

	cmd := exec.Command(s.phpExe, "-S", addr, admFile)
	cmd.Dir = admDir

	// Không gắn stdout/stderr để tránh blocking
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start PHP server: %w", err)
	}

	s.cmd = cmd
	s.port = port

	return s.URL(), nil
}

// Stop dừng PHP built-in server
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
		s.cmd = nil
		s.port = 0
	}
}

// IsRunning kiểm tra server có đang chạy không
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cmd != nil && s.cmd.Process != nil
}

// URL trả về địa chỉ Adminer (không cần lock vì chỉ đọc port)
func (s *Server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

// AdminerFound kiểm tra có tìm thấy Adminer không
func (s *Server) AdminerFound() bool {
	return s.admPath != ""
}

// PHPFound kiểm tra có tìm thấy PHP không
func (s *Server) PHPFound() bool {
	return s.phpExe != ""
}

// AdminerPath trả về đường dẫn adminer (dùng để hiện thị)
func (s *Server) AdminerPath() string {
	return s.admPath
}

// PHPPath trả về đường dẫn PHP
func (s *Server) PHPPath() string {
	return s.phpExe
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// freePort tìm port còn trống bắt đầu từ startPort
func freePort(startPort int) (int, error) {
	for port := startPort; port < startPort+20; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
		if !strings.Contains(err.Error(), "address already in use") &&
			!strings.Contains(err.Error(), "bind") {
			return 0, err
		}
	}
	return 0, fmt.Errorf("no free port in range %d-%d", startPort, startPort+20)
}
