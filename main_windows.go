//go:build windows

package main

import (
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func applyPlatformOptions(opts *options.App) {
	opts.Windows = &windows.Options{
		WebviewIsTransparent: false,
		WindowIsTranslucent:  false,
		Theme:                windows.Light,
	}
}
