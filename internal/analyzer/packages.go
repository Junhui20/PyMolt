package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Package represents an installed Python package.
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// OutdatedPackage is a package with an available update.
type OutdatedPackage struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
}

// ListPackages returns all installed packages for a given Python executable.
func ListPackages(executable string) ([]Package, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, "-m", "pip", "list", "--format=json", "--disable-pip-version-check")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var pkgs []Package
	if err := json.Unmarshal(out, &pkgs); err != nil {
		return nil, err
	}
	return pkgs, nil
}

// ListOutdated returns packages with available updates.
func ListOutdated(executable string) ([]OutdatedPackage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, "-m", "pip", "list", "--outdated", "--format=json", "--disable-pip-version-check")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var raw []struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		LatestVersion string `json:"latest_version"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}

	var result []OutdatedPackage
	for _, r := range raw {
		result = append(result, OutdatedPackage{
			Name:           r.Name,
			CurrentVersion: r.Version,
			LatestVersion:  r.LatestVersion,
		})
	}
	return result, nil
}

// InstallPackage installs a package into the given Python environment.
func InstallPackage(executable, packageName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, "-m", "pip", "install", packageName, "--disable-pip-version-check")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// UninstallPackage removes a package from the given Python environment.
func UninstallPackage(executable, packageName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, "-m", "pip", "uninstall", packageName, "-y", "--disable-pip-version-check")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// ExportRequirements generates requirements.txt content.
func ExportRequirements(executable string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, "-m", "pip", "freeze", "--disable-pip-version-check")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
