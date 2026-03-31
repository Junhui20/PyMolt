package detector

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// ChocolateyDetector finds Python installed via Chocolatey.
type ChocolateyDetector struct{}

func (d ChocolateyDetector) Name() string { return "Chocolatey" }

func (d ChocolateyDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	chocoLib := filepath.Join(os.Getenv("ChocolateyInstall"), "lib")
	if chocoLib == `\lib` {
		chocoLib = `C:\ProgramData\chocolatey\lib`
	}

	matches, err := filepath.Glob(filepath.Join(chocoLib, "python*"))
	if err != nil {
		return nil
	}

	for _, dir := range matches {
		toolsDir := filepath.Join(dir, "tools")
		inst := MakeInstallation(toolsDir, models.SourceChocolatey)
		if inst != nil {
			results = append(results, *inst)
		}
		// Also check the dir itself
		inst = MakeInstallation(dir, models.SourceChocolatey)
		if inst != nil {
			results = append(results, *inst)
		}
	}

	// Check chocolatey bin for python shims
	chocoBin := filepath.Join(filepath.Dir(chocoLib), "bin")
	for _, name := range []string{"python.exe", "python3.exe"} {
		exe := filepath.Join(chocoBin, name)
		if _, err := os.Stat(exe); err == nil {
			version := GetPythonVersion(exe)
			if version != "" {
				results = append(results, models.PythonInstallation{
					Version:    version,
					MajorMinor: ExtractMajorMinor(version),
					Path:       chocoBin,
					Executable: exe,
					Source:     models.SourceChocolatey,
					SizeBytes:  0, // shim only
					InPath:     IsInPath(chocoBin),
				})
			}
		}
	}
	return results
}
