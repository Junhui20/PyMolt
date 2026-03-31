package detector

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// MiseDetector finds Python versions installed by mise (formerly rtx).
type MiseDetector struct{}

func (d MiseDetector) Name() string { return "mise" }

func (d MiseDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	dataDirs := miseDataDirs()
	for _, dataDir := range dataDirs {
		pythonDir := filepath.Join(dataDir, "installs", "python")
		entries, err := os.ReadDir(pythonDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dir := filepath.Join(pythonDir, entry.Name())
			inst := MakeInstallation(dir, models.SourcePyenv)
			if inst == nil && runtime.GOOS != "windows" {
				inst = MakeInstallation(filepath.Join(dir, "bin"), models.SourcePyenv)
			}
			if inst != nil {
				results = append(results, *inst)
			}
		}
	}
	return results
}

func miseDataDirs() []string {
	var dirs []string

	if d := os.Getenv("MISE_DATA_DIR"); d != "" {
		dirs = append(dirs, d)
	}

	home := HomeDir()
	if runtime.GOOS == "windows" {
		dirs = append(dirs, filepath.Join(os.Getenv("LOCALAPPDATA"), "mise"))
	} else {
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			dirs = append(dirs, filepath.Join(xdg, "mise"))
		}
		dirs = append(dirs, filepath.Join(home, ".local", "share", "mise"))
	}
	return dirs
}
