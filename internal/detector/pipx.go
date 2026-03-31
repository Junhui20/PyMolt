package detector

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// PipxDetector finds Python interpreters used by pipx-installed tools.
type PipxDetector struct{}

func (d PipxDetector) Name() string { return "pipx" }

func (d PipxDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	venvsDir := pipxVenvsDir()
	entries, err := os.ReadDir(venvsDir)
	if err != nil {
		return nil
	}

	seen := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		venvDir := filepath.Join(venvsDir, entry.Name())
		scriptsDir := filepath.Join(venvDir, venvScriptsDir())
		exe := findExecutable(scriptsDir)
		if exe == "" {
			continue
		}
		version := GetPythonVersion(exe)
		if version == "" || seen[version] {
			continue
		}
		seen[version] = true

		results = append(results, models.PythonInstallation{
			Version:    version,
			MajorMinor: ExtractMajorMinor(version),
			Path:       venvDir,
			Executable: exe,
			Source:     models.SourceVenv,
			SizeBytes:  DirSize(venvDir),
			InPath:     false,
		})
	}
	return results
}

func pipxVenvsDir() string {
	if d := os.Getenv("PIPX_HOME"); d != "" {
		return filepath.Join(d, "venvs")
	}
	home := HomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "pipx", "venvs")
	}
	return filepath.Join(home, ".local", "pipx", "venvs")
}
