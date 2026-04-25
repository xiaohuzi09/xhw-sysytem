package main

import (
	"changeme/services"
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	imageService := &services.ImageService{}
	autoUploadService := &services.AutoUploadService{}

	err := wails.Run(&options.App{
		Title:  "图片模板管理器",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 245, B: 245, A: 1},
		OnStartup: func(ctx context.Context) {
			imageService.Startup(ctx)
			autoUploadService.Startup(ctx)
		},
		Bind: []interface{}{
			imageService,
			autoUploadService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
