package terminal

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"

	gopty "github.com/aymanbagabas/go-pty"
)

// Session đại diện cho một phiên terminal
type Session struct {
	pty    gopty.Pty
	cmd    *gopty.Cmd
	cancel context.CancelFunc
}

// New tạo session terminal mới
// Trả về session, channel nhận output bytes, và error
func New(ctx context.Context, cwd string) (*Session, <-chan []byte, error) {
	ctx, cancel := context.WithCancel(ctx)

	pty, err := gopty.New()
	if err != nil {
		cancel()
		return nil, nil, err
	}

	shell, args := getShell()

	// Tạo command gắn với PTY
	cmd := pty.CommandContext(ctx, shell, args...)
	cmd.Env = os.Environ()
	if cwd != "" {
		if _, err := os.Stat(cwd); err == nil {
			cmd.Dir = cwd
		}
	}

	if err := cmd.Start(); err != nil {
		pty.Close()
		cancel()
		return nil, nil, err
	}

	out := make(chan []byte, 256)

	// Goroutine: đọc output từ PTY → channel
	go func() {
		defer close(out)
		buf := make([]byte, 4096)
		for {
			n, err := pty.Read(buf)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				select {
				case out <- data:
				case <-ctx.Done():
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					_ = err // PTY closed — bình thường
				}
				return
			}
		}
	}()

	return &Session{pty: pty, cmd: cmd, cancel: cancel}, out, nil
}

// Write gửi input từ user vào PTY
func (s *Session) Write(data []byte) error {
	_, err := s.pty.Write(data)
	return err
}

// Resize thay đổi kích thước terminal window
// go-pty: Resize(width, height int) — width=cols, height=rows
func (s *Session) Resize(rows, cols uint16) error {
	return s.pty.Resize(int(cols), int(rows))
}

// Close đóng session và kill process
func (s *Session) Close() {
	s.cancel()
	s.pty.Close()
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
}

// getShell trả về shell mặc định theo OS
func getShell() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		if ps, err := exec.LookPath("pwsh"); err == nil {
			return ps, []string{"-NoLogo"}
		}
		if ps, err := exec.LookPath("powershell"); err == nil {
			return ps, []string{"-NoLogo"}
		}
		return "cmd.exe", []string{}
	case "darwin":
		if zsh, err := exec.LookPath("zsh"); err == nil {
			return zsh, []string{}
		}
		return "/bin/bash", []string{}
	default:
		if bash, err := exec.LookPath("bash"); err == nil {
			return bash, []string{}
		}
		return "/bin/sh", []string{}
	}
}
