//go:build windows

package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// OfficialDetector finds Python installed via python.org installer.
type OfficialDetector struct{}

func (d OfficialDetector) Name() string { return "Official Installer" }

func (d OfficialDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Check Windows Registry
	results = append(results, d.fromRegistry()...)

	// Check known installation directories
	results = append(results, d.fromKnownPaths()...)

	return dedup(results)
}

func (d OfficialDetector) fromRegistry() []models.PythonInstallation {
	var results []models.PythonInstallation
	// Vendor-agnostic PEP 514 read: every Company\Tag under SOFTWARE\Python, not
	// just PythonCore. ContinuumAnalytics (conda) and %LocalAppData%\Python
	// (PyManager) are owned by their dedicated detectors, so skip them here.
	for _, e := range enumeratePEP514() {
		if strings.EqualFold(e.Company, "ContinuumAnalytics") || underLocalAppDataPython(e.InstallPath) {
			continue
		}
		if inst := MakeInstallation(e.InstallPath, models.SourceOfficial); inst != nil {
			results = append(results, *inst)
		}
	}
	return results
}

func (d OfficialDetector) fromKnownPaths() []models.PythonInstallation {
	var results []models.PythonInstallation

	patterns := []string{
		`C:\Python*`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Python*"),
		`C:\Program Files\Python*`,
		`C:\Program Files (x86)\Python*`,
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, dir := range matches {
			inst := MakeInstallation(dir, models.SourceOfficial)
			if inst != nil {
				results = append(results, *inst)
			}
		}
	}
	return results
}

func dedup(installs []models.PythonInstallation) []models.PythonInstallation {
	seen := make(map[string]bool)
	var result []models.PythonInstallation
	for _, inst := range installs {
		key := strings.ToLower(filepath.Clean(inst.Executable))
		if !seen[key] {
			seen[key] = true
			result = append(result, inst)
		}
	}
	return result
}
