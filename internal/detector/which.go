package detector

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// WhichDetector finds Python by running which/where to catch installations
// that other detectors might miss.
type WhichDetector struct{}

func (d WhichDetector) Name() string { return "PATH Discovery" }

func (d WhichDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation
	seen := map[string]bool{}

	names := pythonExeNames()
	for _, name := range names {
		paths := findAllInPath(name)
		for _, exe := range paths {
			clean := strings.ToLower(filepath.Clean(exe))
			if seen[clean] {
				continue
			}
			seen[clean] = true

			version := GetPythonVersion(exe)
			if version == "" {
				continue
			}

			dir := filepath.Dir(exe)
			results = append(results, models.PythonInstallation{
				Version:      version,
				MajorMinor:   ExtractMajorMinor(version),
				Path:         dir,
				Executable:   exe,
				Source:       models.SourceUnknown,
				SizeBytes:    0,
				InPath:       true,
				Architecture: GetArchitecture(exe),
			})
		}
	}

	return results
}

// findAllInPath finds all instances of a command in PATH.
func findAllInPath(name string) []string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where", name)
	} else {
		cmd = exec.Command("which", "-a", name)
	}
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths
}
