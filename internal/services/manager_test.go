package services

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestManager() *Manager {
	return NewManager(
		map[ServiceName]string{},
		map[ServiceName]string{},
		map[ServiceName]string{},
		map[ServiceName]int{},
	)
}

func TestExePath(t *testing.T) {
	tests := []struct {
		name   string
		binDir string
		exe    string
		want   string
	}{
		{"empty binDir returns name as-is", "", "httpd.exe", "httpd.exe"},
		{"binDir joined with name", "C:/bin", "httpd.exe", filepath.Join("C:/bin", "httpd.exe")},
		{"binDir with nested path", "/usr/local/bin", "php", filepath.Join("/usr/local/bin", "php")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := exePath(tc.binDir, tc.exe)
			if got != tc.want {
				t.Errorf("exePath(%q, %q) = %q, want %q", tc.binDir, tc.exe, got, tc.want)
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	m := newTestManager()

	expectedServices := []ServiceName{
		ServiceApache, ServiceNginx, ServiceMySQL, ServicePHP, ServiceRedis,
	}

	for _, name := range expectedServices {
		t.Run(string(name)+" exists and stopped", func(t *testing.T) {
			info, err := m.GetOne(name)
			if err != nil {
				t.Fatalf("GetOne(%q) error: %v", name, err)
			}
			if info.Status != StatusStopped {
				t.Errorf("service %q initial status = %q, want %q", name, info.Status, StatusStopped)
			}
		})
	}
}

func TestGetAll_Order(t *testing.T) {
	m := newTestManager()
	all := m.GetAll()

	want := []ServiceName{ServiceApache, ServiceMySQL, ServicePHP, ServiceRedis, ServiceNginx}

	if len(all) != len(want) {
		t.Fatalf("GetAll() returned %d services, want %d", len(all), len(want))
	}

	for i, info := range all {
		t.Run("position "+string(rune('0'+i)), func(t *testing.T) {
			if info.Name != want[i] {
				t.Errorf("GetAll()[%d].Name = %q, want %q", i, info.Name, want[i])
			}
		})
	}
}

func TestGetOne(t *testing.T) {
	m := newTestManager()

	t.Run("valid service name", func(t *testing.T) {
		info, err := m.GetOne(ServiceApache)
		if err != nil {
			t.Errorf("GetOne(apache) unexpected error: %v", err)
		}
		if info.Name != ServiceApache {
			t.Errorf("GetOne(apache).Name = %q, want %q", info.Name, ServiceApache)
		}
	})

	t.Run("invalid service name returns error", func(t *testing.T) {
		_, err := m.GetOne("nonexistent")
		if err == nil {
			t.Error("GetOne(nonexistent) expected error, got nil")
		}
	})
}

func TestIsCrashLoop(t *testing.T) {
	t.Run("not crash loop with 2 restarts", func(t *testing.T) {
		m := newTestManager()
		m.RecordRestart(ServiceApache)
		m.RecordRestart(ServiceApache)

		if m.IsCrashLoop(ServiceApache) {
			t.Error("IsCrashLoop with 2 restarts should be false")
		}
	})

	t.Run("crash loop with 4 restarts", func(t *testing.T) {
		m := newTestManager()
		for i := 0; i < 4; i++ {
			m.RecordRestart(ServiceMySQL)
		}

		if !m.IsCrashLoop(ServiceMySQL) {
			t.Error("IsCrashLoop with 4 restarts should be true")
		}
	})

	t.Run("not crash loop with 3 restarts exactly", func(t *testing.T) {
		m := newTestManager()
		for i := 0; i < 3; i++ {
			m.RecordRestart(ServiceNginx)
		}

		if m.IsCrashLoop(ServiceNginx) {
			t.Error("IsCrashLoop with exactly 3 restarts should be false (threshold is >3)")
		}
	})

	t.Run("old restarts outside 1 minute window not counted", func(t *testing.T) {
		m := newTestManager()
		// Manually inject old timestamps by recording then waiting — impractical,
		// so inject directly via the internal map
		svc := m.services[ServiceRedis]
		oldTime := time.Now().Add(-2 * time.Minute)
		for i := 0; i < 5; i++ {
			svc.restartTimes = append(svc.restartTimes, oldTime)
		}
		// No recent restarts — should not be crash loop
		if m.IsCrashLoop(ServiceRedis) {
			t.Error("IsCrashLoop with only old restarts should be false")
		}
	})

	t.Run("unknown service returns false", func(t *testing.T) {
		m := newTestManager()
		if m.IsCrashLoop("unknown") {
			t.Error("IsCrashLoop for unknown service should be false")
		}
	})
}

func TestRecordRestart(t *testing.T) {
	m := newTestManager()

	for i := 1; i <= 3; i++ {
		m.RecordRestart(ServicePHP)
		info, err := m.GetOne(ServicePHP)
		if err != nil {
			t.Fatalf("GetOne error: %v", err)
		}
		if info.RestartCount != i {
			t.Errorf("after %d RecordRestart calls, RestartCount = %d, want %d", i, info.RestartCount, i)
		}
	}
}

func TestSetAutoRecover(t *testing.T) {
	m := newTestManager()

	t.Run("enable auto recover", func(t *testing.T) {
		m.SetAutoRecover(ServiceApache, true)
		info, _ := m.GetOne(ServiceApache)
		if !info.AutoRecover {
			t.Error("AutoRecover should be true after SetAutoRecover(true)")
		}
	})

	t.Run("disable auto recover", func(t *testing.T) {
		m.SetAutoRecover(ServiceApache, false)
		info, _ := m.GetOne(ServiceApache)
		if info.AutoRecover {
			t.Error("AutoRecover should be false after SetAutoRecover(false)")
		}
	})
}

func TestUpdatePort(t *testing.T) {
	m := newTestManager()

	t.Run("port changes after UpdatePort", func(t *testing.T) {
		m.UpdatePort(ServiceMySQL, 3307)
		info, err := m.GetOne(ServiceMySQL)
		if err != nil {
			t.Fatalf("GetOne error: %v", err)
		}
		if info.Port != 3307 {
			t.Errorf("Port = %d, want 3307", info.Port)
		}
	})

	t.Run("default port before update", func(t *testing.T) {
		m2 := newTestManager()
		info, _ := m2.GetOne(ServiceMySQL)
		if info.Port != 3306 {
			t.Errorf("default MySQL port = %d, want 3306", info.Port)
		}
	})
}

func TestSetCrashLoop(t *testing.T) {
	m := newTestManager()

	t.Run("set crash loop true", func(t *testing.T) {
		m.SetCrashLoop(ServiceRedis, true)
		info, _ := m.GetOne(ServiceRedis)
		if !info.CrashLoop {
			t.Error("CrashLoop should be true after SetCrashLoop(true)")
		}
	})

	t.Run("set crash loop false", func(t *testing.T) {
		m.SetCrashLoop(ServiceRedis, false)
		info, _ := m.GetOne(ServiceRedis)
		if info.CrashLoop {
			t.Error("CrashLoop should be false after SetCrashLoop(false)")
		}
	})
}
