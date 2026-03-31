package detector

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// CondaDetector finds Python in conda/miniconda/anaconda environments.
type CondaDetector struct{}

func (d CondaDetector) Name() string { return "Conda" }

func (d CondaDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	bases := condaBases()

	for _, base := range bases {
		if base == "" {
			continue
		}
		// Check base environment
		inst := MakeInstallation(base, models.SourceConda)
		if inst == nil && runtime.GOOS != "windows" {
			inst = MakeInstallation(filepath.Join(base, "bin"), models.SourceConda)
		}
		if inst != nil {
			results = append(results, *inst)
		}
		// Check named environments
		envsDir := filepath.Join(base, "envs")
		entries, err := os.ReadDir(envsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			envDir := filepath.Join(envsDir, entry.Name())
			inst := MakeInstallation(envDir, models.SourceConda)
			if inst == nil && runtime.GOOS != "windows" {
				inst = MakeInstallation(filepath.Join(envDir, "bin"), models.SourceConda)
			}
			if inst != nil {
				results = append(results, *inst)
			}
		}
	}
	return results
}

func condaBases() []string {
	home := HomeDir()
	bases := []string{
		filepath.Join(home, "anaconda3"),
		filepath.Join(home, "miniconda3"),
		filepath.Join(home, "miniforge3"),
		os.Getenv("CONDA_PREFIX"),
	}

	if runtime.GOOS == "windows" {
		bases = append(bases,
			`C:\ProgramData\anaconda3`,
			`C:\ProgramData\miniconda3`,
		)
	} else {
		bases = append(bases,
			"/opt/anaconda3",
			"/opt/miniconda3",
			"/opt/miniforge3",
		)
	}

	return bases
}
