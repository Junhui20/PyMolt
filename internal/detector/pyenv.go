package detector

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Junhui20/PyMolt/internal/models"
)

// PyenvDetector finds Python versions installed by pyenv (or pyenv-win).
type PyenvDetector struct{}

func (d PyenvDetector) Name() string { return "pyenv" }

func (d PyenvDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	roots := pyenvRoots()

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
			// On Unix, python is in <version>/bin/; MakeInstallation handles this
			inst := MakeInstallation(dir, models.SourcePyenv)
			if inst == nil {
				// Try bin/ subdirectory (Unix layout)
				inst = MakeInstallation(filepath.Join(dir, "bin"), models.SourcePyenv)
			}
			if inst != nil {
				results = append(results, *inst)
			}
		}
		break // only use first valid root
	}
	return results
}

func pyenvRoots() []string {
	roots := []string{
		os.Getenv("PYENV"),
		os.Getenv("PYENV_ROOT"),
		os.Getenv("PYENV_HOME"),
	}

	if runtime.GOOS == "windows" {
		roots = append(roots, filepath.Join(HomeDir(), ".pyenv", "pyenv-win"))
	} else {
		roots = append(roots, filepath.Join(HomeDir(), ".pyenv"))
	}

	return roots
}
