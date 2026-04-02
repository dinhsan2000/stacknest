package downloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// ── Catalog file types (internal, for catalog.json) ──────────────────────────

// PlatformSpec mô tả cách tải binary cho một hệ điều hành cụ thể
type PlatformSpec struct {
	ExeName   string `json:"exe_name"`
	URL       string `json:"url"`
	ZipStrip  string `json:"zip_strip"`   // prefix bị strip khỏi ZIP entries
	ExeSubDir string `json:"exe_sub_dir"` // thư mục con chứa exe (rỗng = root)
}

// VersionEntry mô tả một phiên bản với spec riêng cho từng platform
type VersionEntry struct {
	Version   string                  `json:"version"`
	Platforms map[string]PlatformSpec `json:"platforms"` // "windows" | "darwin" | "linux"
}

// ServiceEntry danh sách các phiên bản của một service trong catalog file
type ServiceEntry struct {
	Versions []VersionEntry `json:"versions"`
}

// CatalogFile là cấu trúc của file catalog.json — có thể người dùng tự chỉnh sửa
type CatalogFile map[string]ServiceEntry

// ── Default catalog ───────────────────────────────────────────────────────────

// defaultCatalog là nội dung mặc định được ghi vào catalog.json khi chưa có file
// ghRelease is the base URL for Stacknest prebuilt service binaries.
const ghRelease = "https://github.com/dinhsan2000/stacknest/releases/download/binaries-v1"

