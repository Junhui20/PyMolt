//go:build windows

package detector

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// PyManagerDetector finds Python runtimes installed by Microsoft's PyManager
// (the `py install` manager, PEP 773), which live under %LocalAppData%\Python
// and register themselves via PEP 514.
type PyManagerDetector struct{}

func (d PyManagerDetector) Name() string { return "PyManager" }

func (d PyManagerDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Authoritative: PEP 514 registrations whose InstallPath is under
	// %LocalAppData%\Python (registry gives the exact runtime directory).
	for _, e := range enumeratePEP514() {
		if underLocalAppDataPython(e.InstallPath) {
			if inst := MakeInstallation(e.InstallPath, models.SourcePyManager); inst != nil {
				results = append(results, *inst)
			}
		}
	}

	// Fallback: scan the install root for runtimes that aren't (yet) registered.
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "Python")
	if entries, err := os.ReadDir(base); err == nil {
		for _, ent := range entries {
			if !ent.IsDir() {
				continue
			}
			if inst := MakeInstallation(filepath.Join(base, ent.Name()), models.SourcePyManager); inst != nil {
				results = append(results, *inst)
			}
		}
	}

	return dedup(results)
}
