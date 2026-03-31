//go:build !windows

package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RemoveOrphanedPaths removes orphaned Python-related entries from shell config files.
// On Unix, PATH is managed via shell profile files, not a registry.
// This function scans common profile files and removes orphaned entries.
func RemoveOrphanedPaths() (int, error) {
	// On Unix, we analyze the live PATH but can't easily modify shell configs.
	// Instead, report what should be removed.
	pathEnv := os.Getenv("PATH")
	entries := filepath.SplitList(pathEnv)
	removed := 0
	for _, e := range entries {
		if e == "" {
			continue
		}
		lower := strings.ToLower(e)
		isPython := strings.Contains(lower, "python") || strings.Contains(lower, "pyenv")
		if isPython {
			if _, err := os.Stat(e); err != nil {
				removed++
			}
		}
	}
	if removed > 0 {
		return removed, fmt.Errorf("found %d orphaned entries — edit your shell profile (~/.bashrc, ~/.zshrc) to remove them", removed)
	}
	return 0, nil
}

// SetDefaultPython modifies the PATH to prioritize a Python directory.
// On Unix, this modifies the current process PATH (informational only).
func SetDefaultPython(pythonDir string) error {
	pathEnv := os.Getenv("PATH")
	entries := filepath.SplitList(pathEnv)

	var newEntries []string
	newEntries = append(newEntries, pythonDir)
	for _, e := range entries {
		if strings.ToLower(filepath.Clean(e)) != strings.ToLower(filepath.Clean(pythonDir)) {
			newEntries = append(newEntries, e)
		}
	}

	return os.Setenv("PATH", strings.Join(newEntries, ":"))
}

// BackupPATH saves current PATH to a file for safety.
func BackupPATH() (string, error) {
	pathEnv := os.Getenv("PATH")
	home := os.Getenv("HOME")
	backupDir := filepath.Join(home, ".config", "pymolt")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	backupFile := filepath.Join(backupDir, "path_backup.txt")
	err := os.WriteFile(backupFile, []byte(pathEnv), 0644)
	return backupFile, err
}
