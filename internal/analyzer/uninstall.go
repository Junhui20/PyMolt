package analyzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
	"golang.org/x/sys/windows/registry"
)

// UninstallResult reports what happened.
type UninstallResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	SpaceFreed int64  `json:"spaceFreed"`
}

// Uninstall removes a Python installation based on its source.
func Uninstall(inst models.PythonInstallation) *UninstallResult {
	switch inst.Source {
	case models.SourceVenv:
		return deleteDirectory(inst)
	case models.SourceUV:
		return uninstallUV(inst)
	case models.SourcePyenv:
		return deleteDirectory(inst)
	case models.SourceOfficial:
		return uninstallOfficial(inst)
	case models.SourceChocolatey:
		return uninstallChocolatey(inst)
	case models.SourceScoop:
		return uninstallScoop(inst)
	default:
		return deleteDirectory(inst)
	}
}

// dangerousPrefixes lists paths that must never be deleted.
var dangerousPrefixes = []string{
	`c:\`,
	`c:\windows`,
	`c:\program files`,
	`c:\program files (x86)`,
	`c:\users`,
}

func deleteDirectory(inst models.PythonInstallation) *UninstallResult {
	clean := strings.ToLower(filepath.Clean(inst.Path))
	for _, prefix := range dangerousPrefixes {
		if clean == prefix {
			return &UninstallResult{Success: false, Message: fmt.Sprintf("Refusing to delete protected path: %s", inst.Path)}
		}
	}
	size := inst.SizeBytes
	err := os.RemoveAll(inst.Path)
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Failed to delete %s: %v", inst.Path, err)}
	}
	return &UninstallResult{Success: true, Message: fmt.Sprintf("Deleted %s", inst.Path), SpaceFreed: size}
}

func uninstallUV(inst models.PythonInstallation) *UninstallResult {
	// Extract version like "3.13" from "3.13.9"
	ver := inst.MajorMinor
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "uv", "python", "uninstall", ver)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to directory deletion
		return deleteDirectory(inst)
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: inst.SizeBytes}
}

func uninstallOfficial(inst models.PythonInstallation) *UninstallResult {
	// Try to find the uninstaller executable from registry
	uninstallerExe := findUninstaller(inst.Version)
	if uninstallerExe == "" {
		return &UninstallResult{Success: false, Message: "Uninstaller not found. Please uninstall manually from Settings > Apps."}
	}

	// Validate the uninstaller path is an actual .exe file to prevent shell injection.
	if !strings.EqualFold(filepath.Ext(uninstallerExe), ".exe") {
		return &UninstallResult{Success: false, Message: "Unexpected uninstaller format. Please uninstall manually from Settings > Apps."}
	}

	// Run the uninstaller executable directly (no shell), passing /quiet flag.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, uninstallerExe, "/quiet")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	err := cmd.Run()
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Uninstaller failed: %v. Try uninstalling from Settings > Apps.", err)}
	}
	return &UninstallResult{Success: true, Message: "Uninstalled Python " + inst.Version, SpaceFreed: inst.SizeBytes}
}

// findUninstaller returns the uninstaller executable path by parsing the
// registry UninstallString. It extracts only the .exe path to prevent shell
// injection when the value contains arguments or shell metacharacters.
func findUninstaller(version string) string {
	// Search HKCU uninstall keys
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

// parseUninstallerExe extracts only the executable path from an UninstallString.
// The string may be quoted (e.g. "C:\...\uninstall.exe" /args) or unquoted.
func parseUninstallerExe(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, `"`) {
		// Quoted path: extract up to the closing quote
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			return ""
		}
		return s[1 : end+1]
	}
	// Unquoted: the executable is the first space-delimited token
	parts := strings.SplitN(s, " ", 2)
	return parts[0]
}

func uninstallChocolatey(inst models.PythonInstallation) *UninstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "choco", "uninstall", "python"+strings.ReplaceAll(inst.MajorMinor, ".", ""), "-y")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return deleteDirectory(inst)
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: inst.SizeBytes}
}

// GetPipCacheSize returns the size of pip cache in bytes.
func GetPipCacheSize() (int64, string) {
	home := os.Getenv("LOCALAPPDATA")
	cachePath := filepath.Join(home, "pip", "cache")

	// Also check the new location
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	size, _ := GetUVCacheSize()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Failed: %v", err)}
	}
	return &UninstallResult{Success: true, Message: strings.TrimSpace(string(out)), SpaceFreed: size}
}
