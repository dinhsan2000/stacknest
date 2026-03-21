package database

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed embed/adminer.php embed/index.php
var adminerFS embed.FS

// Server quản lý Adminer PHP built-in server
type Server struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	port     int
	phpExe   string
	admDir   string // thư mục chứa adminer.php đã extract
	rootPath string
}

func NewServer(rootPath string) *Server {
	s := &Server{rootPath: rootPath}
	s.phpExe = s.FindPHPExe()
	s.admDir = s.extractAdminer()
	return s
}

// extractAdminer giải nén adminer.php + index.php từ embed vào thư mục data
func (s *Server) extractAdminer() string {
	dir := filepath.Join(s.rootPath, "etc", "adminer")

	// Nếu cả 2 file đã tồn tại, không cần extract lại
	indexPath := filepath.Join(dir, "index.php")
	adminerPath := filepath.Join(dir, "adminer.php")
	if _, err1 := os.Stat(indexPath); err1 == nil {
		if _, err2 := os.Stat(adminerPath); err2 == nil {
			return dir
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}

	// Extract cả 2 file
	for _, name := range []string{"embed/adminer.php", "embed/index.php"} {
		data, err := adminerFS.ReadFile(name)
		if err != nil {
			return ""
		}
		out := filepath.Join(dir, filepath.Base(name))
		if err := os.WriteFile(out, data, 0644); err != nil {
			return ""
		}
	}
	return dir
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

// Start khởi động PHP built-in server để chạy Adminer
// Trả về URL để mở trong browser
func (s *Server) Start() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		return s.URL(), nil // đã chạy rồi
	}

	// Re-detect PHP nếu chưa tìm thấy
	if s.phpExe == "" {
		s.phpExe = s.FindPHPExe()
	}
	if s.phpExe == "" {
		return "", fmt.Errorf("PHP not found — install PHP or configure it in PHP Versions")
	}

	// Đảm bảo adminer files đã extract (force nếu thiếu index.php)
	s.admDir = s.extractAdminer()
	if s.admDir == "" {
		return "", fmt.Errorf("failed to extract bundled Adminer")
	}

	port, err := freePort(8484)
	if err != nil {
		return "", fmt.Errorf("no free port available: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	cmd := exec.Command(s.phpExe, "-S", addr, "index.php")
	cmd.Dir = s.admDir

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

// AdminerFound kiểm tra có bundled Adminer không
func (s *Server) AdminerFound() bool {
	return s.admDir != ""
}

// PHPFound kiểm tra có tìm thấy PHP không
func (s *Server) PHPFound() bool {
	return s.phpExe != ""
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
