package services

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Manager quản lý tất cả services
type Manager struct {
	mu       sync.RWMutex
	services map[ServiceName]*serviceProcess
}

type serviceProcess struct {
	info    ServiceInfo
	cmd     *exec.Cmd
	binDir  string // thư mục chứa binary của service
	dataDir string // thư mục lưu data của service (MySQL); rỗng = không dùng
	logDir  string // thư mục lưu log tập trung; rỗng = dùng default của service
}

// NewManager khởi tạo Manager.
// binPaths:  ServiceName → thư mục chứa executable.
// dataPaths: ServiceName → thư mục data (chỉ MySQL cần; các service khác bỏ qua).
// logPaths:  ServiceName → thư mục log tập trung (tạo sẵn nếu chưa có).
func NewManager(binPaths, dataPaths, logPaths map[ServiceName]string) *Manager {
	bin := func(name ServiceName) string { return binPaths[name] }
	data := func(name ServiceName) string { return dataPaths[name] }
	log := func(name ServiceName) string {
		dir := logPaths[name]
		if dir != "" {
			os.MkdirAll(dir, 0755) //nolint:errcheck
		}
		return dir
	}
	return &Manager{
		services: map[ServiceName]*serviceProcess{
			ServiceApache: {info: ServiceInfo{Name: ServiceApache, Display: "Apache", Status: StatusStopped, Port: 80, Version: "2.4"}, binDir: bin(ServiceApache), logDir: log(ServiceApache)},
			ServiceNginx:  {info: ServiceInfo{Name: ServiceNginx, Display: "Nginx", Status: StatusStopped, Port: 8080, Version: "1.25"}, binDir: bin(ServiceNginx), logDir: log(ServiceNginx)},
			ServiceMySQL:  {info: ServiceInfo{Name: ServiceMySQL, Display: "MySQL", Status: StatusStopped, Port: 3306, Version: "8.0"}, binDir: bin(ServiceMySQL), dataDir: data(ServiceMySQL), logDir: log(ServiceMySQL)},
			ServiceRedis:  {info: ServiceInfo{Name: ServiceRedis, Display: "Redis", Status: StatusStopped, Port: 6379, Version: "7.0"}, binDir: bin(ServiceRedis), logDir: log(ServiceRedis)},
			// PHP-CGI chạy ở port 9000 — cần khi dùng Nginx (FastCGI proxy).
			// Khi dùng Apache: không cần start (PHP loaded as module).
			ServicePHP: {info: ServiceInfo{Name: ServicePHP, Display: "PHP", Status: StatusStopped, Port: 9000, Version: "8.2"}, binDir: bin(ServicePHP), logDir: log(ServicePHP)},
		},
	}
}

// serviceOrder định nghĩa thứ tự hiển thị cố định để UI không bị nhảy
var serviceOrder = []ServiceName{
	ServiceApache,
	ServiceMySQL,
	ServicePHP,
	ServiceRedis,
	ServiceNginx,
}

// GetAll trả về danh sách services theo thứ tự cố định
func (m *Manager) GetAll() []ServiceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ServiceInfo, 0, len(serviceOrder))
	for _, name := range serviceOrder {
		if svc, ok := m.services[name]; ok {
			result = append(result, svc.info)
		}
	}
	return result
}

// GetOne trả về thông tin một service
func (m *Manager) GetOne(name ServiceName) (ServiceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	svc, ok := m.services[name]
	if !ok {
		return ServiceInfo{}, fmt.Errorf("service %s not found", name)
	}
	return svc.info, nil
}

// Start khởi động một service
func (m *Manager) Start(name ServiceName) error {
	m.mu.Lock()
	svc, ok := m.services[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("service %s not found", name)
	}
	svc.info.Status = StatusStarting
	binDir := svc.binDir
	dataDir := svc.dataDir
	logDir := svc.logDir
	m.mu.Unlock()

	// Dọn dẹp tiến trình cũ còn sót lại (do app crash, force-quit, v.v.)
	killStaleProcesses(name, dataDir)

	cmd, err := buildCommand(name, binDir, dataDir, logDir)
	if err != nil {
		m.setError(name, err.Error())
		return err
	}

	// Capture stderr để báo lỗi khi process exit sớm
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		m.setError(name, err.Error())
		return fmt.Errorf("failed to start %s: %w", name, err)
	}

	m.mu.Lock()
	svc.cmd = cmd
	svc.info.PID = cmd.Process.Pid
	svc.info.Status = StatusRunning
	svc.info.Error = ""
	m.mu.Unlock()

	// Watch process in background — nếu exit sớm thì báo lỗi
	go func() {
		exitErr := cmd.Wait()
		m.mu.Lock()
		if svc.info.Status == StatusRunning {
			svc.info.PID = 0
			if exitErr != nil {
				errMsg := strings.TrimSpace(stderr.String())
				if errMsg == "" {
					errMsg = exitErr.Error()
				}
				svc.info.Status = StatusError
				svc.info.Error = errMsg
			} else {
				svc.info.Status = StatusStopped
			}
		}
		m.mu.Unlock()
	}()

	return nil
}

