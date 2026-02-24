package portcheck

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ConflictInfo thông tin process đang chiếm port
type ConflictInfo struct {
	Port    int    `json:"port"`
	PID     int    `json:"pid"`
	Process string `json:"process"`
	InUse   bool   `json:"in_use"`
}

// Check kiểm tra port có đang bị chiếm không
// Trả về ConflictInfo với InUse=false nếu port trống
func Check(port int) ConflictInfo {
	if !isPortInUse(port) {
		return ConflictInfo{Port: port, InUse: false}
	}

	pid, procName := findProcess(port)
	return ConflictInfo{
		Port:    port,
		PID:     pid,
		Process: procName,
		InUse:   true,
	}
}

// KillProcess kill process theo PID
func KillProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	} else {
		cmd = exec.Command("kill", "-9", strconv.Itoa(pid))
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kill failed: %s", string(out))
	}
	return nil
}

// isPortInUse kiểm tra nhanh port có đang dùng không
func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}

// findProcess tìm PID và tên process đang dùng port
func findProcess(port int) (pid int, name string) {
	switch runtime.GOOS {
	case "windows":
		return findProcessWindows(port)
	default:
		return findProcessUnix(port)
	}
}

func findProcessWindows(port int) (pid int, name string) {
	// netstat -ano | findstr ":PORT "
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return 0, "unknown"
	}

	portStr := fmt.Sprintf(":%d", port)
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, portStr) {
			continue
		}
		// Format: Proto  Local Address  Foreign Address  State  PID
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		// Chỉ lấy LISTENING hoặc dòng có local address khớp
		localAddr := fields[1]
		if !strings.HasSuffix(localAddr, portStr) {
			continue
		}
		pid, err = strconv.Atoi(fields[len(fields)-1])
		if err != nil || pid <= 0 {
			continue
		}
		name = getProcessNameWindows(pid)
		return pid, name
	}
	return 0, "unknown"
}

func getProcessNameWindows(pid int) string {
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(out))
	if line == "" || strings.Contains(line, "No tasks") {
		return "unknown"
	}
	// Format CSV: "process.exe","PID","Session","#","Mem"
	parts := strings.Split(line, ",")
	if len(parts) > 0 {
		return strings.Trim(parts[0], `"`)
	}
	return "unknown"
}

func findProcessUnix(port int) (pid int, name string) {
	// lsof -i :PORT -t  → trả về PID
	out, err := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t").Output()
	if err != nil {
		// fallback: ss -tlnp
		return findProcessSS(port)
	}
	pidStr := strings.TrimSpace(string(out))
	pid, err = strconv.Atoi(strings.Split(pidStr, "\n")[0])
	if err != nil {
		return 0, "unknown"
	}
	// Lấy tên process từ /proc/<pid>/comm
	commOut, err := exec.Command("cat", fmt.Sprintf("/proc/%d/comm", pid)).Output()
	if err != nil {
		return pid, "unknown"
	}
	return pid, strings.TrimSpace(string(commOut))
}

func findProcessSS(port int) (int, string) {
	out, err := exec.Command("ss", "-tlnp", fmt.Sprintf("sport = :%d", port)).Output()
	if err != nil {
		return 0, "unknown"
	}
	// Parse pid= từ output
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, fmt.Sprintf(":%d", port)) {
			continue
		}
		if idx := strings.Index(line, "pid="); idx >= 0 {
			rest := line[idx+4:]
			end := strings.IndexAny(rest, ",)")
			if end > 0 {
				pid, _ := strconv.Atoi(rest[:end])
				return pid, "unknown"
			}
		}
	}
	return 0, "unknown"
}
