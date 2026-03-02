package phpswitch

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// State lưu trạng thái PHP versions đã biết
type State struct {
	ActivePath string       `json:"active_path"`
	Installs   []PHPInstall `json:"installs"`
	ExtraDirs  []string     `json:"extra_dirs"`
}

// Switcher quản lý PHP version switching
type Switcher struct {
	statePath string
	state     State
}

func NewSwitcher(configDir string) *Switcher {
	s := &Switcher{
		statePath: filepath.Join(configDir, "php_versions.json"),
	}
	s.load()
	return s
}

// GetInstalls trả về danh sách PHP đã biết (cộng scan mới)
func (s *Switcher) GetInstalls() []PHPInstall {
	// Re-scan với extra dirs
	installs := Scan(s.state.ExtraDirs)

	// Đánh dấu active
	for i := range installs {
		installs[i].Active = installs[i].Path == s.state.ActivePath
	}

	// Nếu chưa có active, dùng cái đầu tiên
	if s.state.ActivePath == "" && len(installs) > 0 {
		installs[0].Active = true
		s.state.ActivePath = installs[0].Path
		_ = s.save()
	}

	s.state.Installs = installs
	return installs
}

// GetActive trả về PHP install đang active
func (s *Switcher) GetActive() *PHPInstall {
	installs := s.GetInstalls()
	for i := range installs {
		if installs[i].Active {
			return &installs[i]
		}
	}
	if len(installs) > 0 {
		return &installs[0]
	}
	return nil
}

// Switch chuyển sang PHP version theo path
func (s *Switcher) Switch(phpPath string) error {
	// Xác nhận file tồn tại và là PHP
	ver := GetVersion(phpPath)
	if ver == "" {
		return fmt.Errorf("'%s' is not a valid PHP executable", phpPath)
	}

	s.state.ActivePath = phpPath
	if err := s.save(); err != nil {
		return err
	}

	// Trên Windows: cập nhật symlink/junction "current" → PHP dir.
	// Không fatal nếu thất bại — switch vẫn thành công qua ActivePath.
	if runtime.GOOS == "windows" {
		if err := s.updateWindowsSymlink(phpPath); err != nil {
			// Junction/symlink là tiện ích PATH, không ảnh hưởng hoạt động chính.
			_ = err
		}
	}

	return nil
}

// AddCustomPath thêm thư mục PHP tùy chỉnh vào danh sách scan
func (s *Switcher) AddCustomPath(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("directory not found: %s", dir)
	}
	for _, d := range s.state.ExtraDirs {
		if d == dir {
			return nil // already added
		}
	}
	s.state.ExtraDirs = append(s.state.ExtraDirs, dir)
	return s.save()
}

// ActivePHPPath trả về đường dẫn PHP executable đang active
func (s *Switcher) ActivePHPPath() string {
	return s.state.ActivePath
}

// ─── Windows symlink helper ───────────────────────────────────────────────────

// updateWindowsSymlink cập nhật symlink/junction "current" → PHP dir.
// Ví dụ: {bin}\php\current → {bin}\php\php-8.2.10
//
// Thử theo thứ tự:
//  1. os.Symlink (cần admin hoặc Developer Mode)
//  2. mklink /J (directory junction — không cần admin)
func (s *Switcher) updateWindowsSymlink(phpExe string) error {
	phpDir := filepath.Dir(phpExe)
	parentDir := filepath.Dir(phpDir)
	linkPath := filepath.Join(parentDir, "current")

	// Xóa symlink/junction cũ (bỏ qua lỗi nếu không tồn tại)
	_ = os.Remove(linkPath)
	// os.Remove không xóa được junction — dùng RemoveAll để chắc
	_ = os.RemoveAll(linkPath)

	// 1. Thử symbolic link (cần admin hoặc Developer Mode)
	if err := os.Symlink(phpDir, linkPath); err == nil {
		return nil
	}

	// 2. Fallback: directory junction (không cần admin)
	out, err := exec.Command("cmd", "/C", "mklink", "/J",
		filepath.FromSlash(linkPath),
		filepath.FromSlash(phpDir),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot create PHP junction %s → %s: %w — %s",
			linkPath, phpDir, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ─── Persistence ─────────────────────────────────────────────────────────────

func (s *Switcher) load() {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.state)
}

func (s *Switcher) save() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(s.statePath), 0755)
	return os.WriteFile(s.statePath, data, 0644)
}