// Stop dừng một service
func (m *Manager) Stop(name ServiceName) error {
	m.mu.Lock()
	svc, ok := m.services[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("service %s not found", name)
	}

	if svc.info.Status != StatusRunning {
		m.mu.Unlock()
		return fmt.Errorf("service %s is not running", name)
	}

	svc.info.Status = StatusStopping
	cmd := svc.cmd
	m.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		pid := cmd.Process.Pid
		if err := cmd.Process.Kill(); err != nil {
			m.setError(name, err.Error())
			return err
		}
		// Trên Windows, kill process tree để dọn sạch worker processes (nginx, httpd...)
		if runtime.GOOS == "windows" {
			exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run() //nolint:errcheck
		}
	}

	m.mu.Lock()
	svc.info.Status = StatusStopped
	svc.info.PID = 0
	svc.cmd = nil
	m.mu.Unlock()

	return nil
}

// Restart khởi động lại service
func (m *Manager) Restart(name ServiceName) error {
	if err := m.Stop(name); err != nil {
		// Bỏ qua lỗi nếu service đang stopped
		if svc, _ := m.GetOne(name); svc.Status != StatusStopped {
			return err
		}
	}
	return m.Start(name)
}

// IsPortInUse kiểm tra port có đang được dùng không
func (m *Manager) IsPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}

// StopAll dừng tất cả services
func (m *Manager) StopAll() {
	var wg sync.WaitGroup
	for name := range m.services {
		wg.Add(1)
		go func(n ServiceName) {
			defer wg.Done()
			m.Stop(n)
		}(name)
	}
	wg.Wait()
}

// UpdateBinDir cập nhật đường dẫn binary directory của một service tại runtime.
func (m *Manager) UpdateBinDir(name ServiceName, binDir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.binDir = binDir
	}
}

// UpdateDataDir cập nhật đường dẫn data directory của MySQL tại runtime.
func (m *Manager) UpdateDataDir(name ServiceName, dataDir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.dataDir = dataDir
	}
}

func (m *Manager) setError(name ServiceName, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.info.Status = StatusError
		svc.info.Error = errMsg
	}
}

// exePath trả về đường dẫn đầy đủ tới executable trong binDir.
func exePath(binDir, name string) string {
	if binDir == "" {
		return name
	}
	return filepath.Join(binDir, name)
}

// staleExeNames maps service → executable file name(s) to kill before starting.
var staleExeNames = map[ServiceName][]string{
	ServiceApache: {"httpd.exe", "httpd"},
	ServiceNginx:  {"nginx.exe", "nginx"},
	ServiceMySQL:  {"mysqld.exe", "mysqld"},
	ServiceRedis:  {"redis-server.exe", "redis-server"},
	ServicePHP:    {"php-cgi.exe", "php-fpm"},
}

// killStaleProcesses dừng mọi tiến trình cũ của service (app crash, orphaned process).
// Trên Windows dùng taskkill /IM; trên Unix dùng pkill.
// Với MySQL còn xóa luôn file .pid trong dataDir để tránh lỗi "already running".
func killStaleProcesses(name ServiceName, dataDir string) {
	exes, ok := staleExeNames[name]
	if !ok {
		return
	}

	if runtime.GOOS == "windows" {
		for _, exe := range exes {
			if strings.HasSuffix(exe, ".exe") {
				exec.Command("taskkill", "/F", "/IM", exe).Run() //nolint:errcheck
			}
		}
	} else {
		for _, exe := range exes {
			if !strings.HasSuffix(exe, ".exe") {
				exec.Command("pkill", "-f", exe).Run() //nolint:errcheck
			}
		}
	}

	// MySQL: xóa stale pid file nếu có
	if name == ServiceMySQL && dataDir != "" {
		entries, err := os.ReadDir(dataDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".pid") {
					os.Remove(filepath.Join(dataDir, e.Name())) //nolint:errcheck
				}
			}
		}
	}
}

