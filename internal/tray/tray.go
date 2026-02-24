package tray

import (
	"fmt"
	"stacknest/internal/services"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
)

// Callbacks nhận từ App để tray có thể gọi lại
type Callbacks struct {
	ShowWindow func()
	HideWindow func()
	StartAll   func()
	StopAll    func()
	StartSvc   func(name string)
	StopSvc    func(name string)
	OpenWWW    func()
	Quit       func()
}

// ─── Windows System Tray (energye/systray) ────────────────────────────────────

// Run khởi động system tray icon (gọi trong goroutine riêng)
func Run(iconData []byte, cb Callbacks) {
	systray.Run(func() {
		onReady(iconData, cb)
	}, func() {
		// onExit - không cần làm gì
	})
}

func onReady(iconData []byte, cb Callbacks) {
	systray.SetIcon(iconData)
	systray.SetTitle("Stacknest")
	systray.SetTooltip("Stacknest — Dev Environment Manager")

	// Show / Hide
	mShow := systray.AddMenuItem("Show Window", "Open Stacknest")
	mHide := systray.AddMenuItem("Hide Window", "Minimize to tray")
	systray.AddSeparator()

	// Services
	mStartAll := systray.AddMenuItem("Start All", "Start all services")
	mStopAll := systray.AddMenuItem("Stop All", "Stop all services")
	systray.AddSeparator()

	// Per-service submenus
	type svcItem struct {
		name  string
		start *systray.MenuItem
		stop  *systray.MenuItem
	}
	var svcItems []svcItem
	for _, svc := range []string{"apache", "nginx", "mysql", "php", "redis"} {
		label := map[string]string{
			"apache": "Apache", "nginx": "Nginx",
			"mysql": "MySQL", "php": "PHP", "redis": "Redis",
		}[svc]
		sub := systray.AddMenuItem(label, "")
		start := sub.AddSubMenuItem("Start", "Start "+label)
		stop := sub.AddSubMenuItem("Stop", "Stop "+label)
		svcItems = append(svcItems, svcItem{svc, start, stop})
	}
	systray.AddSeparator()

	// Tools
	mOpenWWW := systray.AddMenuItem("Open WWW Folder", "")
	systray.AddSeparator()

	// Quit
	mQuit := systray.AddMenuItem("Quit Stacknest", "Exit the application")

	// Gán callback click cho từng menu item
	mShow.Click(func() { cb.ShowWindow() })
	mHide.Click(func() { cb.HideWindow() })
	mStartAll.Click(func() { cb.StartAll() })
	mStopAll.Click(func() { cb.StopAll() })
	mOpenWWW.Click(func() { cb.OpenWWW() })
	mQuit.Click(func() {
		systray.Quit()
		cb.Quit()
	})

	// Callback cho từng service
	for _, item := range svcItems {
		item := item // capture loop variable
		item.start.Click(func() { cb.StartSvc(item.name) })
		item.stop.Click(func() { cb.StopSvc(item.name) })
	}
}

// UpdateTooltip cập nhật tooltip tray theo trạng thái services
func UpdateTooltip(svcs []services.ServiceInfo) {
	running := 0
	for _, s := range svcs {
		if s.Status == services.StatusRunning {
			running++
		}
	}
	if running == 0 {
		systray.SetTooltip("Stacknest — All stopped")
	} else {
		systray.SetTooltip(fmt.Sprintf("Stacknest — %d service(s) running", running))
	}
}

// ─── Wails App Menu (macOS menu bar) ─────────────────────────────────────────

// BuildAppMenu xây dựng application menu cho macOS menu bar
func BuildAppMenu(cb Callbacks) *menu.Menu {
	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Show Window", keys.CmdOrCtrl("w"), func(_ *menu.CallbackData) { cb.ShowWindow() })
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit Stacknest", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) { cb.Quit() })

	svcMenu := appMenu.AddSubmenu("Services")
	svcMenu.AddText("Start All", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) { cb.StartAll() })
	svcMenu.AddText("Stop All", keys.CmdOrCtrl("x"), func(_ *menu.CallbackData) { cb.StopAll() })
	svcMenu.AddSeparator()
	for _, svc := range []services.ServiceName{
		services.ServiceApache, services.ServiceNginx,
		services.ServiceMySQL, services.ServicePHP, services.ServiceRedis,
	} {
		name := string(svc)
		sub := svcMenu.AddSubmenu(string(svc))
		sub.AddText("Start", nil, func(_ *menu.CallbackData) { cb.StartSvc(name) })
		sub.AddText("Stop", nil, func(_ *menu.CallbackData) { cb.StopSvc(name) })
	}

	toolsMenu := appMenu.AddSubmenu("Tools")
	toolsMenu.AddText("Open WWW Folder", nil, func(_ *menu.CallbackData) { cb.OpenWWW() })

	return appMenu
}
