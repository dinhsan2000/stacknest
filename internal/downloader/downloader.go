package downloader

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// VersionSpec mô tả cách tải và giải nén một phiên bản cụ thể của service
type VersionSpec struct {
	Version   string `json:"version"`
	URL       string `json:"url"`
	ZipStrip  string `json:"zip_strip"`   // prefix bị cắt khỏi ZIP entries
	ExeSubDir string `json:"exe_sub_dir"` // thư mục con trong destDir chứa exe (rỗng = root)
}

// ServiceCatalog danh mục các phiên bản có thể tải của một service
type ServiceCatalog struct {
	ExeName  string        `json:"exe_name"`
	Versions []VersionSpec `json:"versions"`
}

// Catalog được khởi tạo bởi InitCatalog(rootPath) khi app khởi động.
// Chứa các phiên bản đã được resolve cho platform hiện tại.
var Catalog map[string]ServiceCatalog

// ActiveVersions lưu phiên bản đang active của mỗi service: service → version string
type ActiveVersions map[string]string

// VersionStatus trạng thái của một phiên bản cụ thể
type VersionStatus struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
	ExePath   string `json:"exe_path"`
}

// ServiceVersionStatus trạng thái tất cả phiên bản của một service
type ServiceVersionStatus struct {
	Service  string          `json:"service"`
	Versions []VersionStatus `json:"versions"`
}

// ProgressFunc được gọi trong quá trình tải với số byte đã tải và tổng số byte (-1 nếu không rõ)
type ProgressFunc func(downloaded, total int64)

func versionsFilePath(binPath string) string {
	return filepath.Join(binPath, "versions.json")
}

// LoadActiveVersions đọc versions.json; trả về empty map nếu chưa có hoặc lỗi
func LoadActiveVersions(binPath string) ActiveVersions {
	data, err := os.ReadFile(versionsFilePath(binPath))
	if err != nil {
		return ActiveVersions{}
	}
	var avs ActiveVersions
	if err := json.Unmarshal(data, &avs); err != nil {
		return ActiveVersions{}
	}
	return avs
}

// SaveActiveVersions ghi versions.json
func SaveActiveVersions(binPath string, avs ActiveVersions) error {
	if err := os.MkdirAll(binPath, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(avs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionsFilePath(binPath), data, 0644)
}

// SetActiveVersion đặt phiên bản active cho một service và lưu vào versions.json
func SetActiveVersion(binPath, service, version string) error {
	avs := LoadActiveVersions(binPath)
	avs[service] = version
	return SaveActiveVersions(binPath, avs)
}

// versionDir trả về thư mục của một phiên bản cụ thể: binPath/{service}/{version}
func versionDir(binPath, service, version string) string {
	return filepath.Join(binPath, service, version)
}

// exeDirFor trả về thư mục chứa executable trong versionDir
func exeDirFor(vDir, exeSubDir string) string {
	if exeSubDir == "" {
		return vDir
	}
	return filepath.Join(vDir, exeSubDir)
}

// ActiveExeDir trả về thư mục chứa executable của phiên bản đang active.
// Nếu chưa đặt active version, fallback về phiên bản đầu tiên trong catalog.
// Trả về chuỗi rỗng nếu service không tồn tại trong catalog.
func ActiveExeDir(binPath, service string) string {
	cat, ok := Catalog[service]
	if !ok || len(cat.Versions) == 0 {
		return ""
	}
	avs := LoadActiveVersions(binPath)
	version := avs[service]
	if version == "" {
		version = cat.Versions[0].Version
	}
	for _, v := range cat.Versions {
		if v.Version == version {
			return exeDirFor(versionDir(binPath, service, version), v.ExeSubDir)
		}
	}
	return ""
}

// GetStatus trả về trạng thái tất cả phiên bản của tất cả services theo thứ tự cố định
func GetStatus(binPath string) []ServiceVersionStatus {
	avs := LoadActiveVersions(binPath)
	order := []string{"apache", "mysql", "php", "redis", "nginx"}
	result := make([]ServiceVersionStatus, 0, len(order))
	for _, svc := range order {
		cat, ok := Catalog[svc]
		if !ok {
			continue
		}
		activeVer := avs[svc]
		versions := make([]VersionStatus, 0, len(cat.Versions))
		for _, vspec := range cat.Versions {
			vDir := versionDir(binPath, svc, vspec.Version)
			eDir := exeDirFor(vDir, vspec.ExeSubDir)
			exePath := filepath.Join(eDir, cat.ExeName)
			installed := false
			if _, err := os.Stat(exePath); err == nil {
				installed = true
			}
			versions = append(versions, VersionStatus{
				Version:   vspec.Version,
				Installed: installed,
				Active:    vspec.Version == activeVer,
				ExePath:   exePath,
			})
		}
		result = append(result, ServiceVersionStatus{
			Service:  svc,
			Versions: versions,
		})
	}
	return result
}

// Download tải và giải nén binary cho service/version vào binPath/{service}/{version}/
func Download(ctx context.Context, service, version, binPath string, onProgress ProgressFunc) error {
	cat, ok := Catalog[service]
	if !ok {
		return fmt.Errorf("unknown service: %s", service)
	}
	var vspec *VersionSpec
	for i := range cat.Versions {
		if cat.Versions[i].Version == version {
			vspec = &cat.Versions[i]
			break
		}
	}
	if vspec == nil {
		return fmt.Errorf("version %s not found for service %s", version, service)
	}

	destDir := versionDir(binPath, service, version)

	tmp, err := os.CreateTemp("", "stacknest-*.zip")
	if err != nil {
		return fmt.Errorf("tạo temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := downloadToFile(ctx, vspec.URL, tmp, onProgress); err != nil {
		tmp.Close()
		return fmt.Errorf("tải %s: %w", vspec.URL, err)
	}
	tmp.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("tạo thư mục: %w", err)
	}
	return extractZip(tmpPath, destDir, vspec.ZipStrip)
}

// Delete xóa binary đã cài của service/version.
// Không cho xóa version đang active — phải switch sang version khác trước.
func Delete(service, version, binPath string) error {
	vDir := versionDir(binPath, service, version)
	if _, err := os.Stat(vDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s/%s not found", service, version)
	}

	avs := LoadActiveVersions(binPath)
	if avs[service] == version {
		return fmt.Errorf("cannot delete active version %s/%s — switch to another version first", service, version)
	}

	return os.RemoveAll(vDir)
}

func downloadToFile(ctx context.Context, url string, w io.Writer, onProgress ProgressFunc) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		// Check cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			if onProgress != nil {
				onProgress(downloaded, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}

func extractZip(zipPath, destDir, stripPrefix string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("mở zip: %w", err)
	}
	defer r.Close()

	cleanDest := filepath.Clean(destDir) + string(os.PathSeparator)

	for _, f := range r.File {
		relPath := filepath.ToSlash(f.Name)

		if stripPrefix != "" {
			if !strings.HasPrefix(relPath, stripPrefix) {
				continue
			}
			relPath = strings.TrimPrefix(relPath, stripPrefix)
		}

		if relPath == "" {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(relPath))

		// Ngăn path traversal attack
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanDest) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := writeZipEntry(f, target); err != nil {
			return fmt.Errorf("giải nén %s: %w", relPath, err)
		}
	}
	return nil
}

func writeZipEntry(f *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, rc)
	return err
}
