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

//go:embed templates/reports/2m.xlsx
//go:embed templates/reports/2s.xlsx
//go:embed templates/reports/3m.xlsx
//go:embed templates/reports/3s.xlsx
//go:embed templates/reports/4m.xlsx
//go:embed templates/reports/4s.xlsx
//go:embed templates/reports/5m.xlsx
//go:embed templates/reports/5s.xlsx
//go:embed templates/reports/6m.xlsx
//go:embed templates/reports/6s.xlsx
//go:embed templates/reports/7m.xlsx
//go:embed templates/reports/7s.xlsx
//go:embed templates/reports/8m.xlsx
//go:embed templates/reports/8s.xlsx
//go:embed templates/reports/9m.xlsx
//go:embed templates/reports/9s.xlsx
//go:embed templates/reports/10m.xlsx
//go:embed templates/reports/10s.xlsx
//go:embed templates/reports/11m.xlsx
//go:embed templates/reports/11s.xlsx
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
