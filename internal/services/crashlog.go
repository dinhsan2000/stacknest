package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CrashLogger ghi crash events vào file log riêng
type CrashLogger struct {
	logDir string
}

// NewCrashLogger tạo crash logger
func NewCrashLogger(logDir string) *CrashLogger {
	_ = os.MkdirAll(logDir, 0755)
	return &CrashLogger{logDir: logDir}
}

// Log ghi một crash event
func (cl *CrashLogger) Log(service ServiceName, errMsg string, autoRestarted bool) {
	f, err := os.OpenFile(
		filepath.Join(cl.logDir, "crash.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644,
	)
	if err != nil {
		return
	}
	defer f.Close()

	action := "not restarted"
	if autoRestarted {
		action = "auto-restarted"
	}
	fmt.Fprintf(f, "[%s] %s crashed: %s (%s)\n",
		time.Now().Format("2006-01-02 15:04:05"), service, errMsg, action)
}
