package detector

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
)

// hideWindow, findExecutable, pythonExeNames, venvScriptsDir, activateScript
// are defined in common_windows.go / common_unix.go

// FindExecutable is the exported wrapper for findExecutable.
func FindExecutable(dir string) string {
	return findExecutable(dir)
}

// GetPythonVersion runs python --version and returns the version string.
func GetPythonVersion(executable string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "--version")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	s = strings.TrimPrefix(s, "Python ")
	return s
}

// ExtractMajorMinor returns "3.13" from "3.13.9".
func ExtractMajorMinor(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}

// DirSize calculates total size of a directory in bytes.
func DirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// GetArchitecture returns "64-bit" or "32-bit" based on the executable.
func GetArchitecture(executable string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-c", "import struct; print(struct.calcsize('P')*8)")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	s := strings.TrimSpace(string(out))
	if s == "64" {
		return "64-bit"
	}
	return s + "-bit"
}

// IsInPath checks if a directory is in the current PATH.
func IsInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	normDir := strings.ToLower(filepath.Clean(dir))
	for _, p := range filepath.SplitList(pathEnv) {
		if strings.ToLower(filepath.Clean(p)) == normDir {
			return true
		}
	}
	return false
}

// MakeInstallation builds a PythonInstallation from a directory.
func MakeInstallation(dir string, source models.PythonSource) *models.PythonInstallation {
	exe := findExecutable(dir)
	if exe == "" {
		return nil
	}
	version := GetPythonVersion(exe)
	if version == "" {
		return nil
	}
	return &models.PythonInstallation{
		Version:      version,
		MajorMinor:   ExtractMajorMinor(version),
		Path:         dir,
		Executable:   exe,
		Source:       source,
		SizeBytes:    DirSize(dir),
		InPath:       IsInPath(dir),
		Architecture: GetArchitecture(exe),
	}
}

// HomeDir returns the user's home directory cross-platform.
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
