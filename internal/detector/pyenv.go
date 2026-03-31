package detector

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// PyenvDetector finds Python versions installed by pyenv-win.
type PyenvDetector struct{}

func (d PyenvDetector) Name() string { return "pyenv-win" }

func (d PyenvDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Check PYENV, PYENV_ROOT, PYENV_HOME env vars, fallback to default
	roots := []string{
		os.Getenv("PYENV"),
		os.Getenv("PYENV_ROOT"),
		os.Getenv("PYENV_HOME"),
		filepath.Join(os.Getenv("USERPROFILE"), ".pyenv", "pyenv-win"),
	}

	for _, root := range roots {
		if root == "" {
			continue
		}
		versionsDir := filepath.Join(root, "versions")
		entries, err := os.ReadDir(versionsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dir := filepath.Join(versionsDir, entry.Name())
			inst := MakeInstallation(dir, models.SourcePyenv)
			if inst != nil {
				results = append(results, *inst)
			}
		}
		break // only use first valid root
	}
	return results
}
