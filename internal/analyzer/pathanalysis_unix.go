//go:build !windows

package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// AnalyzePATH inspects PATH for Python-related entries on Unix systems.
func AnalyzePATH() *PathAnalysis {
	result := &PathAnalysis{}

	pathEnv := os.Getenv("PATH")
	entries := filepath.SplitList(pathEnv)

	priority := 0
	seenVersion := map[string]bool{}
	firstPython := ""

	for _, entry := range entries {
		lower := strings.ToLower(entry)
		isPythonRelated := strings.Contains(lower, "python") ||
			strings.Contains(lower, "pyenv") ||
			strings.Contains(lower, "conda") ||
			strings.Contains(lower, "uv") ||
			strings.Contains(lower, "virtualenv") ||
			strings.Contains(lower, "homebrew")

		// Also check common bin dirs
		if !isPythonRelated {
			if entry == "/usr/bin" || entry == "/usr/local/bin" || entry == "/opt/homebrew/bin" {
				isPythonRelated = true
			}
		}

		if !isPythonRelated {
			priority++
			continue
		}

		pe := PathEntry{
			Path:     entry,
			Source:   "Environment",
			Priority: priority,
		}

		if info, err := os.Stat(entry); err == nil && info.IsDir() {
			pe.Exists = true
			for _, name := range []string{"python3", "python"} {
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