var defaultCatalog = CatalogFile{
	"apache": {Versions: []VersionEntry{
		{Version: "2.4.63", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "httpd.exe",
				URL:      "https://www.apachelounge.com/download/VS17/binaries/httpd-2.4.63-win64-VS17.zip",
				ZipStrip: "Apache24/", ExeSubDir: "bin",
			},
			"darwin": {
				ExeName:  "httpd",
				URL:      ghRelease + "/apache-2.4.63-macos-arm64.tar.gz",
				ZipStrip: "apache/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "httpd",
				URL:      ghRelease + "/apache-2.4.63-linux-x64.tar.gz",
				ZipStrip: "apache/", ExeSubDir: "bin",
			},
		}},
	}},
	"nginx": {Versions: []VersionEntry{
		{Version: "1.26.3", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "nginx.exe",
				URL:      "https://nginx.org/download/nginx-1.26.3.zip",
				ZipStrip: "nginx-1.26.3/", ExeSubDir: "",
			},
			"darwin": {
				ExeName:  "nginx",
				URL:      ghRelease + "/nginx-1.26.3-macos-arm64.tar.gz",
				ZipStrip: "nginx/", ExeSubDir: "sbin",
			},
			"linux": {
				ExeName:  "nginx",
				URL:      ghRelease + "/nginx-1.26.3-linux-x64.tar.gz",
				ZipStrip: "nginx/", ExeSubDir: "sbin",
			},
		}},
		{Version: "1.25.5", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "nginx.exe",
				URL:      "https://nginx.org/download/nginx-1.25.5.zip",
				ZipStrip: "nginx-1.25.5/", ExeSubDir: "",
			},
		}},
	}},
	"mysql": {Versions: []VersionEntry{
		{Version: "8.0.41", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "mysqld.exe",
				URL:      "https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.41-winx64.zip",
				ZipStrip: "mysql-8.0.41-winx64/", ExeSubDir: "bin",
			},
			"darwin": {
				ExeName:  "mysqld",
				URL:      "https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-8.0.41-macos14-arm64.tar.gz",
				ZipStrip: "mysql-8.0.41-macos14-arm64/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "mysqld",
				URL:      "https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-8.0.41-linux-glibc2.17-x86_64.tar.xz",
				ZipStrip: "mysql-8.0.41-linux-glibc2.17-x86_64/", ExeSubDir: "bin",
			},
		}},
		{Version: "5.7.44", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "mysqld.exe",
				URL:      "https://dev.mysql.com/get/Downloads/MySQL-5.7/mysql-5.7.44-winx64.zip",
				ZipStrip: "mysql-5.7.44-winx64/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "mysqld",
				URL:      "https://cdn.mysql.com/Downloads/MySQL-5.7/mysql-5.7.44-linux-glibc2.12-x86_64.tar.gz",
				ZipStrip: "mysql-5.7.44-linux-glibc2.12-x86_64/", ExeSubDir: "bin",
			},
		}},
	}},
	"php": {Versions: []VersionEntry{
		{Version: "8.3.16", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "php-cgi.exe",
				URL:      "https://windows.php.net/downloads/releases/php-8.3.16-Win32-vs16-x64.zip",
				ZipStrip: "", ExeSubDir: "",
			},
			"darwin": {
				ExeName:  "php-fpm",
				URL:      ghRelease + "/php-8.3.16-macos-arm64.tar.gz",
				ZipStrip: "php/", ExeSubDir: "sbin",
			},
			"linux": {
				ExeName:  "php-fpm",
				URL:      ghRelease + "/php-8.3.16-linux-x64.tar.gz",
				ZipStrip: "php/", ExeSubDir: "sbin",
			},
		}},
		{Version: "8.2.28", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "php-cgi.exe",
				URL:      "https://windows.php.net/downloads/releases/php-8.2.28-Win32-vs16-x64.zip",
				ZipStrip: "", ExeSubDir: "",
			},
			"darwin": {
				ExeName:  "php-fpm",
				URL:      ghRelease + "/php-8.2.28-macos-arm64.tar.gz",
				ZipStrip: "php/", ExeSubDir: "sbin",
			},
			"linux": {
				ExeName:  "php-fpm",
				URL:      ghRelease + "/php-8.2.28-linux-x64.tar.gz",
				ZipStrip: "php/", ExeSubDir: "sbin",
			},
		}},
		{Version: "8.1.31", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "php-cgi.exe",
				URL:      "https://windows.php.net/downloads/releases/php-8.1.31-Win32-vs16-x64.zip",
				ZipStrip: "", ExeSubDir: "",
			},
		}},
		{Version: "7.4.33", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "php-cgi.exe",
				URL:      "https://windows.php.net/downloads/releases/archives/php-7.4.33-Win32-vc15-x64.zip",
				ZipStrip: "", ExeSubDir: "",
			},
		}},
	}},
	"postgres": {Versions: []VersionEntry{
		{Version: "17.4", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "postgres.exe",
				URL:      "https://get.enterprisedb.com/postgresql/postgresql-17.4-1-windows-x64-binaries.zip",
				ZipStrip: "pgsql/", ExeSubDir: "bin",
			},
			"darwin": {
				ExeName:  "postgres",
				URL:      "https://get.enterprisedb.com/postgresql/postgresql-17.4-1-osx-binaries.zip",
				ZipStrip: "pgsql/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "postgres",
				URL:      "https://get.enterprisedb.com/postgresql/postgresql-17.4-1-linux-x64-binaries.tar.gz",
				ZipStrip: "pgsql/", ExeSubDir: "bin",
			},
		}},
		{Version: "16.8", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "postgres.exe",
				URL:      "https://get.enterprisedb.com/postgresql/postgresql-16.8-1-windows-x64-binaries.zip",
				ZipStrip: "pgsql/", ExeSubDir: "bin",
			},
			"darwin": {
				ExeName:  "postgres",
				URL:      "https://get.enterprisedb.com/postgresql/postgresql-16.8-1-osx-binaries.zip",
				ZipStrip: "pgsql/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "postgres",
				URL:      "https://get.enterprisedb.com/postgresql/postgresql-16.8-1-linux-x64-binaries.tar.gz",
				ZipStrip: "pgsql/", ExeSubDir: "bin",
			},
		}},
	}},
	"mongodb": {Versions: []VersionEntry{
		{Version: "8.0.4", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "mongod.exe",
				URL:      "https://fastdl.mongodb.org/windows/mongodb-windows-x86_64-8.0.4.zip",
				ZipStrip: "mongodb-win32-x86_64-windows-8.0.4/", ExeSubDir: "bin",
			},
			"darwin": {
				ExeName:  "mongod",
				URL:      "https://fastdl.mongodb.org/osx/mongodb-macos-arm64-8.0.4.tgz",
				ZipStrip: "mongodb-macos-aarch64-8.0.4/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "mongod",
				URL:      "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu2204-8.0.4.tgz",
				ZipStrip: "mongodb-linux-x86_64-ubuntu2204-8.0.4/", ExeSubDir: "bin",
			},
		}},
		{Version: "7.0.16", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "mongod.exe",
				URL:      "https://fastdl.mongodb.org/windows/mongodb-windows-x86_64-7.0.16.zip",
				ZipStrip: "mongodb-win32-x86_64-windows-7.0.16/", ExeSubDir: "bin",
			},
			"darwin": {
				ExeName:  "mongod",
				URL:      "https://fastdl.mongodb.org/osx/mongodb-macos-arm64-7.0.16.tgz",
				ZipStrip: "mongodb-macos-aarch64-7.0.16/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "mongod",
				URL:      "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu2204-7.0.16.tgz",
				ZipStrip: "mongodb-linux-x86_64-ubuntu2204-7.0.16/", ExeSubDir: "bin",
			},
		}},
	}},
	"redis": {Versions: []VersionEntry{
		{Version: "7.2.7", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "redis-server.exe",
				URL:      "https://github.com/redis-windows/redis-windows/releases/download/7.2.7/Redis-7.2.7-Windows-x64-cygwin.zip",
				ZipStrip: "", ExeSubDir: "",
			},
			"darwin": {
				ExeName:  "redis-server",
				URL:      ghRelease + "/redis-7.2.7-macos-arm64.tar.gz",
				ZipStrip: "redis/", ExeSubDir: "bin",
			},
			"linux": {
				ExeName:  "redis-server",
				URL:      ghRelease + "/redis-7.2.7-linux-x64.tar.gz",
				ZipStrip: "redis/", ExeSubDir: "bin",
			},
		}},
		{Version: "5.0.14", Platforms: map[string]PlatformSpec{
			"windows": {
				ExeName:  "redis-server.exe",
				URL:      "https://github.com/tporadowski/redis/releases/download/v5.0.14.1/Redis-x64-5.0.14.1.zip",
				ZipStrip: "", ExeSubDir: "",
			},
		}},
	}},
}

