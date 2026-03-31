package detector

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// UVDetector finds Python versions installed by uv.
type UVDetector struct{}

func (d UVDetector) Name() string { return "uv" }

func (d UVDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	uvDir := uvPythonDir()
	entries, err := os.ReadDir(uvDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(uvDir, entry.Name())
		inst := MakeInstallation(dir, models.SourceUV)
		if inst != nil {
			results = append(results, *inst)
		}
	}
	return results
}

func uvPythonDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "uv", "python")
	}
	// Unix: ~/.local/share/uv/python
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "uv", "python")
	}
	return filepath.Join(HomeDir(), ".local", "share", "uv", "python")
}
