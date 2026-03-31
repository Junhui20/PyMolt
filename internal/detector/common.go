package detector

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
)

// hideWindow sets SysProcAttr to hide console windows on Windows.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// GetPythonVersion runs python.exe --version and returns the version string.
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

// FindExecutable looks for python.exe in a directory.
func FindExecutable(dir string) string {
	for _, name := range []string{"python.exe", "python3.exe"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
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
	exe := FindExecutable(dir)
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
		Source:        source,
		SizeBytes:    DirSize(dir),
		InPath:       IsInPath(dir),
		Architecture: GetArchitecture(exe),
	}
}
