package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
	"golang.org/x/sys/windows/registry"
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

	keys := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Python\PythonCore`},
		{registry.CURRENT_USER, `SOFTWARE\Python\PythonCore`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Python\PythonCore`},
	}

	for _, k := range keys {
		parent, err := registry.OpenKey(k.root, k.path, registry.READ)
		if err != nil {
			continue
		}
		versions, err := parent.ReadSubKeyNames(-1)
		parent.Close()
		if err != nil {
			continue
		}
		for _, ver := range versions {
			installKey, err := registry.OpenKey(k.root, k.path+`\`+ver+`\InstallPath`, registry.READ)
			if err != nil {
				continue
			}
			installPath, _, err := installKey.GetStringValue("")
			installKey.Close()
			if err != nil || installPath == "" {
				continue
			}
			inst := MakeInstallation(strings.TrimRight(installPath, `\`), models.SourceOfficial)
			if inst != nil {
				results = append(results, *inst)
			}
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
