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
	"time"
)

// Manager quản lý tất cả services
type Manager struct {
	mu       sync.RWMutex
	services map[ServiceName]*serviceProcess
}

type serviceProcess struct {
	info         ServiceInfo
	cmd          *exec.Cmd
	binDir       string      // thư mục chứa binary của service
	dataDir      string      // thư mục lưu data của service (MySQL); rỗng = không dùng
	logDir       string      // thư mục lưu log tập trung; rỗng = dùng default của service
	restartTimes []time.Time // timestamps of recent auto-restarts
	onCrash      func(ServiceName)
}

// NewManager khởi tạo Manager.
// binPaths:  ServiceName → thư mục chứa executable.
// dataPaths: ServiceName → thư mục data (chỉ MySQL cần; các service khác bỏ qua).
// logPaths:  ServiceName → thư mục log tập trung (tạo sẵn nếu chưa có).
// ports:     ServiceName → port (từ config; dùng để hiển thị trên Dashboard).
func NewManager(binPaths, dataPaths, logPaths map[ServiceName]string, ports map[ServiceName]int) *Manager {
	bin := func(name ServiceName) string { return binPaths[name] }
	data := func(name ServiceName) string { return dataPaths[name] }
	log := func(name ServiceName) string {
		dir := logPaths[name]
		if dir != "" {
			os.MkdirAll(dir, 0755) //nolint:errcheck
		}
		return dir
	}
	port := func(name ServiceName, fallback int) int {
		if p, ok := ports[name]; ok && p > 0 {
			return p
		}
		return fallback
	}
	return &Manager{
		services: map[ServiceName]*serviceProcess{
			ServiceApache:   {info: ServiceInfo{Name: ServiceApache, Display: "Apache", Status: StatusStopped, Port: port(ServiceApache, 80), Version: "2.4"}, binDir: bin(ServiceApache), logDir: log(ServiceApache)},
			ServiceNginx:    {info: ServiceInfo{Name: ServiceNginx, Display: "Nginx", Status: StatusStopped, Port: port(ServiceNginx, 8080), Version: "1.25"}, binDir: bin(ServiceNginx), logDir: log(ServiceNginx)},
			ServiceMySQL:    {info: ServiceInfo{Name: ServiceMySQL, Display: "MySQL", Status: StatusStopped, Port: port(ServiceMySQL, 3306), Version: "8.0"}, binDir: bin(ServiceMySQL), dataDir: data(ServiceMySQL), logDir: log(ServiceMySQL)},
			ServicePostgres: {info: ServiceInfo{Name: ServicePostgres, Display: "PostgreSQL", Status: StatusStopped, Port: port(ServicePostgres, 5432), Version: "17"}, binDir: bin(ServicePostgres), dataDir: data(ServicePostgres), logDir: log(ServicePostgres)},
			ServiceMongoDB:  {info: ServiceInfo{Name: ServiceMongoDB, Display: "MongoDB", Status: StatusStopped, Port: port(ServiceMongoDB, 27017), Version: "8.0"}, binDir: bin(ServiceMongoDB), dataDir: data(ServiceMongoDB), logDir: log(ServiceMongoDB)},
			ServiceRedis:    {info: ServiceInfo{Name: ServiceRedis, Display: "Redis", Status: StatusStopped, Port: port(ServiceRedis, 6379), Version: "7.0"}, binDir: bin(ServiceRedis), logDir: log(ServiceRedis)},
			// PHP-CGI chạy ở port 9000 — cần khi dùng Nginx (FastCGI proxy).
			// Khi dùng Apache: không cần start (PHP loaded as module).
			ServicePHP: {info: ServiceInfo{Name: ServicePHP, Display: "PHP", Status: StatusStopped, Port: port(ServicePHP, 9000), Version: "8.2"}, binDir: bin(ServicePHP), logDir: log(ServicePHP)},
		},
	}
}

// serviceOrder định nghĩa thứ tự hiển thị cố định để UI không bị nhảy
var serviceOrder = []ServiceName{
	ServiceApache,
	ServiceMySQL,
	ServicePostgres,
	ServiceMongoDB,
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
	killStaleProcesses(name, binDir, dataDir)

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
	svc.info.UptimeSince = time.Now().Unix()
	svc.info.CrashLoop = false
	m.mu.Unlock()

	// Watch process in background — nếu exit sớm thì báo lỗi + trigger crash callback
	go func() {
		exitErr := cmd.Wait()
		m.mu.Lock()
		if svc.info.Status == StatusRunning {
			svc.info.PID = 0
			svc.info.UptimeSince = 0
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
			cb := svc.onCrash
			m.mu.Unlock()
			if cb != nil {
				cb(name)
			}
			return
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

	if svc.info.Status != StatusRunning && svc.info.Status != StatusError {
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
	svc.info.UptimeSince = 0
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

// UpdatePort cập nhật port hiển thị của một service tại runtime.
func (m *Manager) UpdatePort(name ServiceName, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.info.Port = port
	}
}

// SetCrashCallback đặt callback cho tất cả services khi crash
func (m *Manager) SetCrashCallback(cb func(ServiceName)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, svc := range m.services {
		svc.onCrash = cb
	}
}

// SetAutoRecover cập nhật flag auto_recover cho service
func (m *Manager) SetAutoRecover(name ServiceName, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.info.AutoRecover = enabled
	}
}

// IsCrashLoop kiểm tra service có đang crash loop (>3 restarts/phút)
func (m *Manager) IsCrashLoop(name ServiceName) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	svc, ok := m.services[name]
	if !ok {
		return false
	}
	cutoff := time.Now().Add(-1 * time.Minute)
	recent := make([]time.Time, 0, len(svc.restartTimes))
	for _, t := range svc.restartTimes {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	svc.restartTimes = recent
	return len(recent) > 3
}

// RecordRestart ghi nhận một lần auto-restart
func (m *Manager) RecordRestart(name ServiceName) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.restartTimes = append(svc.restartTimes, time.Now())
		svc.info.RestartCount = len(svc.restartTimes)
	}
}

