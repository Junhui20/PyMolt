package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
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
	case models.SourceHomebrew:
		return uninstallHomebrew(inst)
	default:
		return deleteDirectory(inst)
	}
}

// dangerousPaths lists paths that must never be deleted.
var dangerousPaths = map[string]bool{
	"/":            true,
	"/usr":         true,
	"/usr/bin":     true,
	"/usr/local":   true,
	"/usr/local/bin": true,
	"/home":        true,
	"/opt":         true,
}

func deleteDirectory(inst models.PythonInstallation) *UninstallResult {
	clean := strings.ToLower(filepath.Clean(inst.Path))

	// Check platform-specific dangerous paths
	if isDangerousPath(clean) {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Refusing to delete protected path: %s", inst.Path)}
	}
	if dangerousPaths[clean] {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Refusing to delete protected path: %s", inst.Path)}
	}

	size := inst.SizeBytes
	err := os.RemoveAll(inst.Path)
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Failed to delete %s: %v", inst.Path, err)}
	}
	return &UninstallResult{Success: true, Message: fmt.Sprintf("Deleted %s", inst.Path), SpaceFreed: size}
}

// uninstallUV, uninstallOfficial, uninstallChocolatey, uninstallScoop, uninstallHomebrew,
// isDangerousPath, cache functions
// are defined in uninstall_windows.go / uninstall_unix.go
