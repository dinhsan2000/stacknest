package main

import (
	"embed"
	"stacknest/internal/tray"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var iconData []byte

func main() {
	app := NewApp()

	// Callbacks dùng chung cho cả systray và app menu
	cb := tray.Callbacks{
		ShowWindow: func() {
			if app.ctx != nil {
				runtime.WindowShow(app.ctx)
			}
		},
		HideWindow: func() {
			if app.ctx != nil {
				runtime.WindowHide(app.ctx)
			}
		},
		StartAll: func() { app.StartAll() },
		StopAll:  func() { app.StopAll() },
		StartSvc: func(name string) { app.StartService(name) },
		StopSvc:  func(name string) { app.StopService(name) },
		OpenWWW:  func() { app.OpenFolder(app.GetWWWPath()) },
		Quit: func() {
			if app.ctx != nil {
				runtime.Quit(app.ctx)
			}
		},
	}

	// Khởi động Windows system tray trong goroutine riêng
	go tray.Run(iconData, cb)

	// App menu cho macOS menu bar
	appMenu := tray.BuildAppMenu(cb)

	err := wails.Run(&options.App{
		Title:            "Stacknest",
		Width:            1200,
		Height:           750,
		MinWidth:         900,
		MinHeight:        600,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 20, B: 30, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
		Menu:             appMenu,
		HideWindowOnClose: true,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
