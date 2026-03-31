//go:build darwin

package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// OfficialDetector finds Python installed via python.org installer on macOS.
type OfficialDetector struct{}

func (d OfficialDetector) Name() string { return "Official Installer" }

func (d OfficialDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Python.org Framework installations
	results = append(results, d.fromFramework()...)

	// Common binary paths
	results = append(results, d.fromKnownPaths()...)

	return dedup(results)
}

func (d OfficialDetector) fromFramework() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Python.org installs to /Library/Frameworks/Python.framework/Versions/X.Y
	matches, err := filepath.Glob("/Library/Frameworks/Python.framework/Versions/*/bin")
	if err != nil {
		return nil
	}
	for _, binDir := range matches {
		base := filepath.Dir(binDir)
		ver := filepath.Base(base)
		if ver == "Current" {
			continue
		}
		inst := MakeInstallation(binDir, models.SourceOfficial)
		if inst != nil {
			results = append(results, *inst)
		}
	}
	return results
}

func (d OfficialDetector) fromKnownPaths() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Check common locations
	dirs := []string{
		"/usr/local/bin",
		"/usr/bin",
	}

	for _, dir := range dirs {
		for _, name := range pythonExeNames() {
			exe := filepath.Join(dir, name)
			if _, err := os.Stat(exe); err != nil {
				continue
			}
			// Skip if it's a symlink to Homebrew (handled by HomebrewDetector)
			if target, err := os.Readlink(exe); err == nil {
				if strings.Contains(target, "homebrew") || strings.Contains(target, "Cellar") {
					continue
				}
			}
			// Skip Apple's system Python stub
			if isAppleStub(exe) {
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
				Source:       models.SourceOfficial,
				SizeBytes:    0, // Don't measure system dirs
				InPath:       IsInPath(dir),
				Architecture: GetArchitecture(exe),
			})
		}
	}
	return results
}

// isAppleStub returns true if the executable is Apple's developer tools placeholder.
func isAppleStub(exe string) bool {
	cmd := exec.Command(exe, "--version")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return true // Apple stub often triggers xcode-select dialog
	}
	output := string(out)
	return strings.Contains(output, "xcode-select") || strings.Contains(output, "CommandLineTools")
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
