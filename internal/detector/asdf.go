package detector

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// AsdfDetector finds Python versions installed by asdf version manager.
type AsdfDetector struct{}

func (d AsdfDetector) Name() string { return "asdf" }

func (d AsdfDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	dataDir := os.Getenv("ASDF_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(HomeDir(), ".asdf")
	}

	pythonDir := filepath.Join(dataDir, "installs", "python")
	entries, err := os.ReadDir(pythonDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(pythonDir, entry.Name())
		// asdf puts executables in <version>/bin/ on Unix, root on Windows
		inst := MakeInstallation(dir, models.SourcePyenv) // reuse pyenv source label
		if inst == nil && runtime.GOOS != "windows" {
			inst = MakeInstallation(filepath.Join(dir, "bin"), models.SourcePyenv)
		}
		if inst != nil {
			results = append(results, *inst)
		}
	}
	return results
}
