package detector

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// UVDetector finds Python versions installed by uv.
type UVDetector struct{}

func (d UVDetector) Name() string { return "uv" }

func (d UVDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	uvDir := filepath.Join(os.Getenv("APPDATA"), "uv", "python")
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
