//go:build !windows

package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// OpenTerminal opens a new terminal window with the given Python or venv activated.
func OpenTerminal(inst models.PythonInstallation) error {
	var shellCmd string

	if inst.Source == models.SourceVenv {
		activateScript := filepath.Join(inst.Path, "bin", "activate")
		if _, err := os.Stat(activateScript); err != nil {
			return fmt.Errorf("activate script not found in %s", inst.Path)
		}
		shellCmd = fmt.Sprintf("source %s && exec $SHELL", activateScript)
	} else {
		shellCmd = fmt.Sprintf("export PATH=\"%s:$PATH\" && python3 --version && exec $SHELL", inst.Path)
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" {
		// Use osascript to open Terminal.app
		script := fmt.Sprintf(`tell application "Terminal" to do script "%s"`, shellCmd)
		cmd = exec.Command("osascript", "-e", script)
	} else {
		// Linux: try common terminal emulators
		terminal := findTerminalEmulator()
		if terminal == "" {
			return fmt.Errorf("no terminal emulator found")
		}
		cmd = exec.Command(terminal, "-e", "bash", "-c", shellCmd)
	}

	return cmd.Start()
}

// AddToPATH adds a Python installation's directory to the PATH.
func AddToPATH(pythonDir string) error {
	return SetDefaultPython(pythonDir)
}

func findTerminalEmulator() string {
	for _, term := range []string{
		"gnome-terminal", "konsole", "xfce4-terminal",
		"mate-terminal", "tilix", "alacritty", "kitty",
		"xterm",
	} {
		if path, err := exec.LookPath(term); err == nil {
			return path
		}
	}
	return ""
}
