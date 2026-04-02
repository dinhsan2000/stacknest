package downloader

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
// Thứ tự ưu tiên:
//  1. Downloaded binary trong binPath/{service}/{version}/
//  2. System binary (macOS: Homebrew, Linux: apt/yum) — chỉ trên non-Windows
//
// Trả về chuỗi rỗng nếu không tìm thấy binary nào.
func ActiveExeDir(binPath, service string) string {
	// Try downloaded binary first
	cat, ok := Catalog[service]
	if ok && len(cat.Versions) > 0 {
		avs := LoadActiveVersions(binPath)
		version := avs[service]
		if version == "" {
			version = cat.Versions[0].Version
		}
		for _, v := range cat.Versions {
			if v.Version == version {
				dir := exeDirFor(versionDir(binPath, service, version), v.ExeSubDir)
				exePath := filepath.Join(dir, cat.ExeName)
				if _, err := os.Stat(exePath); err == nil {
					return dir
				}
			}
		}
	}

	// Fallback: detect system-installed binary (macOS/Linux)
	return FindSystemBinary(service)
}

// GetStatus trả về trạng thái tất cả phiên bản của tất cả services theo thứ tự cố định.
// Bao gồm cả system-installed binary (macOS/Linux) nếu không có downloaded version.
func GetStatus(binPath string) []ServiceVersionStatus {
	avs := LoadActiveVersions(binPath)
	order := []string{"apache", "mysql", "postgres", "mongodb", "php", "redis", "nginx"}
	result := make([]ServiceVersionStatus, 0, len(order))
	for _, svc := range order {
		cat, ok := Catalog[svc]
		if !ok {
			continue
		}
		activeVer := avs[svc]
		versions := make([]VersionStatus, 0, len(cat.Versions)+1)

		hasInstalled := false
		for _, vspec := range cat.Versions {
			vDir := versionDir(binPath, svc, vspec.Version)
			eDir := exeDirFor(vDir, vspec.ExeSubDir)
			exePath := filepath.Join(eDir, cat.ExeName)
			installed := false
			if _, err := os.Stat(exePath); err == nil {
				installed = true
				hasInstalled = true
			}
			versions = append(versions, VersionStatus{
				Version:   vspec.Version,
				Installed: installed,
				Active:    vspec.Version == activeVer,
				ExePath:   exePath,
			})
		}

		// On macOS/Linux: show system binary if no downloaded version exists
		if !hasInstalled {
			if sysDir := FindSystemBinary(svc); sysDir != "" {
				sysVer := DetectSystemVersion(svc, sysDir)
				versions = append([]VersionStatus{{
					Version:   sysVer,
					Installed: true,
					Active:    activeVer == "" || activeVer == sysVer,
					ExePath:   sysDir,
				}}, versions...)
			}
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

	// Determine temp file extension from URL
	ext := ".zip"
	urlLower := strings.ToLower(vspec.URL)
	if strings.HasSuffix(urlLower, ".tar.gz") || strings.HasSuffix(urlLower, ".tgz") {
		ext = ".tar.gz"
	} else if strings.HasSuffix(urlLower, ".tar.xz") {
		ext = ".tar.xz"
	}

	tmp, err := os.CreateTemp("", "stacknest-*"+ext)
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

	// Extract based on archive type
	switch ext {
	case ".tar.gz":
		return extractTarGz(tmpPath, destDir, vspec.ZipStrip)
	case ".tar.xz":
		return extractTarXz(tmpPath, destDir, vspec.ZipStrip)
	default:
		return extractZip(tmpPath, destDir, vspec.ZipStrip)
	}
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

// extractTarGz extracts a .tar.gz archive, stripping prefix from paths.
func extractTarGz(archivePath, destDir, stripPrefix string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	return extractTar(tar.NewReader(gz), destDir, stripPrefix)
}

// extractTarXz extracts a .tar.xz archive via external xz command.
// Go stdlib lacks xz support, so we shell out to xz/unxz.
func extractTarXz(archivePath, destDir, stripPrefix string) error {
	// Decompress xz to tar first
	tarPath := strings.TrimSuffix(archivePath, ".xz")
	if tarPath == archivePath {
		tarPath = archivePath + ".tar"
	}
	defer os.Remove(tarPath)

	// Use xz -dk (decompress, keep original)
	cmd := execCommand("xz", "-dk", archivePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xz decompress failed: %s: %w", string(out), err)
	}

	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("open tar: %w", err)
	}
	defer f.Close()

	return extractTar(tar.NewReader(f), destDir, stripPrefix)
}

// extractTar reads entries from a tar reader and writes them to destDir.
func extractTar(tr *tar.Reader, destDir, stripPrefix string) error {
	cleanDest := filepath.Clean(destDir) + string(os.PathSeparator)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		relPath := filepath.ToSlash(header.Name)
		if stripPrefix != "" {
			if !strings.HasPrefix(relPath, stripPrefix) {
				continue
			}
			relPath = strings.TrimPrefix(relPath, stripPrefix)
		}
		if relPath == "" || relPath == "." {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(relPath))

		// Prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanDest) &&
			filepath.Clean(target) != filepath.Clean(destDir) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755) //nolint:errcheck
		case tar.TypeReg:
			if err := writeTarEntry(tr, target, header.FileInfo().Mode()); err != nil {
				return fmt.Errorf("extract %s: %w", relPath, err)
			}
		case tar.TypeSymlink:
			os.Remove(target) //nolint:errcheck
			os.Symlink(header.Linkname, target) //nolint:errcheck
		}
	}
	return nil
}

func writeTarEntry(r io.Reader, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	// Preserve executable bit
	if mode == 0 {
		mode = 0644
	}
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, r)
	return err
}

// execCommand wraps exec.Command — allows testing to stub it out.
var execCommand = exec.Command
