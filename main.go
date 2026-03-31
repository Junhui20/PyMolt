package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/Junhui20/PyMolt/internal"
	"github.com/Junhui20/PyMolt/internal/cli"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// CLI mode: if args provided, run CLI and exit
	if cli.Run(os.Args) {
		return
	}

	// GUI mode
	app := internal.NewApp()

	appOpts := &options.App{
		Title:     "PyMolt",
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
	}

	// Apply platform-specific options
	applyPlatformOptions(appOpts)

	err := wails.Run(appOpts)
	if err != nil {
		log.Fatal(err)
	}
}
