//go:build windows

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
	"golang.org/x/sys/windows/registry"
)

var windowsDangerousPrefixes = []string{
	`c:\`,
	`c:\windows`,
	`c:\program files`,
	`c:\program files (x86)`,
	`c:\users`,
}

func isDangerousPath(clean string) bool {
	for _, prefix := range windowsDangerousPrefixes {
		if clean == prefix {
			return true
		}
	}
	return false
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
	uninstallerExe := findUninstaller(inst.Version)
	if uninstallerExe == "" {
		return &UninstallResult{Success: false, Message: "Uninstaller not found. Please uninstall manually from Settings > Apps."}
	}

	if !strings.EqualFold(filepath.Ext(uninstallerExe), ".exe") {
		return &UninstallResult{Success: false, Message: "Unexpected uninstaller format. Please uninstall manually from Settings > Apps."}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, uninstallerExe, "/quiet")
	hideWindow(cmd)
	err := cmd.Run()
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Uninstaller failed: %v. Try uninstalling from Settings > Apps.", err)}
	}
	return &UninstallResult{Success: true, Message: "Uninstalled Python " + inst.Version, SpaceFreed: inst.SizeBytes}
}

func findUninstaller(version string) string {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Uninstall`, registry.READ)
	if err != nil {
		return ""
	}
	defer key.Close()

	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return ""
	}

	for _, name := range names {
		sub, err := registry.OpenKey(registry.CURRENT_USER,
			`Software\Microsoft\Windows\CurrentVersion\Uninstall\`+name, registry.READ)
		if err != nil {
			continue
		}
		displayName, _, _ := sub.GetStringValue("DisplayName")
		uninstall, _, _ := sub.GetStringValue("UninstallString")
		sub.Close()

		if strings.Contains(displayName, "Python "+version) && uninstall != "" {
			return parseUninstallerExe(uninstall)
		}
	}
	return ""
}

func parseUninstallerExe(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, `"`) {
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			return ""
		}
		return s[1 : end+1]
	}
	parts := strings.SplitN(s, " ", 2)
	return parts[0]
}

func uninstallChocolatey(inst models.PythonInstallation) *UninstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "choco", "uninstall", "python"+strings.ReplaceAll(inst.MajorMinor, ".", ""), "-y")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return deleteDirectory(inst)
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: inst.SizeBytes}
}

func uninstallScoop(inst models.PythonInstallation) *UninstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "scoop", "uninstall", "python")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return deleteDirectory(inst)
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: inst.SizeBytes}
}

func uninstallHomebrew(_ models.PythonInstallation) *UninstallResult {
	return &UninstallResult{Success: false, Message: "Homebrew is not available on Windows"}
}

// GetPipCacheSize returns the size of pip cache in bytes.
func GetPipCacheSize() (int64, string) {
	home := os.Getenv("LOCALAPPDATA")
	cachePath := filepath.Join(home, "pip", "cache")
	if _, err := os.Stat(cachePath); err != nil {
		cachePath = filepath.Join(home, "pip", "Cache")
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
	home := os.Getenv("LOCALAPPDATA")
	cachePath := filepath.Join(home, "uv", "cache")
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
