package analyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PythonVersion represents an installable Python version.
type PythonVersion struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
}

// ListPythonVersions returns all available and installed Python versions via uv.
func ListPythonVersions() ([]PythonVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "uv", "python", "list")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var versions []PythonVersion
	seen := map[string]bool{}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		id := parts[0]
		idParts := strings.Split(id, "-")
		if len(idParts) < 2 {
			continue
		}
		ver := idParts[1]

		if strings.Contains(id, "freethreaded") {
			continue
		}

		majorMinor := extractMajorMinor(ver)
		installed := !strings.Contains(line, "<download available>")
		path := ""
		if installed {
			path = parts[len(parts)-1]
		}

		if seen[majorMinor] {
			for i, v := range versions {
				if extractMajorMinor(v.Version) == majorMinor && !v.Installed && installed {
					versions[i].Installed = true
					versions[i].Path = path
					versions[i].Version = ver
				}
			}
			continue
		}

		seen[majorMinor] = true
		versions = append(versions, PythonVersion{
			Version:   ver,
			Installed: installed,
			Path:      path,
		})
	}

	return versions, nil
}

func extractMajorMinor(ver string) string {
	parts := strings.SplitN(ver, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return ver
}

// InstallPythonVersion installs a Python version via uv.
func InstallPythonVersion(version string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "uv", "python", "install", version)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// UninstallPythonVersion removes a Python version via uv.
func UninstallPythonVersion(version string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "uv", "python", "uninstall", version)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
