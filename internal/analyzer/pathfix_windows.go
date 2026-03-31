//go:build windows

package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// RemoveOrphanedPaths removes PATH entries that point to non-existent directories.
func RemoveOrphanedPaths() (int, error) {
	removed := 0

	// Fix User PATH
	n, err := cleanPath(registry.CURRENT_USER, `Environment`)
	if err != nil {
		return 0, fmt.Errorf("failed to fix user PATH: %w", err)
	}
	removed += n

	return removed, nil
}

// SetDefaultPython reorders User PATH so the given Python directory comes first.
func SetDefaultPython(pythonDir string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}
	defer key.Close()

	val, _, err := key.GetStringValue("Path")
	if err != nil {
		return fmt.Errorf("failed to read PATH: %w", err)
	}

	entries := filepath.SplitList(val)
	normTarget := strings.ToLower(filepath.Clean(pythonDir))

	var targetEntries []string
	var otherEntries []string

	for _, e := range entries {
		if strings.ToLower(filepath.Clean(e)) == normTarget {
			targetEntries = append(targetEntries, e)
		} else {
			otherEntries = append(otherEntries, e)
		}
	}

	if len(targetEntries) == 0 {
		// Add the Python dir to the front
		targetEntries = []string{pythonDir}
	}

	newPath := strings.Join(append(targetEntries, otherEntries...), ";")
	return key.SetStringValue("Path", newPath)
}

// BackupPATH saves current PATH to a file for safety.
func BackupPATH() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	val, _, err := key.GetStringValue("Path")
	if err != nil {
		return "", err
	}

	backupDir := filepath.Join(os.Getenv("APPDATA"), "PythonManager")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	backupFile := filepath.Join(backupDir, "path_backup.txt")
	err = os.WriteFile(backupFile, []byte(val), 0644)
	return backupFile, err
}

func cleanPath(root registry.Key, subkey string) (int, error) {
	key, err := registry.OpenKey(root, subkey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return 0, err
	}
	defer key.Close()

	val, _, err := key.GetStringValue("Path")
	if err != nil {
		return 0, err
	}

	entries := filepath.SplitList(val)
	var clean []string
	removed := 0

	for _, e := range entries {
		if e == "" {
			continue
		}
		lower := strings.ToLower(e)
		isPython := strings.Contains(lower, "python") || strings.Contains(lower, "pyenv")

		if isPython {
			if _, err := os.Stat(e); err != nil {
				// Orphaned Python PATH entry - remove it
				removed++
				continue
			}
		}
		clean = append(clean, e)
	}

	if removed > 0 {
		newPath := strings.Join(clean, ";")
		if err := key.SetStringValue("Path", newPath); err != nil {
			return 0, err
		}
	}
	return removed, nil
}
