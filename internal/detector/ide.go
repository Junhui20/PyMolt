package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// IDEDetector finds Python interpreters configured in VS Code and PyCharm.
type IDEDetector struct{}

func (d IDEDetector) Name() string { return "IDE Configs" }

func (d IDEDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	results = append(results, d.fromVSCode()...)
	results = append(results, d.fromPyCharm()...)

	return results
}

func (d IDEDetector) fromVSCode() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Find VS Code settings.json
	settingsPaths := vsCodeSettingsPaths()

	for _, settingsPath := range settingsPaths {
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			continue
		}

		var settings map[string]interface{}
		if json.Unmarshal(data, &settings) != nil {
			continue
		}

		// Check python.defaultInterpreterPath and python.pythonPath (legacy)
		for _, key := range []string{"python.defaultInterpreterPath", "python.pythonPath"} {
			if val, ok := settings[key]; ok {
				if path, ok := val.(string); ok && path != "" && path != "python" {
					path = expandPath(path)
					if _, err := os.Stat(path); err == nil {
						version := GetPythonVersion(path)
						if version != "" {
							results = append(results, models.PythonInstallation{
								Version:      version,
								MajorMinor:   ExtractMajorMinor(version),
								Path:         filepath.Dir(path),
								Executable:   path,
								Source:       models.SourceIDE,
								SizeBytes:    0,
								InPath:       false,
								Architecture: GetArchitecture(path),
							})
						}
					}
				}
			}
		}
	}
	return results
}

func (d IDEDetector) fromPyCharm() []models.PythonInstallation {
	var results []models.PythonInstallation

	// Scan workspace .idea/misc.xml for Python SDK
	home := HomeDir()
	workspaceDirs := []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Projects"),
		filepath.Join(home, "repos"),
		filepath.Join(home, "dev"),
	}
	if runtime.GOOS != "windows" {
		workspaceDirs = append(workspaceDirs,
			filepath.Join(home, "projects"),
			filepath.Join(home, "src"),
		)
	}

	seen := map[string]bool{}
	for _, wsDir := range workspaceDirs {
		entries, err := os.ReadDir(wsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			miscXML := filepath.Join(wsDir, entry.Name(), ".idea", "misc.xml")
			data, err := os.ReadFile(miscXML)
			if err != nil {
				continue
			}
			content := string(data)
			// Look for project-jdk-name containing Python path
			idx := strings.Index(content, "project-jdk-name=")
			if idx < 0 {
				continue
			}
			rest := content[idx+len("project-jdk-name="):]
			if len(rest) < 2 || rest[0] != '"' {
				continue
			}
			end := strings.Index(rest[1:], "\"")
			if end < 0 {
				continue
			}
			jdkName := rest[1 : end+1]
			// PyCharm SDK name format: "Python 3.13 (/path/to/python)"
			if !strings.Contains(jdkName, "Python") {
				continue
			}
			// Extract path from parentheses
			start := strings.LastIndex(jdkName, "(")
			endP := strings.LastIndex(jdkName, ")")
			if start < 0 || endP < 0 || endP <= start {
				continue
			}
			pythonPath := jdkName[start+1 : endP]
			if seen[pythonPath] {
				continue
			}
			seen[pythonPath] = true

			if _, err := os.Stat(pythonPath); err != nil {
				continue
			}
			version := GetPythonVersion(pythonPath)
			if version != "" {
				results = append(results, models.PythonInstallation{
					Version:      version,
					MajorMinor:   ExtractMajorMinor(version),
					Path:         filepath.Dir(pythonPath),
					Executable:   pythonPath,
					Source:       models.SourceIDE,
					SizeBytes:    0,
					InPath:       false,
					Architecture: GetArchitecture(pythonPath),
				})
			}
		}
	}
	return results
}

func vsCodeSettingsPaths() []string {
	home := HomeDir()
	var paths []string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		paths = append(paths,
			filepath.Join(appData, "Code", "User", "settings.json"),
			filepath.Join(appData, "Code - Insiders", "User", "settings.json"),
		)
	case "darwin":
		paths = append(paths,
			filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"),
			filepath.Join(home, "Library", "Application Support", "Code - Insiders", "User", "settings.json"),
		)
	default: // linux
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		paths = append(paths,
			filepath.Join(configDir, "Code", "User", "settings.json"),
			filepath.Join(configDir, "Code - Insiders", "User", "settings.json"),
		)
	}
	return paths
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		return filepath.Join(HomeDir(), path[1:])
	}
	return path
}
