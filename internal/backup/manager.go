package backup

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"stacknest/internal/downloader"
	"strings"
	"sync"
	"time"
)

// BackupInfo mô tả một file backup
type BackupInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Database  string `json:"database"`
	CreatedAt string `json:"created_at"`
}

// Manager quản lý backup/restore MySQL
type Manager struct {
	mu      sync.Mutex
	binPath string
	dataPath string
	port    int
	EmitFn  func(event string, data ...interface{})
}

// NewManager tạo backup manager
func NewManager(binPath, dataPath string, port int) *Manager {
	return &Manager{
		binPath:  binPath,
		dataPath: dataPath,
		port:     port,
	}
}

// UpdatePort cập nhật port MySQL
func (m *Manager) UpdatePort(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.port = port
}

// BackupDir trả về thư mục chứa backups
func (m *Manager) BackupDir() string {
	dir := filepath.Join(m.dataPath, "backups")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// mysqlBin trả về đường dẫn đến một binary trong MySQL bin dir
func (m *Manager) mysqlBin(name string) (string, error) {
	binDir := downloader.ActiveExeDir(m.binPath, "mysql")
	if binDir == "" {
		return "", fmt.Errorf("MySQL binary not found — install a MySQL version first")
	}
	exe := filepath.Join(binDir, name)
	if _, err := os.Stat(exe); err != nil {
		return "", fmt.Errorf("%s not found at %s", name, exe)
	}
	return exe, nil
}

func (m *Manager) emit(event string, data ...interface{}) {
	if m.EmitFn != nil {
		m.EmitFn(event, data...)
	}
}

// CreateBackup chạy mysqldump và nén thành .sql.gz
// database: tên DB cụ thể hoặc "all" cho --all-databases
func (m *Manager) CreateBackup(database string) (*BackupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mysqldump, err := m.mysqlBin("mysqldump.exe")
	if err != nil {
		return nil, err
	}

	if database == "" {
		database = "all"
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s-%s.sql.gz", database, timestamp)
	outPath := filepath.Join(m.BackupDir(), filename)

	// Build mysqldump args
	args := []string{
		"--host=127.0.0.1",
		fmt.Sprintf("--port=%d", m.port),
		"--user=root",
		"--single-transaction",
		"--routines",
		"--triggers",
	}
	if database == "all" {
		args = append(args, "--all-databases")
	} else {
		args = append(args, "--databases", database)
	}

	m.emit("backup:progress", map[string]string{"status": "dumping", "filename": filename})

	cmd := exec.Command(mysqldump, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mysqldump failed to start: %w", err)
	}

	// Pipe mysqldump output → gzip → file
	outFile, err := os.Create(outPath)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}

	gz := gzip.NewWriter(outFile)
	_, copyErr := io.Copy(gz, stdout)
	gz.Close()
	outFile.Close()

	waitErr := cmd.Wait()
	if copyErr != nil {
		os.Remove(outPath)
		return nil, fmt.Errorf("backup write error: %w", copyErr)
	}
	if waitErr != nil {
		os.Remove(outPath)
		return nil, fmt.Errorf("mysqldump error: %w", waitErr)
	}

	stat, _ := os.Stat(outPath)
	info := &BackupInfo{
		Name:      filename,
		Size:      stat.Size(),
		Database:  database,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	m.emit("backup:progress", map[string]string{"status": "done", "filename": filename})
	return info, nil
}

// ListBackups liệt kê tất cả backups
func (m *Manager) ListBackups() ([]BackupInfo, error) {
	dir := m.BackupDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []BackupInfo{}, nil
	}

	var backups []BackupInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql.gz") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}

		// Parse database name from filename: "{database}-{timestamp}.sql.gz"
		name := e.Name()
		db := "all"
		if idx := strings.LastIndex(name, "-2"); idx > 0 {
			db = name[:idx]
		}

		backups = append(backups, BackupInfo{
			Name:      name,
			Size:      info.Size(),
			Database:  db,
			CreatedAt: info.ModTime().Format(time.RFC3339),
		})
	}

	// Sort by date descending (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

// RestoreBackup khôi phục từ file backup .sql.gz
func (m *Manager) RestoreBackup(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate filename — chống path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		return fmt.Errorf("invalid filename")
	}

	mysqlExe, err := m.mysqlBin("mysql.exe")
	if err != nil {
		return err
	}

	backupPath := filepath.Join(m.BackupDir(), filename)
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not found: %s", filename)
	}

	m.emit("backup:progress", map[string]string{"status": "restoring", "filename": filename})

	// Mở file → decompress gzip → pipe vào mysql
	f, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("invalid gzip file: %w", err)
	}
	defer gz.Close()

	args := []string{
		"--host=127.0.0.1",
		fmt.Sprintf("--port=%d", m.port),
		"--user=root",
	}

	cmd := exec.Command(mysqlExe, args...)
	cmd.Stdin = gz

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore failed: %s — %w", string(output), err)
	}

	m.emit("backup:progress", map[string]string{"status": "done", "filename": filename})
	return nil
}

// DeleteBackup xóa file backup
func (m *Manager) DeleteBackup(filename string) error {
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		return fmt.Errorf("invalid filename")
	}
	return os.Remove(filepath.Join(m.BackupDir(), filename))
}

// ListDatabases liệt kê databases trong MySQL
func (m *Manager) ListDatabases() ([]string, error) {
	mysqlExe, err := m.mysqlBin("mysql.exe")
	if err != nil {
		return nil, err
	}

	args := []string{
		"--host=127.0.0.1",
		fmt.Sprintf("--port=%d", m.port),
		"--user=root",
		"-e", "SHOW DATABASES",
		"--skip-column-names",
	}

	out, err := exec.Command(mysqlExe, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	systemDBs := map[string]bool{
		"information_schema": true,
		"performance_schema": true,
		"mysql":              true,
		"sys":                true,
	}

	var dbs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		db := strings.TrimSpace(line)
		if db != "" && !systemDBs[db] {
			dbs = append(dbs, db)
		}
	}
	return dbs, nil
}