// ── Catalog initialization ────────────────────────────────────────────────────

func catalogFilePath(rootPath string) string {
	return filepath.Join(rootPath, "catalog.json")
}

// InitCatalog đọc catalog.json từ rootPath, tạo file mặc định nếu chưa có,
// sau đó resolve về Catalog map cho platform hiện tại và cập nhật biến global Catalog.
func InitCatalog(rootPath string) {
	path := catalogFilePath(rootPath)

	// Tạo catalog.json mặc định nếu chưa có
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if data, err := json.MarshalIndent(defaultCatalog, "", "  "); err == nil {
			os.MkdirAll(rootPath, 0755)
			os.WriteFile(path, data, 0644) //nolint:errcheck
		}
	}

	// Đọc file (có thể đã được người dùng chỉnh sửa)
	data, err := os.ReadFile(path)
	var cf CatalogFile
	if err != nil || json.Unmarshal(data, &cf) != nil {
		cf = defaultCatalog
	}

	Catalog = resolveCatalog(cf)
}

// resolveCatalog chuyển CatalogFile (multi-platform) sang Catalog map
// chỉ giữ lại các version có spec cho platform đang chạy.
func resolveCatalog(cf CatalogFile) map[string]ServiceCatalog {
	platform := runtime.GOOS // "windows" | "darwin" | "linux"
	result := make(map[string]ServiceCatalog, len(cf))

	for svcName, entry := range cf {
		var exeName string
		var versions []VersionSpec

		for _, ve := range entry.Versions {
			ps, ok := ve.Platforms[platform]
			if !ok {
				continue // phiên bản này không hỗ trợ platform hiện tại
			}
			if exeName == "" {
				exeName = ps.ExeName
			}
			versions = append(versions, VersionSpec{
				Version:   ve.Version,
				URL:       ps.URL,
				ZipStrip:  ps.ZipStrip,
				ExeSubDir: ps.ExeSubDir,
			})
		}

		if len(versions) > 0 {
			result[svcName] = ServiceCatalog{
				ExeName:  exeName,
				Versions: versions,
			}
		}
	}
	return result
}
