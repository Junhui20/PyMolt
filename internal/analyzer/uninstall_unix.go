//go:build !windows

package analyzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
)

func isDangerousPath(clean string) bool {
	return dangerousPaths[clean]
}

func uninstallUV(inst models.PythonInstallation) *UninstallResult {
	ver := inst.MajorMinor
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "uv", "python", "uninstall", ver)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return deleteDirectory(inst)
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: inst.SizeBytes}
}

func uninstallOfficial(inst models.PythonInstallation) *UninstallResult {
	return &UninstallResult{
		Success: false,
		Message: "Cannot automatically uninstall system Python. Use your package manager (apt, brew, etc.) to remove it.",
	}
}

func uninstallChocolatey(_ models.PythonInstallation) *UninstallResult {
	return &UninstallResult{Success: false, Message: "Chocolatey is not available on this platform"}
}

func uninstallScoop(_ models.PythonInstallation) *UninstallResult {
	return &UninstallResult{Success: false, Message: "Scoop is not available on this platform"}
}

func uninstallHomebrew(inst models.PythonInstallation) *UninstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "brew", "uninstall", "python@"+inst.MajorMinor)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("brew uninstall failed: %s", strings.TrimSpace(string(out)))}
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: inst.SizeBytes}
}

// GetPipCacheSize returns the size of pip cache in bytes.
func GetPipCacheSize() (int64, string) {
	home := os.Getenv("HOME")
	// pip cache on Unix: ~/.cache/pip
	cachePath := filepath.Join(home, ".cache", "pip")
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		cachePath = filepath.Join(xdg, "pip")
	}

	var size int64
	_ = filepath.Walk(cachePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, cachePath
}

// CleanPipCache removes all pip cache files.
func CleanPipCache() *UninstallResult {
	size, cachePath := GetPipCacheSize()
	if size == 0 {
		return &UninstallResult{Success: true, Message: "pip cache is already empty"}
	}
	err := os.RemoveAll(cachePath)
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Failed to clean pip cache: %v", err)}
	}
	return &UninstallResult{Success: true, Message: "Cleaned pip cache", SpaceFreed: size}
}

// GetUVCacheSize returns the size of uv cache.
func GetUVCacheSize() (int64, string) {
	home := os.Getenv("HOME")
	cachePath := filepath.Join(home, ".cache", "uv")
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		cachePath = filepath.Join(xdg, "uv")
	}

	var size int64
	_ = filepath.Walk(cachePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, cachePath
}

// CleanUVCache removes all uv cache files.
func CleanUVCache() *UninstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "uv", "cache", "clean")
	hideWindow(cmd)
	size, _ := GetUVCacheSize()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Failed: %v", err)}
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: size}
}
