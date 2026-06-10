package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:web/dist
var assets embed.FS

//go:embed all:templates/reports
var templateAssets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Cal1604 校准系统",
		Width:             1660,
		Height:            1040,
		WindowStartState:  options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		log.Fatalf("wails run failed: %v", err)
	}
}
