package services

// ServiceName định nghĩa tên các service
type ServiceName string

const (
	ServiceApache ServiceName = "apache"
	ServiceNginx  ServiceName = "nginx"
	ServiceMySQL  ServiceName = "mysql"
	ServicePHP    ServiceName = "php"
	ServiceRedis  ServiceName = "redis"
)

// ServiceStatus trạng thái của service
type ServiceStatus string

const (
	StatusRunning  ServiceStatus = "running"
	StatusStopped  ServiceStatus = "stopped"
	StatusStarting ServiceStatus = "starting"
	StatusStopping ServiceStatus = "stopping"
	StatusError    ServiceStatus = "error"
)

// ServiceInfo thông tin đầy đủ của một service
type ServiceInfo struct {
	Name    ServiceName   `json:"name"`
	Display string        `json:"display"`
	Status  ServiceStatus `json:"status"`
	Port    int           `json:"port"`
	Version string        `json:"version"`
	PID     int           `json:"pid"`
	Error   string        `json:"error,omitempty"`
	Enabled bool          `json:"enabled"`
}

// PHPVersion thông tin về một phiên bản PHP
type PHPVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Active  bool   `json:"active"`
}

// VirtualHost thông tin virtual host
type VirtualHost struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Root   string `json:"root"`
	SSL    bool   `json:"ssl"`
	Active bool   `json:"active"`
}
