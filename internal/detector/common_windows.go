//go:build windows

package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// hideWindow prevents a console window from appearing when running a command.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// findExecutable looks for python.exe in a directory.
func findExecutable(dir string) string {
	for _, name := range []string{"python.exe", "python3.exe"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// pythonExeNames returns the executable names to look for.
func pythonExeNames() []string {
	return []string{"python.exe", "python3.exe"}
}

// venvScriptsDir returns the subdirectory name for venv executables.
func venvScriptsDir() string {
	return "Scripts"
}

// activateScript returns the activate script name for the platform.
func activateScript() string {
	return "activate.bat"
}
