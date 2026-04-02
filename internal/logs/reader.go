package logs

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// LogLevel phân loại dòng log
type LogLevel string

const (
	LevelError   LogLevel = "error"
	LevelWarning LogLevel = "warning"
	LevelInfo    LogLevel = "info"
	LevelDebug   LogLevel = "debug"
)

// LogEntry một dòng log đã được parse
type LogEntry struct {
	Service   string   `json:"service"`
	Line      string   `json:"line"`
	Level     LogLevel `json:"level"`
	Timestamp string   `json:"timestamp"`
}

// LogPaths đường dẫn log mặc định theo từng service và OS
func LogPaths(logRoot string) map[string][]string {
	if logRoot == "" {
		logRoot = defaultLogRoot()
	}
	return map[string][]string{
		"apache": {
			filepath.Join(logRoot, "apache", "error.log"),
			filepath.Join(logRoot, "apache", "access.log"),
		},
		"nginx": {
			filepath.Join(logRoot, "nginx", "error.log"),
			filepath.Join(logRoot, "nginx", "access.log"),
		},
		"mysql": {
			filepath.Join(logRoot, "mysql", "mysql_error.log"),
		},
		"postgres": {
			filepath.Join(logRoot, "postgres", "postgres.log"),
		},
		"mongodb": {
			filepath.Join(logRoot, "mongodb", "mongod.log"),
		},
		"php": {
			filepath.Join(logRoot, "php", "php_error.log"),
		},
	}
}

func defaultLogRoot() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\laragon\logs`
	case "darwin":
		return "/usr/local/var/log"
	default:
		return "/var/log"
	}
}

// ReadLastLines đọc N dòng cuối của file log
func ReadLastLines(path string, n int) ([]LogEntry, error) {
	service := serviceFromPath(path)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, nil // file chưa có log thì trả empty
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Lấy N dòng cuối
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	entries := make([]LogEntry, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		entries = append(entries, LogEntry{
			Service:   service,
			Line:      line,
			Level:     detectLevel(line),
			Timestamp: time.Now().Format("15:04:05"),
		})
	}
	return entries, nil
}

// Watch theo dõi file log và gửi dòng mới qua channel
func Watch(ctx context.Context, path string, out chan<- LogEntry) error {
	service := serviceFromPath(path)

	// Tạo file nếu chưa tồn tại
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	// Seek to end để chỉ đọc log mới
	f.Seek(0, io.SeekEnd)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		f.Close()
		return err
	}

	// Watch thư mục cha (Windows cần watch dir thay vì file)
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		f.Close()
		watcher.Close()
		return err
	}

	go func() {
		defer f.Close()
		defer watcher.Close()

		reader := bufio.NewReader(f)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				if !strings.EqualFold(filepath.Clean(event.Name), filepath.Clean(path)) {
					continue
				}
				// Đọc tất cả dòng mới
				for {
					line, err := reader.ReadString('\n')
					line = strings.TrimRight(line, "\r\n")
					if line != "" {
						out <- LogEntry{
							Service:   service,
							Line:      line,
							Level:     detectLevel(line),
							Timestamp: time.Now().Format("15:04:05"),
						}
					}
					if err != nil {
						break // io.EOF — chờ lần write tiếp
					}
				}
			case <-watcher.Errors:
				// ignore watcher errors
			}
		}
	}()

	return nil
}

// detectLevel phát hiện mức độ log từ nội dung dòng
func detectLevel(line string) LogLevel {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal") || strings.Contains(lower, "crit"):
		return LevelError
	case strings.Contains(lower, "warn"):
		return LevelWarning
	case strings.Contains(lower, "debug") || strings.Contains(lower, "notice"):
		return LevelDebug
	default:
		return LevelInfo
	}
}

func serviceFromPath(path string) string {
	path = strings.ToLower(filepath.ToSlash(path))
	for _, svc := range []string{"apache", "nginx", "mysql", "php", "redis"} {
		if strings.Contains(path, svc) {
			return svc
		}
	}
	return "system"
}
