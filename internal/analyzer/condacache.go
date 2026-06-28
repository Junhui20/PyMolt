package analyzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// condaRoots returns candidate conda/miniconda/anaconda install roots, mirroring
// the locations the conda detector checks. Windows paths are harmless on Unix
// (os.Stat simply fails for them).
func condaRoots() []string {
	home := userHome()
	return []string{
		os.Getenv("CONDA_PREFIX"),
		os.Getenv("CONDA_ROOT"),
		filepath.Join(home, "anaconda3"),
		filepath.Join(home, "miniconda3"),
		filepath.Join(home, "miniforge3"),
		filepath.Join(home, "mambaforge"),
		"/opt/anaconda3",
		"/opt/miniconda3",
		"/opt/miniforge3",
		"/opt/conda",
		`C:\ProgramData\anaconda3`,
		`C:\ProgramData\miniconda3`,
	}
}

// condaPkgsDirs returns the conda package-cache directories that exist on disk.
func condaPkgsDirs() []string {
	seen := map[string]bool{}
	var dirs []string
	add := func(d string) {
		if d == "" {
			return
		}
		c := filepath.Clean(d)
		if !seen[c] {
			seen[c] = true
			dirs = append(dirs, c)
		}
	}
	// The per-user shared cache, plus <root>/pkgs for each candidate root.
	add(filepath.Join(userHome(), ".conda", "pkgs"))
	for _, r := range condaRoots() {
		if r != "" {
			add(filepath.Join(r, "pkgs"))
		}
	}
	return dirs
}

// GetCondaCacheSize returns the total size of conda's package caches and a
// representative path (the first that exists), or (0, "") when conda isn't found.
func GetCondaCacheSize() (int64, string) {
	var total int64
	var firstPath string
	for _, d := range condaPkgsDirs() {
		if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
			continue
		}
		if firstPath == "" {
			firstPath = d
		}
		_ = filepath.Walk(d, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				total += info.Size()
			}
			return nil
		})
	}
	return total, firstPath
}

// findCondaExe locates the conda (or mamba) executable, preferring PATH and
// falling back to a known install root. Returns "" if none is found.
func findCondaExe() string {
	for _, name := range []string{"conda", "mamba"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	for _, r := range condaRoots() {
		if r == "" {
			continue
		}
		for _, rel := range []string{
			filepath.Join("condabin", "conda"),
			filepath.Join("bin", "conda"),
			filepath.Join("condabin", "conda.bat"),
			filepath.Join("Scripts", "conda.exe"),
		} {
			p := filepath.Join(r, rel)
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p
			}
		}
	}
	return ""
}

// CleanCondaCache runs `conda clean --all -y` to remove cached packages, index
// caches, and tarballs. It uses conda's own cleaner (rather than deleting files)
// so conda's bookkeeping stays consistent.
func CleanCondaCache() *UninstallResult {
	size, _ := GetCondaCacheSize()
	exe := findCondaExe()
	if exe == "" {
		return &UninstallResult{Success: false, Message: "conda not found. Open an Anaconda Prompt and run: conda clean --all"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "clean", "--all", "-y")
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("conda clean failed: %s", strings.TrimSpace(string(out)))}
	}
	msg := strings.TrimSpace(string(out))
	if msg == "" {
		msg = "Cleaned conda caches"
	}
	return &UninstallResult{Success: true, Message: msg, SpaceFreed: size}
}
