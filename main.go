package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/MJKhaani/ESP32-Marauder/internal/app"
	"github.com/MJKhaani/ESP32-Marauder/internal/serialtransport"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appInstance, err := app.NewApp(serialtransport.NewSystemPortEnumerator(), "firmware/catalog/targets.json", "marauder_studio.db")
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	err = wails.Run(&options.App{
		Title:  "Marauder Studio",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: appInstance.Startup,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
