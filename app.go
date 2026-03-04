package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"stacknest/internal/config"
	"stacknest/internal/configeditor"
	"stacknest/internal/database"
	"stacknest/internal/downloader"
	"stacknest/internal/logs"
	"stacknest/internal/phpswitch"
	"stacknest/internal/portcheck"
	"stacknest/internal/services"
	"stacknest/internal/ssl"
	"stacknest/internal/terminal"
	"stacknest/internal/vhost"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App là struct chính expose APIs cho frontend
type App struct {
	ctx         context.Context
	cfg         *config.Config
	svcMgr      *services.Manager
	vhostMgr    *vhost.Manager
	phpSwitcher *phpswitch.Switcher
	cfgEditor   *configeditor.Manager
	sslMgr      *ssl.Manager
	adminerSrv  *database.Server
	logCancel   context.CancelFunc
	termMu      sync.Mutex
	termSession *terminal.Session
	dlMu        sync.Mutex
	dlCancels   map[string]context.CancelFunc
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		// Config rơi về defaults; app vẫn chạy được nhưng cần thông báo.
		// ctx chưa có ở đây nên dùng stderr.
		fmt.Fprintf(os.Stderr, "[stacknest] config warning: %v\n", err)
	}
	downloader.InitCatalog(cfg.RootPath)
	return &App{
		cfg:         cfg,
		svcMgr:      services.NewManager(serviceBinPaths(cfg), serviceDataPaths(cfg), serviceLogPaths(cfg), servicePortMap(cfg)),
		vhostMgr:    vhost.NewManager(cfg.RootPath),
		phpSwitcher: phpswitch.NewSwitcher(cfg.RootPath),
		cfgEditor:   configeditor.NewManager(cfg.RootPath),
		sslMgr:      ssl.NewManager(cfg.RootPath),
		adminerSrv:  database.NewServer(cfg.RootPath),
		dlCancels:   make(map[string]context.CancelFunc),
	}
}

// serviceBinPaths trả về đường dẫn exe dir của phiên bản active cho từng service.
// Đọc từ versions.json; fallback về phiên bản đầu tiên trong catalog nếu chưa đặt.
func serviceBinPaths(cfg *config.Config) map[services.ServiceName]string {
	return map[services.ServiceName]string{
		services.ServiceApache: downloader.ActiveExeDir(cfg.BinPath, "apache"),
		services.ServiceNginx:  downloader.ActiveExeDir(cfg.BinPath, "nginx"),
		services.ServiceMySQL:  downloader.ActiveExeDir(cfg.BinPath, "mysql"),
		services.ServiceRedis:  downloader.ActiveExeDir(cfg.BinPath, "redis"),
		services.ServicePHP:    downloader.ActiveExeDir(cfg.BinPath, "php"),
	}
}

// serviceDataPaths trả về thư mục data cho các service cần (hiện tại chỉ MySQL).
// MySQL dùng thư mục riêng theo phiên bản để tránh conflict khi switch version.
func serviceDataPaths(cfg *config.Config) map[services.ServiceName]string {
	avs := downloader.LoadActiveVersions(cfg.BinPath)
	return map[services.ServiceName]string{
		services.ServiceMySQL: cfg.MySQLDataDir(avs["mysql"]),
	}
}

// serviceLogPaths trả về thư mục log tập trung cho từng service.
// Nginx log được set trong nginx.conf nên không cần truyền qua đây.
func serviceLogPaths(cfg *config.Config) map[services.ServiceName]string {
	return map[services.ServiceName]string{
		services.ServiceApache: filepath.Join(cfg.LogPath, "apache"),
		services.ServiceNginx:  filepath.Join(cfg.LogPath, "nginx"),
		services.ServiceMySQL:  filepath.Join(cfg.LogPath, "mysql"),
		services.ServiceRedis:  filepath.Join(cfg.LogPath, "redis"),
		services.ServicePHP:    filepath.Join(cfg.LogPath, "php"),
	}
}

// servicePortMap trả về port từ config cho từng service.
func servicePortMap(cfg *config.Config) map[services.ServiceName]int {
	return map[services.ServiceName]int{
		services.ServiceApache: cfg.Apache.Port,
		services.ServiceNginx:  cfg.Nginx.Port,
		services.ServiceMySQL:  cfg.MySQL.Port,
		services.ServiceRedis:  cfg.Redis.Port,
		services.ServicePHP:    cfg.PHP.Port,
	}
}

