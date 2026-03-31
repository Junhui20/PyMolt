//go:build windows

package detector

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// StoreDetector finds Python from the Microsoft Store.
type StoreDetector struct{}

func (d StoreDetector) Name() string { return "Microsoft Store" }

func (d StoreDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	appsDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps")
	for _, name := range []string{"python.exe", "python3.exe"} {
		exe := filepath.Join(appsDir, name)
		if _, err := os.Stat(exe); err != nil {
			continue
		}
		version := GetPythonVersion(exe)
		if version == "" {
			continue
		}
		results = append(results, models.PythonInstallation{
			Version:    version,
			MajorMinor: ExtractMajorMinor(version),
			Path:       appsDir,
			Executable: exe,
			Source:     models.SourceStore,
			SizeBytes:  0,
			InPath:     IsInPath(appsDir),
		})
	}
	return results
}
