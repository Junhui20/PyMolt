//go:build windows

package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// OpenTerminal opens a new terminal window with the given Python or venv activated.
func OpenTerminal(inst models.PythonInstallation) error {
	var cmd *exec.Cmd

	if inst.Source == models.SourceVenv {
		activateScript := filepath.Join(inst.Path, "Scripts", "activate.bat")
		if _, err := os.Stat(activateScript); err != nil {
			return fmt.Errorf("activate.bat not found in %s", inst.Path)
		}
		cmd = exec.Command("cmd")
		cmd.Dir = filepath.Dir(inst.Path)
		cmd.Args = []string{"cmd", "/k", activateScript}
	} else {
		batContent := fmt.Sprintf("@echo off\r\ntitle Python %s\r\nset \"PATH=%s;%%PATH%%\"\r\npython --version\r\n",
			inst.Version, inst.Path)
		batFile := filepath.Join(os.TempDir(), "pymanager_term.bat")
		if err := os.WriteFile(batFile, []byte(batContent), 0644); err != nil {
			return fmt.Errorf("failed to write temp bat: %w", err)
		}
		cmd = exec.Command("cmd", "/k", batFile)
	}

	newConsole(cmd)
	return cmd.Start()
}

// AddToPATH adds a Python installation's directory to the User PATH.
func AddToPATH(pythonDir string) error {
	return SetDefaultPython(pythonDir)
}
