//go:build linux

package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// OfficialDetector finds Python installed via system package manager or python.org on Linux.
type OfficialDetector struct{}

func (d OfficialDetector) Name() string { return "System / Official" }

func (d OfficialDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	results = append(results, d.fromKnownPaths()...)
	results = append(results, d.fromDeadsnakes()...)

	return dedup(results)
}

func (d OfficialDetector) fromKnownPaths() []models.PythonInstallation {
	var results []models.PythonInstallation

	dirs := []string{
		"/usr/bin",
		"/usr/local/bin",
		"/snap/bin",
	}

	for _, dir := range dirs {
		for _, name := range pythonExeNames() {
			exe := filepath.Join(dir, name)
			if _, err := os.Stat(exe); err != nil {
				continue
			}
			version := GetPythonVersion(exe)
			if version == "" {
				continue
			}
			source := models.SourceSystem
			if dir == "/usr/local/bin" {
				source = models.SourceOfficial
			}
			results = append(results, models.PythonInstallation{
				Version:      version,
				MajorMinor:   ExtractMajorMinor(version),
				Path:         dir,
				Executable:   exe,
				Source:       source,
				SizeBytes:    0, // Don't measure system dirs
				InPath:       IsInPath(dir),
				Architecture: GetArchitecture(exe),
			})
		}

		// Also check versioned binaries like python3.11, python3.12
		matches, _ := filepath.Glob(filepath.Join(dir, "python3.*"))
		for _, exe := range matches {
			// Skip .so files, .py files, etc.
			base := filepath.Base(exe)
			if strings.Contains(base, ".") && !strings.HasPrefix(base, "python3.") {
				continue
			}
			// Skip if already found
			info, err := os.Stat(exe)
			if err != nil || info.IsDir() {
				continue
			}
			version := GetPythonVersion(exe)
			if version == "" {
				continue
			}
			results = append(results, models.PythonInstallation{
				Version:      version,
				MajorMinor:   ExtractMajorMinor(version),
				Path:         dir,
				Executable:   exe,
				Source:       models.SourceSystem,
				SizeBytes:    0,
				InPath:       IsInPath(dir),
				Architecture: GetArchitecture(exe),
			})
		}
	}
	return results
}

func (d OfficialDetector) fromDeadsnakes() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Deadsnakes PPA installs to /usr/bin/pythonX.Y
	// These are already picked up by fromKnownPaths, but we re-tag them
	// Check if deadsnakes is in apt sources
	if _, err := os.Stat("/etc/apt/sources.list.d"); err != nil {
		return nil
	}

	matches, _ := filepath.Glob("/etc/apt/sources.list.d/*deadsnakes*")
	if len(matches) == 0 {
		return nil
	}

	// deadsnakes is available, re-scan for alternate Python versions
	// They're installed as /usr/bin/python3.X and handled in fromKnownPaths
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
