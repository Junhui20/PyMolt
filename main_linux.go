//go:build linux

package main

import (
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

func applyPlatformOptions(opts *options.App) {
	opts.Linux = &linux.Options{}
}