// SetCrashLoop đánh dấu service đang crash loop
func (m *Manager) SetCrashLoop(name ServiceName, loop bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc, ok := m.services[name]; ok {
		svc.info.CrashLoop = loop
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
	ServiceApache:   {"httpd.exe", "httpd"},
	ServiceNginx:    {"nginx.exe", "nginx"},
	ServiceMySQL:    {"mysqld.exe", "mysqld"},
	ServicePostgres: {"postgres.exe", "postgres"},
	ServiceMongoDB:  {"mongod.exe", "mongod"},
	ServiceRedis:    {"redis-server.exe", "redis-server"},
	ServicePHP:      {"php-cgi.exe", "php-fpm"},
}

// killStaleProcesses dừng mọi tiến trình cũ của service (app crash, orphaned process).
// Trên Windows: dùng wmic để chỉ kill process có executable path khớp với binDir.
// Điều này tránh kill nhầm Apache/MySQL của Laragon, XAMPP, v.v.
// Với MySQL còn xóa luôn file .pid trong dataDir để tránh lỗi "already running".
func killStaleProcesses(name ServiceName, binDir, dataDir string) {
	exes, ok := staleExeNames[name]
	if !ok || binDir == "" {
		return
	}

	if runtime.GOOS == "windows" {
		for _, exe := range exes {
			if !strings.HasSuffix(exe, ".exe") {
				continue
			}
			killByPathWindows(exe, binDir)
		}
	} else {
		for _, exe := range exes {
			if strings.HasSuffix(exe, ".exe") {
				continue
			}
			// Trên Unix: pkill -f với đường dẫn đầy đủ để chỉ kill process của Stacknest
			fullPath := filepath.Join(binDir, exe)
			exec.Command("pkill", "-f", fullPath).Run() //nolint:errcheck
		}
	}

	// MySQL/PostgreSQL: xóa stale pid file nếu có
	if (name == ServiceMySQL || name == ServicePostgres) && dataDir != "" {
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

// killByPathWindows tìm process theo tên và chỉ kill nếu executable path nằm trong binDir.
// Sử dụng wmic để lấy ExecutablePath, so sánh với binDir trước khi kill.
func killByPathWindows(exeName, binDir string) {
	// Dùng wmic để lấy danh sách process có tên khớp, kèm executable path
	out, err := exec.Command("wmic", "process", "where",
		fmt.Sprintf(`Name='%s'`, exeName),
		"get", "ExecutablePath,ProcessId", "/FORMAT:CSV").Output()
	if err != nil {
		return
	}

	targetDir := strings.ToLower(filepath.Clean(binDir))

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		// CSV format: Node,ExecutablePath,ProcessId
		procPath := strings.ToLower(filepath.Clean(strings.TrimSpace(parts[1])))
		pidStr := strings.TrimSpace(parts[2])

		// Chỉ kill nếu executable nằm trong thư mục binDir của Stacknest
		if strings.HasPrefix(procPath, targetDir) {
			pid, _ := strconv.Atoi(pidStr)
			if pid > 0 {
				exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run() //nolint:errcheck
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
	case ServicePostgres:
		return buildPostgresCmd(binDir, dataDir, logDir)
	case ServiceMongoDB:
		return buildMongoDBCmd(binDir, dataDir, logDir)
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

func buildPostgresCmd(binDir, dataDir, logDir string) (*exec.Cmd, error) {
	// PostgreSQL cần data directory đã được initdb trước khi start.
	// -D chỉ định datadir, -p chỉ định port (dùng default 5432 nếu không set).
	args := []string{"-D", filepath.ToSlash(dataDir)}
	if logDir != "" {
		logFile := filepath.ToSlash(filepath.Join(logDir, "postgres.log"))
		args = append(args, "-l", logFile)
	}
	if runtime.GOOS == "windows" {
		return exec.Command(exePath(binDir, "postgres.exe"), args...), nil
	}
	return exec.Command(exePath(binDir, "postgres"), args...), nil
}

func buildMongoDBCmd(binDir, dataDir, logDir string) (*exec.Cmd, error) {
	// MongoDB: --dbpath chỉ định thư mục lưu data, --bind_ip giới hạn localhost.
	args := []string{
		"--dbpath", filepath.ToSlash(dataDir),
		"--bind_ip", "127.0.0.1",
	}
	if logDir != "" {
		logFile := filepath.ToSlash(filepath.Join(logDir, "mongod.log"))
		args = append(args, "--logpath", logFile, "--logappend")
	}
	if runtime.GOOS == "windows" {
		return exec.Command(exePath(binDir, "mongod.exe"), args...), nil
	}
	return exec.Command(exePath(binDir, "mongod"), args...), nil
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
