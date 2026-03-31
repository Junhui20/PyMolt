//go:build !windows

package detector

import (
	"os"
	"os/exec"
	"path/filepath"
)

// hideWindow is a no-op on Unix (no console window to hide).
func hideWindow(cmd *exec.Cmd) {
	// No-op on macOS/Linux
	_ = cmd
}

// findExecutable looks for python3 or python in a directory.
func findExecutable(dir string) string {
	for _, name := range []string{"python3", "python"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// pythonExeNames returns the executable names to look for.
func pythonExeNames() []string {
	return []string{"python3", "python"}
}

// venvScriptsDir returns the subdirectory name for venv executables.
func venvScriptsDir() string {
	return "bin"
}

// activateScript returns the activate script name for the platform.
func activateScript() string {
	return "activate"
}
