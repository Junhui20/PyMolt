package detector

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// ScoopDetector finds Python installed via Scoop.
type ScoopDetector struct{}

func (d ScoopDetector) Name() string { return "Scoop" }

func (d ScoopDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	scoopDir := os.Getenv("SCOOP")
	if scoopDir == "" {
		scoopDir = filepath.Join(os.Getenv("USERPROFILE"), "scoop")
	}

	appsDir := filepath.Join(scoopDir, "apps", "python")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "current" {
			continue
		}
		dir := filepath.Join(appsDir, entry.Name())
		inst := MakeInstallation(dir, models.SourceScoop)
		if inst != nil {
			results = append(results, *inst)
		}
	}
	return results
}
