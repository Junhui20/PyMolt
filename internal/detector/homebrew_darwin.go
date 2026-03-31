//go:build darwin

package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// HomebrewDetector finds Python installed via Homebrew on macOS.
type HomebrewDetector struct{}

func (d HomebrewDetector) Name() string { return "Homebrew" }

func (d HomebrewDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Find Homebrew prefix
	prefix := homebrewPrefix()
	if prefix == "" {
		return nil
	}

	// Check Homebrew Cellar for python@ formula
	cellarDir := filepath.Join(prefix, "Cellar")
	matches, err := filepath.Glob(filepath.Join(cellarDir, "python@*"))
	if err != nil {
		return nil
	}

	for _, pyDir := range matches {
		entries, err := os.ReadDir(pyDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			binDir := filepath.Join(pyDir, entry.Name(), "bin")
			inst := MakeInstallation(binDir, models.SourceHomebrew)
			if inst != nil {
				results = append(results, *inst)
			}
		}
	}

	// Also check the Homebrew bin for linked python3
	brewBin := filepath.Join(prefix, "bin")
	for _, name := range pythonExeNames() {
		exe := filepath.Join(brewBin, name)
		if _, err := os.Stat(exe); err != nil {
			continue
		}
		// Verify it's actually from Homebrew
		if target, err := os.Readlink(exe); err == nil {
			if !strings.Contains(target, "Cellar") && !strings.Contains(target, "homebrew") {
				continue
			}
		}
		version := GetPythonVersion(exe)
		if version == "" {
			continue
		}
		// Check if already found via Cellar scan
		found := false
		for _, r := range results {
			if r.Version == version {
				found = true
				break
			}
		}
		if !found {
			results = append(results, models.PythonInstallation{
				Version:      version,
				MajorMinor:   ExtractMajorMinor(version),
				Path:         brewBin,
				Executable:   exe,
				Source:       models.SourceHomebrew,
				SizeBytes:    0,
				InPath:       IsInPath(brewBin),
				Architecture: GetArchitecture(exe),
			})
		}
	}

	return results
}

func homebrewPrefix() string {
	cmd := exec.Command("brew", "--prefix")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		// Fallback to common locations
		for _, p := range []string{"/opt/homebrew", "/usr/local"} {
			if _, err := os.Stat(filepath.Join(p, "bin", "brew")); err == nil {
				return p
			}
		}
		return ""
	}
	return strings.TrimSpace(string(out))
}
