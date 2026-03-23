# Stacknest Roadmap

Lộ trình phát triển tính năng — cập nhật 2026-03-21.

---

## Giai đoạn 1 — Nâng trải nghiệm cốt lõi

### 1.1 Project Manager
- Tạo/chọn project (mỗi project = bộ config riêng: service nào bật, vhost nào, PHP version nào)
- Click chuyển project → auto-switch tất cả config + restart services
- Quick Create: nhập tên → tạo folder WWW, thêm vhost, generate SSL, mở browser
- Hỗ trợ template: Laravel, WordPress, static HTML

### 1.2 Database Backup/Restore UI
- Một nút backup MySQL database (mysqldump → .sql.gz)
- Danh sách backups với restore, download, xóa
- Lên lịch auto-backup (daily/weekly)
- Hiển thị dung lượng từng backup

### 1.3 Service Auto-Recovery
- Phát hiện service crash → tự restart (configurable)
- Đếm restart count, nếu crash liên tục (>3 lần/phút) thì dừng + thông báo
- Hiển thị uptime trên Dashboard
- Log crash events riêng

---

## Giai đoạn 2 — Tiện ích

### 2.1 Desktop Notifications
- Thông báo OS-level khi: service crash, download xong, port conflict
- Notification preferences trong Settings (bật/tắt từng loại)
- Tích hợp system tray: click notification → mở app đúng page

### 2.2 Redis GUI
- Tab xem danh sách keys (scan, filter by pattern)
- Get/Set/Delete key với value preview (string, list, hash, set)
- Monitor commands real-time
- Flush database với confirm

### 2.3 Logs nâng cao
- Export log ra file (CSV, JSON, plain text)
- Full-text search across tất cả service logs cùng lúc
- Log retention: auto-xóa log cũ hơn N ngày
- Bookmark/pin log entries quan trọng

### 2.4 Service Startup Order
- Drag-drop sắp xếp thứ tự start services
- Dependency graph: Nginx cần PHP, Adminer cần MySQL + PHP
- Start All tôn trọng thứ tự + dependencies
- Delay configurable giữa các service

---

## Giai đoạn 3 — Nâng cao

### 3.1 Performance Dashboard
- CPU/RAM usage per service (poll từ OS process info)
- Request count + error rate cho Apache/Nginx (parse access/error log)
- MySQL slow query highlight (parse slow query log)
- Biểu đồ mini (sparkline) trên Dashboard

### 3.2 Multi-site Reverse Proxy
- Nginx reverse proxy config UI cho Node.js/Go/Python apps
- Auto-detect port của app đang chạy
- Load balancing giữa nhiều backend instances

### 3.3 Docker Integration (tùy chọn)
- Cho user chọn chạy service bằng Docker thay vì binary local
- Docker Compose generate từ project config
- Container status hiện trên Dashboard cùng native services

### 3.4 Plugin System
- Cho phép community tạo plugin: thêm service mới, thêm tool
- Plugin marketplace trong app
- API cho plugin: register service, add page, add menu item

---

## Giai đoạn 4 — Polish & Quality

### 4.1 Onboarding
- First-run wizard: chọn root path, download binaries cần thiết
- Interactive tour cho user mới
- Troubleshooting guide khi service không start

### 4.2 Testing
- Unit tests cho Go packages (services, downloader, vhost, ssl)
- Integration tests cho IPC flow
- E2E tests cho critical paths (start service, create vhost, switch PHP)

### 4.3 Cross-platform
- Test và fix trên macOS (Homebrew paths, launchctl)
- Test và fix trên Linux (systemd, apt/pacman paths)
- CI build matrix: Windows + macOS + Linux

### 4.4 Config Migration
- Versioned config schema
- Auto-migrate khi user update app
- Backup config cũ trước khi migrate

---

## Đã hoàn thành

- [x] Service management (start/stop/restart/enable 5 services)
- [x] Binary version download + switch (Apache, Nginx, MySQL, PHP, Redis)
- [x] Virtual Hosts + SSL certificates (CA trust, per-domain certs)
- [x] Config Editor với CodeMirror + auto-backup
- [x] PHP version switcher (scan + custom paths)
- [x] Real-time log viewer (fsnotify, level filter, text search)
- [x] PTY terminal (xterm.js + go-pty)
- [x] Bundled Adminer (Go embed, auto-start MySQL)
- [x] Port conflict detection + kill process
- [x] i18n (English + Vietnamese, 150+ keys)
- [x] Lucide-react icons toàn bộ UI
- [x] Dashboard version switcher + port quick-edit
- [x] Reload catalog.json không cần restart app
- [x] System tray minimize
