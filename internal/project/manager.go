package project

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

func httpGet(url string) (*http.Response, error) {
	return http.Get(url) //nolint:gosec
}

// Project mô tả một project phát triển
type Project struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	DocRoot   string          `json:"doc_root"`
	Domain    string          `json:"domain"`
	Server    string          `json:"server"`    // "apache" | "nginx"
	SSL       bool            `json:"ssl"`
	PHPPath   string          `json:"php_path"`  // rỗng = kế thừa active PHP
	Services  map[string]bool `json:"services"`  // {"apache": true, "mysql": true, ...}
	CreatedAt string          `json:"created_at"`
	Active    bool            `json:"active"`
}

// Manager quản lý danh sách projects
type Manager struct {
	mu       sync.RWMutex
	projects []Project
	filePath string
}

// NewManager tạo manager mới và load projects từ disk
func NewManager(rootPath string) *Manager {
	m := &Manager{
		filePath: filepath.Join(rootPath, "projects.json"),
	}
	m.load()
	return m
}

// GetAll trả về tất cả projects
func (m *Manager) GetAll() []Project {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Project, len(m.projects))
	copy(out, m.projects)
	return out
}

// Get trả về project theo ID
func (m *Manager) Get(id string) (*Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range m.projects {
		if m.projects[i].ID == id {
			p := m.projects[i]
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project %q not found", id)
}

// Create thêm project mới
func (m *Manager) Create(p Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if p.ID == "" {
		p.ID = slugify(p.Name)
	}

	// Kiểm tra trùng ID
	for _, existing := range m.projects {
		if existing.ID == p.ID {
			return fmt.Errorf("project %q already exists", p.ID)
		}
	}

	if p.CreatedAt == "" {
		p.CreatedAt = time.Now().Format(time.RFC3339)
	}
	if p.Services == nil {
		p.Services = defaultServices(p.Server)
	}

	m.projects = append(m.projects, p)
	return m.save()
}

// Update cập nhật project theo ID
func (m *Manager) Update(p Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.projects {
		if m.projects[i].ID == p.ID {
			// Giữ nguyên created_at và active
			p.CreatedAt = m.projects[i].CreatedAt
			p.Active = m.projects[i].Active
			m.projects[i] = p
			return m.save()
		}
	}
	return fmt.Errorf("project %q not found", p.ID)
}

// Delete xóa project theo ID (không xóa files trên disk)
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.projects {
		if m.projects[i].ID == id {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return m.save()
		}
	}
	return fmt.Errorf("project %q not found", id)
}

// SetActive đánh dấu project là active, clear tất cả project khác
func (m *Manager) SetActive(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	found := false
	for i := range m.projects {
		if m.projects[i].ID == id {
			m.projects[i].Active = true
			found = true
		} else {
			m.projects[i].Active = false
		}
	}
	if !found {
		return fmt.Errorf("project %q not found", id)
	}
	return m.save()
}

// ClearActive bỏ active khỏi tất cả projects
func (m *Manager) ClearActive() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.projects {
		m.projects[i].Active = false
	}
	return m.save()
}

// QuickCreate tạo project nhanh: tạo folder + domain .test
// template: "blank", "laravel", "wordpress"
func (m *Manager) QuickCreate(name, wwwPath, server, template string, ssl bool) (*Project, error) {
	id := slugify(name)
	docRoot := filepath.Join(wwwPath, id)

	// Tạo thư mục project
	if err := os.MkdirAll(docRoot, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project folder: %w", err)
	}

	// Scaffold theo template
	if err := scaffoldTemplate(template, docRoot, name); err != nil {
		return nil, fmt.Errorf("template scaffold failed: %w", err)
	}

	// Laravel docRoot trỏ vào /public
	projDocRoot := docRoot
	if template == "laravel" {
		projDocRoot = filepath.Join(docRoot, "public")
	}

	p := Project{
		ID:       id,
		Name:     name,
		DocRoot:  projDocRoot,
		Domain:   id + ".test",
		Server:   server,
		SSL:      ssl,
		Services: defaultServices(server),
	}

	if err := m.Create(p); err != nil {
		return nil, err
	}

	return &p, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (m *Manager) load() {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		m.projects = []Project{}
		return
	}
	if err := json.Unmarshal(data, &m.projects); err != nil {
		m.projects = []Project{}
	}
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "project"
	}
	return s
}

// scaffoldTemplate tạo file/folder ban đầu theo template
func scaffoldTemplate(template, docRoot, name string) error {
	switch template {
	case "laravel":
		return scaffoldLaravel(docRoot)
	case "wordpress":
		return scaffoldWordPress(docRoot)
	default: // "blank"
		return scaffoldBlank(docRoot, name)
	}
}

func scaffoldBlank(docRoot, name string) error {
	indexPath := filepath.Join(docRoot, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		html := fmt.Sprintf("<!DOCTYPE html>\n<html>\n<head><title>%s</title></head>\n<body><h1>%s</h1><p>Project created by Stacknest</p></body>\n</html>\n", name, name)
		return os.WriteFile(indexPath, []byte(html), 0644)
	}
	return nil
}

func scaffoldLaravel(docRoot string) error {
	// Chạy composer create-project vào thư mục tạm rồi move
	// composer tạo trực tiếp vào docRoot (phải rỗng hoặc chưa tồn tại)
	// Xóa folder trước vì composer cần folder rỗng
	_ = os.RemoveAll(docRoot)
	cmd := exec.Command("composer", "create-project", "laravel/laravel", docRoot, "--no-interaction")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func scaffoldWordPress(docRoot string) error {
	// Tải WordPress latest ZIP và extract
	zipURL := "https://wordpress.org/latest.zip"
	return downloadAndExtract(zipURL, docRoot, "wordpress/")
}

// downloadAndExtract tải ZIP từ URL và extract vào dest, strip prefix
func downloadAndExtract(url, dest, stripPrefix string) error {
	resp, err := httpGet(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	// Lưu vào file tạm
	tmpFile, err := os.CreateTemp("", "stacknest-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	return extractZip(tmpPath, dest, stripPrefix)
}

func extractZip(zipPath, dest, stripPrefix string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := f.Name
		if stripPrefix != "" {
			if !strings.HasPrefix(name, stripPrefix) {
				continue
			}
			name = strings.TrimPrefix(name, stripPrefix)
		}
		if name == "" {
			continue
		}

		target := filepath.Join(dest, filepath.FromSlash(name))

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func defaultServices(server string) map[string]bool {
	svcs := map[string]bool{
		"apache": false,
		"nginx":  false,
		"mysql":  true,
		"php":    true,
		"redis":  false,
	}
	if server == "nginx" {
		svcs["nginx"] = true
	} else {
		svcs["apache"] = true
	}
	return svcs
}