// startup được gọi khi app khởi động
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Đảm bảo nginx.conf đúng (trỏ vào www/) nếu binary đã có
	a.ensureNginxConfig(false)

	// Đảm bảo MySQL data dir đã được khởi tạo trước khi start
	go a.ensureMySQLInit()

	// Auto start services nếu được cấu hình
	if a.cfg.AutoStart {
		go a.StartAll()
	}

	// Emit service status events định kỳ
	go a.watchServices()
}

// shutdown được gọi khi app đóng
func (a *App) shutdown(ctx context.Context) {
	a.adminerSrv.Stop()
	a.svcMgr.StopAll()
}

// ─── Service APIs ────────────────────────────────────────────────────────────

// GetServices trả về danh sách tất cả services
func (a *App) GetServices() []services.ServiceInfo {
	return a.servicesWithEnabled()
}

// SetServiceEnabled bật/tắt một service (ảnh hưởng đến StartAll).
func (a *App) SetServiceEnabled(name string, enabled bool) error {
	switch services.ServiceName(name) {
	case services.ServiceApache:
		a.cfg.Apache.Enabled = enabled
	case services.ServiceNginx:
		a.cfg.Nginx.Enabled = enabled
	case services.ServiceMySQL:
		a.cfg.MySQL.Enabled = enabled
	case services.ServicePHP:
		a.cfg.PHP.Enabled = enabled
	case services.ServiceRedis:
		a.cfg.Redis.Enabled = enabled
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
	if err := a.cfg.Save(); err != nil {
		return err
	}
	a.emitServiceUpdate()
	return nil
}

// StartService khởi động một service theo tên.
// Khi start Nginx, PHP (FastCGI) được tự động start cùng.
func (a *App) StartService(name string) error {
	err := a.svcMgr.Start(services.ServiceName(name))
	a.emitServiceUpdate()
	// Nginx cần PHP-CGI làm FastCGI backend → auto-start PHP
	if name == "nginx" && err == nil {
		go func() {
			a.svcMgr.Start(services.ServicePHP) //nolint:errcheck
			a.emitServiceUpdate()
		}()
	}
	return err
}

// StopService dừng một service theo tên.
// Khi stop Nginx, PHP (FastCGI) cũng được tự động stop.
func (a *App) StopService(name string) error {
	err := a.svcMgr.Stop(services.ServiceName(name))
	a.emitServiceUpdate()
	// Auto-stop PHP khi Nginx dừng
	if name == "nginx" {
		go func() {
			a.svcMgr.Stop(services.ServicePHP) //nolint:errcheck
			a.emitServiceUpdate()
		}()
	}
	return err
}

// RestartService khởi động lại một service
func (a *App) RestartService(name string) error {
	err := a.svcMgr.Restart(services.ServiceName(name))
	a.emitServiceUpdate()
	return err
}

// StartAll khởi động các services đang được enabled.
// Nginx enabled sẽ tự động start PHP-CGI cùng.
func (a *App) StartAll() {
	if a.cfg.Apache.Enabled {
		go a.StartService("apache")
	}
	if a.cfg.MySQL.Enabled {
		go a.StartService("mysql")
	}
	// StartService("nginx") tự động start PHP cùng
	if a.cfg.Nginx.Enabled {
		go a.StartService("nginx")
	}
	if a.cfg.Redis.Enabled {
		go a.StartService("redis")
	}
}

// StopAll dừng tất cả services
func (a *App) StopAll() {
	a.svcMgr.StopAll()
	a.emitServiceUpdate()
}

// ─── Binary Management APIs ───────────────────────────────────────────────────

// GetBinaryStatus trả về trạng thái tất cả phiên bản của tất cả services
func (a *App) GetBinaryStatus() []downloader.ServiceVersionStatus {
	return downloader.GetStatus(a.cfg.BinPath)
}

// GetVersionCatalog trả về danh mục tất cả phiên bản có thể tải cho từng service
func (a *App) GetVersionCatalog() map[string]downloader.ServiceCatalog {
	return downloader.Catalog
}

// StartBinaryDownload bắt đầu tải binary cho service/version trong background.
// Events: "binary:progress" {service, version, pct float64}, "binary:done" {service, version, error string}
func (a *App) StartBinaryDownload(service, version string) error {
	key := service + "@" + version

	// Cancel download cũ nếu đang chạy cho cùng service@version
	a.dlMu.Lock()
	if cancel, ok := a.dlCancels[key]; ok {
		cancel()
	}
	dlCtx, cancel := context.WithCancel(a.ctx)
	a.dlCancels[key] = cancel
	a.dlMu.Unlock()

	go func() {
		defer func() {
			a.dlMu.Lock()
			delete(a.dlCancels, key)
			a.dlMu.Unlock()
		}()

		err := downloader.Download(dlCtx, service, version, a.cfg.BinPath, func(downloaded, total int64) {
			if total > 0 {
				pct := float64(downloaded) / float64(total) * 100
				runtime.EventsEmit(a.ctx, "binary:progress", map[string]any{
					"service": service,
					"version": version,
					"pct":     pct,
				})
			}
		})

		if err == nil {
			if service == "nginx" {
				a.ensureNginxConfig(true)
			}
			if service == "mysql" {
				a.ensureMySQLInit()
			}
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		runtime.EventsEmit(a.ctx, "binary:done", map[string]any{
			"service": service,
			"version": version,
			"error":   errMsg,
		})
	}()
	return nil
}

// CancelBinaryDownload hủy download đang chạy cho service/version.
func (a *App) CancelBinaryDownload(service, version string) {
	key := service + "@" + version
	a.dlMu.Lock()
	if cancel, ok := a.dlCancels[key]; ok {
		cancel()
		delete(a.dlCancels, key)
	}
	a.dlMu.Unlock()
}

// DeleteBinary xóa binary đã cài của service/version. Không cho xóa version active.
func (a *App) DeleteBinary(service, version string) error {
	return downloader.Delete(service, version, a.cfg.BinPath)
}

// SetActiveVersion đặt phiên bản active cho một service, cập nhật service manager.
func (a *App) SetActiveVersion(service, version string) error {
	if err := downloader.SetActiveVersion(a.cfg.BinPath, service, version); err != nil {
		return err
	}
	// Cập nhật binDir trong service manager để lần start tiếp theo dùng đúng binary
	newBinDir := downloader.ActiveExeDir(a.cfg.BinPath, service)
	a.svcMgr.UpdateBinDir(services.ServiceName(service), newBinDir)
	// Đồng bộ nginx.conf khi chuyển phiên bản nginx (force để ghi đè stock conf của version mới)
	if service == "nginx" {
		a.ensureNginxConfig(true)
	}
	// Đảm bảo data dir được khởi tạo khi chuyển phiên bản MySQL,
	// và cập nhật dataDir trong service manager để --datadir trỏ đúng version mới.
	if service == "mysql" {
		newDataDir := a.cfg.MySQLDataDir(version)
		a.svcMgr.UpdateDataDir(services.ServiceMySQL, newDataDir)
		go a.ensureMySQLInit()
	}
	return nil
}

// IsPortInUse kiểm tra port có đang được dùng không
func (a *App) IsPortInUse(port int) bool {
	return a.svcMgr.IsPortInUse(port)
}

// ─── Port Conflict APIs ───────────────────────────────────────────────────────

// CheckPortConflict kiểm tra port và trả về thông tin process đang chiếm
func (a *App) CheckPortConflict(port int) portcheck.ConflictInfo {
	return portcheck.Check(port)
}

// CheckAllPortConflicts kiểm tra tất cả ports của các services
func (a *App) CheckAllPortConflicts() []portcheck.ConflictInfo {
	svcs := a.svcMgr.GetAll()
	var conflicts []portcheck.ConflictInfo
	for _, svc := range svcs {
		if svc.Status != services.StatusRunning {
			info := portcheck.Check(svc.Port)
			if info.InUse {
				conflicts = append(conflicts, info)
			}
		}
	}
	return conflicts
}

// KillConflictProcess kill process đang chiếm port
func (a *App) KillConflictProcess(pid int) error {
	return portcheck.KillProcess(pid)
}

// ─── PHP Version APIs ────────────────────────────────────────────────────────

// GetPHPInstalls trả về danh sách PHP versions đã tìm thấy trên máy
func (a *App) GetPHPInstalls() []phpswitch.PHPInstall {
	return a.phpSwitcher.GetInstalls()
}

// GetActivePHP trả về PHP version đang active
func (a *App) GetActivePHP() *phpswitch.PHPInstall {
	return a.phpSwitcher.GetActive()
}

// SwitchPHP chuyển sang PHP version theo path.
// - Với Apache: restart Apache để load đúng PHP module.
// - Với Nginx: cập nhật binDir của PHP service và restart nó.
func (a *App) SwitchPHP(phpPath string) error {
	if err := a.phpSwitcher.Switch(phpPath); err != nil {
		return err
	}
	// Cập nhật binDir trong service manager cho PHP service (Nginx mode)
	// Dùng thư mục chứa PHP exe được chọn, không dùng downloader vì phpSwitcher
	// lưu vào php_versions.json (file khác với versions.json của downloader).
	newBinDir := filepath.Dir(phpPath)
	a.svcMgr.UpdateBinDir(services.ServicePHP, newBinDir)

	// Restart whichever is running
	if svc, _ := a.svcMgr.GetOne(services.ServiceApache); svc.Status == services.StatusRunning {
		return a.svcMgr.Restart(services.ServiceApache)
	}
	if svc, _ := a.svcMgr.GetOne(services.ServicePHP); svc.Status == services.StatusRunning {
		return a.svcMgr.Restart(services.ServicePHP)
	}
	return nil
}

// AddPHPPath thêm thư mục PHP tùy chỉnh để scan
func (a *App) AddPHPPath(dir string) error {
	return a.phpSwitcher.AddCustomPath(dir)
}

// ─── Virtual Host APIs ───────────────────────────────────────────────────────

// GetVirtualHosts trả về danh sách virtual hosts
func (a *App) GetVirtualHosts() []vhost.VirtualHost {
	return a.vhostMgr.GetAll()
}

// AddVirtualHost thêm virtual host mới. server là "apache" hoặc "nginx".
// Nếu ssl=true, tự động generate SSL certificate cho domain.
func (a *App) AddVirtualHost(name, domain, root, server string, ssl bool) error {
	if err := a.vhostMgr.Add(name, domain, root, server, ssl); err != nil {
		return err
	}
	// Auto generate SSL cert khi add vhost với SSL enabled
	if ssl {
		if _, _, err := a.sslMgr.GenerateCert(domain); err != nil {
			// Log warning nhưng không fail việc thêm vhost — user có thể generate lại sau
			runtime.LogWarningf(a.ctx, "Auto SSL cert generation failed for %s: %v", domain, err)
		}
	}
	return nil
}

// RemoveVirtualHost xóa virtual host
func (a *App) RemoveVirtualHost(domain string) error {
	return a.vhostMgr.Remove(domain)
}

// ─── Config APIs ─────────────────────────────────────────────────────────────

// GetConfig trả về cấu hình hiện tại
func (a *App) GetConfig() *config.Config {
	return a.cfg
}

// SaveConfig lưu cấu hình
func (a *App) SaveConfig(cfg config.Config) error {
	a.cfg = &cfg
	// Propagate port thay đổi vào service manager để Dashboard hiển thị đúng
	a.svcMgr.UpdatePort(services.ServiceApache, cfg.Apache.Port)
	a.svcMgr.UpdatePort(services.ServiceNginx, cfg.Nginx.Port)
	a.svcMgr.UpdatePort(services.ServiceMySQL, cfg.MySQL.Port)
	a.svcMgr.UpdatePort(services.ServicePHP, cfg.PHP.Port)
	a.svcMgr.UpdatePort(services.ServiceRedis, cfg.Redis.Port)
	a.emitServiceUpdate()
	return a.cfg.Save()
}

// ─── Database / Adminer APIs ─────────────────────────────────────────────────

// GetAdminerStatus trả về trạng thái của Adminer server và tool paths
func (a *App) GetAdminerStatus() map[string]any {
	return map[string]any{
		"running":       a.adminerSrv.IsRunning(),
		"url":           a.adminerSrv.URL(),
		"adminer_found": a.adminerSrv.AdminerFound(),
		"adminer_path":  a.adminerSrv.AdminerPath(),
		"php_found":     a.adminerSrv.PHPFound(),
		"php_path":      a.adminerSrv.PHPPath(),
		"heidisql_path": database.FindHeidiSQL(),
	}
}

// StartAdminer khởi động PHP server chạy Adminer, trả về URL
func (a *App) StartAdminer() (string, error) {
	url, err := a.adminerSrv.Start()
	if err != nil {
		return "", err
	}
	runtime.BrowserOpenURL(a.ctx, url)
	return url, nil
}

// StopAdminer dừng PHP server Adminer
func (a *App) StopAdminer() {
	a.adminerSrv.Stop()
}

// OpenHeidiSQL mở HeidiSQL native client
func (a *App) OpenHeidiSQL() error {
	path := database.FindHeidiSQL()
	if path == "" {
		return fmt.Errorf("HeidiSQL not found")
	}
	cmd := exec.Command(path)
	return cmd.Start()
}

// ─── SSL APIs ────────────────────────────────────────────────────────────────

// GetSSLCerts trả về danh sách domain SSL certificates đã tạo
func (a *App) GetSSLCerts() []ssl.CertInfo {
	return a.sslMgr.GetCerts()
}

// IsSSLCAInstalled kiểm tra CA có đang được tin cậy trong OS không
func (a *App) IsSSLCAInstalled() bool {
	return a.sslMgr.IsCAInstalled()
}

// TrustSSLCA cài CA vào system certificate store
func (a *App) TrustSSLCA() error {
	return a.sslMgr.TrustCA()
}

// GenerateSSLCert tạo SSL certificate cho một domain
func (a *App) GenerateSSLCert(domain string) error {
	_, _, err := a.sslMgr.GenerateCert(domain)
	return err
}

// RemoveSSLCert xóa SSL certificate của một domain
func (a *App) RemoveSSLCert(domain string) error {
	return a.sslMgr.RemoveCert(domain)
}

// GetCACertPath trả về đường dẫn CA certificate để export
func (a *App) GetCACertPath() string {
	return a.sslMgr.CACertPath()
}

// ─── Config Editor APIs ───────────────────────────────────────────────────────

// GetServiceConfigs trả về danh sách config files của một service
func (a *App) GetServiceConfigs(service string) []configeditor.ConfigFile {
	return a.cfgEditor.GetConfigFiles(service)
}

// ReadConfigFile đọc nội dung một config file
func (a *App) ReadConfigFile(path string) (string, error) {
	return a.cfgEditor.ReadFile(path)
}

// SaveConfigFile lưu nội dung config file (tự động backup)
func (a *App) SaveConfigFile(path, content string) error {
	return a.cfgEditor.SaveFile(path, content)
}

// GetConfigBackups trả về danh sách backups của một config file
func (a *App) GetConfigBackups(path string) []configeditor.BackupInfo {
	return a.cfgEditor.GetBackups(path)
}

// RestoreConfigBackup khôi phục config file từ một backup
func (a *App) RestoreConfigBackup(backupPath, targetPath string) error {
	return a.cfgEditor.RestoreBackup(backupPath, targetPath)
}

// ─── Log APIs ────────────────────────────────────────────────────────────────

// GetLogPaths trả về danh sách đường dẫn log của tất cả services
func (a *App) GetLogPaths() map[string][]string {
	return logs.LogPaths(a.cfg.LogPath)
}

// GetRecentLogs đọc N dòng cuối của log một service
func (a *App) GetRecentLogs(service string, lines int) []logs.LogEntry {
	paths := logs.LogPaths(a.cfg.LogPath)
	svcPaths, ok := paths[service]
	if !ok {
		return nil
	}

	var all []logs.LogEntry
	for _, p := range svcPaths {
		entries, err := logs.ReadLastLines(p, lines)
		if err != nil {
			continue
		}
		all = append(all, entries...)
	}
	return all
}

// StartLogWatch bắt đầu theo dõi log realtime của một service
// Mỗi dòng log mới sẽ được emit qua event "log:line"
func (a *App) StartLogWatch(service string) {
	// Hủy watcher cũ nếu có
	if a.logCancel != nil {
		a.logCancel()
	}

	watchCtx, cancel := context.WithCancel(a.ctx)
	a.logCancel = cancel

	paths := logs.LogPaths(a.cfg.LogPath)
	svcPaths, ok := paths[service]
	if !ok {
		return
	}

	out := make(chan logs.LogEntry, 50)

	// Watch từng file log
	for _, p := range svcPaths {
		if err := logs.Watch(watchCtx, p, out); err != nil {
			runtime.LogWarningf(a.ctx, "Cannot watch log %s: %v", p, err)
		}
	}

	// Forward log entries → frontend events
	go func() {
		for {
			select {
			case <-watchCtx.Done():
				return
			case entry := <-out:
				runtime.EventsEmit(a.ctx, "log:line", entry)
			}
		}
	}()
}

// StopLogWatch dừng theo dõi log
func (a *App) StopLogWatch() {
	if a.logCancel != nil {
		a.logCancel()
		a.logCancel = nil
	}
}

// ─── Terminal APIs ────────────────────────────────────────────────────────────

// TerminalStart tạo session terminal mới, bắt đầu stream output qua event "term:output"
func (a *App) TerminalStart(cwd string) error {
	a.termMu.Lock()
	// Đóng session cũ nếu có
	if a.termSession != nil {
		a.termSession.Close()
		a.termSession = nil
	}

	sess, out, err := terminal.New(a.ctx, cwd)
	if err != nil {
		a.termMu.Unlock()
		return fmt.Errorf("cannot start terminal: %w", err)
	}
	a.termSession = sess
	a.termMu.Unlock()

	// Stream output → frontend
	go func() {
		for data := range out {
			runtime.EventsEmit(a.ctx, "term:output", string(data))
		}
		// Process exited
		runtime.EventsEmit(a.ctx, "term:exit", nil)
		a.termMu.Lock()
		a.termSession = nil
		a.termMu.Unlock()
	}()

	return nil
}

// TerminalWrite gửi input từ user vào shell
func (a *App) TerminalWrite(data string) error {
	a.termMu.Lock()
	sess := a.termSession
	a.termMu.Unlock()
	if sess == nil {
		return fmt.Errorf("no active terminal session")
	}
	return sess.Write([]byte(data))
}

// TerminalResize thay đổi kích thước terminal
func (a *App) TerminalResize(rows, cols uint16) error {
	a.termMu.Lock()
	sess := a.termSession
	a.termMu.Unlock()
	if sess == nil {
		return nil
	}
	return sess.Resize(rows, cols)
}

// TerminalClose đóng session terminal
func (a *App) TerminalClose() {
	a.termMu.Lock()
	defer a.termMu.Unlock()
	if a.termSession != nil {
		a.termSession.Close()
		a.termSession = nil
	}
}

// ─── Window APIs ─────────────────────────────────────────────────────────────

// HideWindow ẩn cửa sổ xuống tray
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// ShowWindow hiện lại cửa sổ
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
}

// ─── System APIs ─────────────────────────────────────────────────────────────

// OpenFolder mở folder trong file explorer
func (a *App) OpenFolder(path string) {
	runtime.BrowserOpenURL(a.ctx, "file://"+path)
}

// SelectFolder mở dialog chọn folder
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder",
	})
}

