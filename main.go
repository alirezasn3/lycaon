package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed wails.json
var wailsConfigBytes []byte

type WailsConfig struct {
	Info WailsConfigInfo `json:"info"`
}

type WailsConfigInfo struct {
	ProductVersion string `json:"productVersion"`
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Lycaon",
		Width:     720,
		Height:    720,
		MinWidth:  720,
		MinHeight: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
