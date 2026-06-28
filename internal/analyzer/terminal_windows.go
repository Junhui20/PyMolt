//go:build windows

package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		// Defense in depth: a quote or newline in the path could break out of the
		// batch `set "PATH=..."` line. Windows filenames cannot contain these, but
		// reject them rather than build a malformed/injected command.
		if strings.ContainsAny(inst.Path, "\"\r\n") || strings.ContainsAny(inst.Version, "\"\r\n%") {
			return fmt.Errorf("refusing to open terminal: path or version contains invalid characters")
		}
		batContent := fmt.Sprintf("@echo off\r\ntitle Python %s\r\nset \"PATH=%s;%%PATH%%\"\r\npython --version\r\n",
			inst.Version, inst.Path)
		// Use a randomly-named temp file (not a predictable shared name) so another
		// user cannot pre-create or symlink it before we execute it.
		f, err := os.CreateTemp("", "pymolt_term_*.bat")
		if err != nil {
			return fmt.Errorf("failed to create temp bat: %w", err)
		}
		batFile := f.Name()
		if _, err := f.WriteString(batContent); err != nil {
			f.Close()
			return fmt.Errorf("failed to write temp bat: %w", err)
		}
		f.Close()
		cmd = exec.Command("cmd", "/k", batFile)
	}

	newConsole(cmd)
	return cmd.Start()
}

// AddToPATH adds a Python installation's directory to the User PATH.
func AddToPATH(pythonDir string) error {
	return SetDefaultPython(pythonDir)
}
