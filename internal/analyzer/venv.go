package analyzer

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// CreateVenvResult reports venv creation outcome.
type CreateVenvResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

// CreateVenv creates a new virtual environment.
func CreateVenv(pythonExe, targetDir, name string) *CreateVenvResult {
	venvPath := targetDir
	if name != "" {
		venvPath = targetDir + "\\" + name
	}

	// Prefer uv if available
	uvPath, err := exec.LookPath("uv")
	if err == nil && uvPath != "" {
		cmd := exec.Command("uv", "venv", venvPath, "--python", pythonExe)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return &CreateVenvResult{Success: false, Message: fmt.Sprintf("uv venv failed: %s\n%s", err, string(out))}
		}
		return &CreateVenvResult{Success: true, Message: strings.TrimSpace(string(out)), Path: venvPath}
	}

	// Fallback to python -m venv
	cmd := exec.Command(pythonExe, "-m", "venv", venvPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &CreateVenvResult{Success: false, Message: fmt.Sprintf("venv creation failed: %s\n%s", err, string(out))}
	}
	return &CreateVenvResult{Success: true, Message: "Created virtual environment", Path: venvPath}
}
