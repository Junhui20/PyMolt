//go:build windows

package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// AnalyzePATH inspects user and system PATH for Python-related entries.
func AnalyzePATH() *PathAnalysis {
	result := &PathAnalysis{}

	userEntries := getRegistryPath(registry.CURRENT_USER)
	sysEntries := getRegistryPath(registry.LOCAL_MACHINE)

	priority := 0
	seenVersion := map[string]bool{}
	firstPython := ""

	for _, entry := range append(sysEntries, userEntries...) {
		source := "System"
		if priority >= len(sysEntries) {
			source = "User"
		}

		lower := strings.ToLower(entry)
		isPythonRelated := strings.Contains(lower, "python") ||
			strings.Contains(lower, "pyenv") ||
			strings.Contains(lower, "conda") ||
			strings.Contains(lower, "uv") ||
			strings.Contains(lower, "virtualenv")

		if !isPythonRelated {
			priority++
			continue
		}

		pe := PathEntry{
			Path:     entry,
			Source:   source,
			Priority: priority,
		}

		if info, err := os.Stat(entry); err == nil && info.IsDir() {
			pe.Exists = true
			for _, name := range []string{"python.exe", "python3.exe"} {
				exe := filepath.Join(entry, name)
				if _, err := os.Stat(exe); err == nil {
					pe.HasPython = true
					pe.Version = getPythonVersionQuick(exe)
					if firstPython == "" {
						firstPython = pe.Version
					} else if pe.Version != "" {
						pe.Shadowed = true
					}
					if pe.Version != "" {
						if seenVersion[pe.Version] {
							result.Conflicts = append(result.Conflicts,
								"Python "+pe.Version+" appears in PATH multiple times")
						}
						seenVersion[pe.Version] = true
					}
					break
				}
			}
		} else {
			pe.Orphaned = true
			result.OrphanedCount++
			result.Conflicts = append(result.Conflicts,
				"Orphaned PATH entry: "+entry+" (directory does not exist)")
		}

		result.Entries = append(result.Entries, pe)
		priority++
	}

	result.DefaultPython = firstPython
	return result
}

func getRegistryPath(root registry.Key) []string {
	var subkey string
	if root == registry.CURRENT_USER {
		subkey = `Environment`
	} else {
		subkey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	}
	k, err := registry.OpenKey(root, subkey, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()
	val, _, err := k.GetStringValue("Path")
	if err != nil {
		return nil
	}
	return filepath.SplitList(val)
}
