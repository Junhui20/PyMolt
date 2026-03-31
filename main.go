package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/Junhui20/PyMolt/internal"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := internal.NewApp()

	err := wails.Run(&options.App{
		Title:     "Python Manager",
		Width:     1100,
		Height:    750,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 249, B: 250, A: 1},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Light,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