// buildCommand tạo command để chạy service theo OS.
func buildCommand(name ServiceName, binDir, dataDir, logDir string) (*exec.Cmd, error) {
	switch name {
	case ServiceApache:
		return buildApacheCmd(binDir, logDir)
	case ServiceNginx:
		return buildNginxCmd(binDir)
	case ServiceMySQL:
		return buildMySQLCmd(binDir, dataDir, logDir)
	case ServiceRedis:
		return buildRedisCmd(binDir, logDir)
	case ServicePHP:
		return buildPHPCmd(binDir, logDir)
	}
	return nil, fmt.Errorf("unknown service: %s", name)
}

func buildApacheCmd(binDir, logDir string) (*exec.Cmd, error) {
	// binDir = {root}/bin/apache/bin — serverRoot là thư mục cha
	serverRoot := filepath.Dir(binDir)
	args := []string{"-d", serverRoot}
	if logDir != "" {
		// -c injects directives AFTER parsing httpd.conf (our paths override defaults)
		errorLog := filepath.ToSlash(filepath.Join(logDir, "error.log"))
		accessLog := filepath.ToSlash(filepath.Join(logDir, "access.log"))
		args = append(args,
			"-c", fmt.Sprintf(`ErrorLog "%s"`, errorLog),
			"-c", fmt.Sprintf(`CustomLog "%s" combined`, accessLog),
		)
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(exePath(binDir, "httpd.exe"), args...)
	} else {
		args = append(args, "-DFOREGROUND")
		cmd = exec.Command(exePath(binDir, "httpd"), args...)
	}
	cmd.Dir = serverRoot
	return cmd, nil
}

func buildNginxCmd(binDir string) (*exec.Cmd, error) {
	// -p đặt prefix directory để nginx tìm conf/, logs/ đúng chỗ
	// Log paths của nginx được set trong nginx.conf (xem ensureNginxConfig).
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(exePath(binDir, "nginx.exe"), "-p", binDir)
	} else {
		cmd = exec.Command(exePath(binDir, "nginx"), "-p", binDir)
	}
	cmd.Dir = binDir
	return cmd, nil
}

func buildMySQLCmd(binDir, dataDir, logDir string) (*exec.Cmd, error) {
	// basedir = thư mục cha của bin/ (= phiên bản MySQL root, chứa share/)
	basedir := filepath.ToSlash(filepath.Dir(binDir))
	args := []string{"--basedir=" + basedir}
	if dataDir != "" {
		// MySQL trên Windows xử lý tốt hơn với forward slash
		args = append(args, "--datadir="+filepath.ToSlash(dataDir))
	}
	if logDir != "" {
		// Ghi log vào thư mục tập trung; không dùng --console khi có log file
		logFile := filepath.ToSlash(filepath.Join(logDir, "mysql_error.log"))
		args = append(args, "--log-error="+logFile)
	} else {
		args = append(args, "--console")
	}
	if runtime.GOOS == "windows" {
		return exec.Command(exePath(binDir, "mysqld.exe"), args...), nil
	}
	return exec.Command(exePath(binDir, "mysqld"), append(args, "--user=mysql")...), nil
}

func buildRedisCmd(binDir, logDir string) (*exec.Cmd, error) {
	var args []string
	if logDir != "" {
		logFile := filepath.ToSlash(filepath.Join(logDir, "redis.log"))
		args = append(args, "--logfile", logFile)
	}
	if runtime.GOOS == "windows" {
		return exec.Command(exePath(binDir, "redis-server.exe"), args...), nil
	}
	return exec.Command(exePath(binDir, "redis-server"), args...), nil
}

func buildPHPCmd(binDir, logDir string) (*exec.Cmd, error) {
	// Chạy PHP ở chế độ FastCGI server, bind port 9000.
	// Nginx sẽ proxy PHP requests tới 127.0.0.1:9000.
	// Với Apache, không cần start service này (PHP là module của Apache).
	args := []string{"-b", "127.0.0.1:9000"}
	if logDir != "" {
		args = append(args, "-d", "error_log="+filepath.Join(logDir, "php_error.log"))
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command(exePath(binDir, "php-cgi.exe"), args...)
		cmd.Dir = binDir
		return cmd, nil
	}
	// Linux/macOS: dùng php-fpm
	cmd := exec.Command(exePath(binDir, "php-fpm"), "--nodaemonize", "--bind-address=127.0.0.1:9000")
	cmd.Dir = binDir
	return cmd, nil
}