// GetWWWPath trả về đường dẫn thư mục www
func (a *App) GetWWWPath() string {
	return a.cfg.WWWPath
}

// ShowNotification hiển thị system notification
func (a *App) ShowNotification(title, message string) {
	runtime.LogInfo(a.ctx, fmt.Sprintf("[Notification] %s: %s", title, message))
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// ensureMySQLInit chạy "mysqld --initialize-insecure" nếu data dir chưa có ibdata1.
// Hàm này block cho đến khi init xong (hoặc không cần init).
func (a *App) ensureMySQLInit() {
	avs := downloader.LoadActiveVersions(a.cfg.BinPath)
	dataDir := a.cfg.MySQLDataDir(avs["mysql"])
	ibdata := filepath.Join(dataDir, "ibdata1")
	if _, err := os.Stat(ibdata); err == nil {
		return // đã init rồi
	}

	// Tìm mysqld binary
	mysqlBinDir := downloader.ActiveExeDir(a.cfg.BinPath, "mysql")
	if mysqlBinDir == "" {
		return // chưa cài MySQL binary
	}
	mysqldPath := filepath.Join(mysqlBinDir, "mysqld.exe")
	if _, err := os.Stat(mysqldPath); err != nil {
		mysqldPath = filepath.Join(mysqlBinDir, "mysqld") // Linux/macOS
		if _, err := os.Stat(mysqldPath); err != nil {
			return // không tìm thấy mysqld
		}
	}

	os.MkdirAll(dataDir, 0755)

	runtime.LogInfof(a.ctx, "Initializing MySQL data dir: %s", dataDir)
	cmd := exec.Command(mysqldPath,
		"--initialize-insecure",
		"--datadir="+dataDir,
		"--console",
	)
	// Chạy trong thư mục cha của mysqld để mysqld tìm được share/
	cmd.Dir = filepath.Dir(filepath.Dir(mysqlBinDir)) // bin/../ = mysql version root
	out, err := cmd.CombinedOutput()
	if err != nil {
		runtime.LogErrorf(a.ctx, "MySQL init failed: %v\n%s", err, out)
	} else {
		runtime.LogInfof(a.ctx, "MySQL init complete")
	}
}

// ensureNginxConfig ghi nginx.conf trỏ vào www/ cho phiên bản nginx active.
// force=true: ghi đè kể cả khi file là stock conf từ ZIP (dùng ngay sau download).
// force=false: chỉ ghi khi file chưa có hoặc đã có marker "# Stacknest managed" (dùng khi startup/switch).
func (a *App) ensureNginxConfig(force bool) {
	nginxExeDir := downloader.ActiveExeDir(a.cfg.BinPath, "nginx")
	if nginxExeDir == "" {
		return
	}

	confDir := filepath.Join(nginxExeDir, "conf")
	confPath := filepath.Join(confDir, "nginx.conf")

	// Log tập trung vào cfg.LogPath/nginx/ (tạo sẵn cả thư mục logs/ trong prefix để nginx không lỗi)
	nginxLogsDir := filepath.Join(a.cfg.LogPath, "nginx")
	os.MkdirAll(confDir, 0755)
	os.MkdirAll(nginxLogsDir, 0755)
	os.MkdirAll(filepath.Join(nginxExeDir, "logs"), 0755) // nginx cần thư mục này tồn tại kể cả khi không ghi vào

	if !force {
		// Chỉ ghi khi: file chưa có, HOẶC file đã có marker của Stacknest.
		// Nếu file tồn tại mà không có marker → user tự quản lý → bỏ qua.
		if existing, err := os.ReadFile(confPath); err == nil && len(existing) > 0 {
			const marker = "# Stacknest managed\n"
			if len(existing) < len(marker) || string(existing[:len(marker)]) != marker {
				return
			}
		}
	}

	wwwPath := filepath.ToSlash(a.cfg.WWWPath)
	accessLog := filepath.ToSlash(filepath.Join(a.cfg.LogPath, "nginx", "access.log"))
	errorLog := filepath.ToSlash(filepath.Join(a.cfg.LogPath, "nginx", "error.log"))
	// Vhosts Nginx: {rootPath}/vhosts/nginx/*.conf
	vhostsNginxDir := filepath.ToSlash(filepath.Join(a.cfg.RootPath, "vhosts", "nginx"))
	os.MkdirAll(filepath.Join(a.cfg.RootPath, "vhosts", "nginx"), 0755)
	port := a.cfg.Nginx.Port
	if port == 0 {
		port = 8080
	}

	conf := fmt.Sprintf(`# Stacknest managed
# File này được Stacknest tự động tạo. Xoá dòng đầu tiên nếu bạn muốn tự quản lý.
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    access_log  %s;
    error_log   %s;

    sendfile        on;
    keepalive_timeout  65;

    server {
        listen       %d;
        server_name  localhost;

        root   %s;
        index  index.php index.html index.htm;

        location / {
            try_files $uri $uri/ /index.php?$query_string;
        }

        # PHP-FPM / php-cgi FastCGI
        location ~ \.php$ {
            fastcgi_pass   127.0.0.1:9000;
            fastcgi_index  index.php;
            fastcgi_param  SCRIPT_FILENAME $document_root$fastcgi_script_name;
            include        fastcgi_params;
        }

        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   html;
        }
    }

    # Virtual hosts được quản lý bởi Stacknest
    include %s/*.conf;
}
`, accessLog, errorLog, port, wwwPath, vhostsNginxDir)

	os.WriteFile(confPath, []byte(conf), 0644) //nolint:errcheck
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// servicesWithEnabled trả về danh sách services kèm trạng thái enabled từ config.
func (a *App) servicesWithEnabled() []services.ServiceInfo {
	all := a.svcMgr.GetAll()
	enabledMap := map[services.ServiceName]bool{
		services.ServiceApache: a.cfg.Apache.Enabled,
		services.ServiceNginx:  a.cfg.Nginx.Enabled,
		services.ServiceMySQL:  a.cfg.MySQL.Enabled,
		services.ServicePHP:    a.cfg.PHP.Enabled,
		services.ServiceRedis:  a.cfg.Redis.Enabled,
	}
	for i := range all {
		all[i].Enabled = enabledMap[all[i].Name]
	}
	return all
}

func (a *App) emitServiceUpdate() {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "services:updated", a.servicesWithEnabled())
	}
}

func (a *App) watchServices() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.emitServiceUpdate()
		}
	}
}
