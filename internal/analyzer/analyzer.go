package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
	"golang.org/x/sys/windows/registry"
)

// PathEntry represents one entry in the PATH related to Python.
type PathEntry struct {
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	HasPython bool   `json:"hasPython"`
	Version   string `json:"version"`   // Python version found, if any
	Source    string `json:"source"`    // "User" or "System"
	Priority  int    `json:"priority"`  // lower = higher priority (resolved first)
	Orphaned  bool   `json:"orphaned"`  // path doesn't exist on disk
	Shadowed  bool   `json:"shadowed"`  // another Python entry comes before this one
}

// PathAnalysis is the full PATH report.
type PathAnalysis struct {
	Entries       []PathEntry `json:"entries"`
	DefaultPython string      `json:"defaultPython"` // which version `python` resolves to
	Conflicts     []string    `json:"conflicts"`     // human-readable conflict descriptions
	OrphanedCount int         `json:"orphanedCount"`
}

// AnalyzePATH inspects user and system PATH for Python-related entries.
func AnalyzePATH() *PathAnalysis {
	result := &PathAnalysis{}

	userEntries := getRegistryPath(registry.CURRENT_USER)
	sysEntries := getRegistryPath(registry.LOCAL_MACHINE)

	priority := 0
	seenVersion := map[string]bool{}
	firstPython := ""

	// System PATH comes first in resolution, then User PATH
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

		// Check if path exists
		if info, err := os.Stat(entry); err == nil && info.IsDir() {
			pe.Exists = true
			// Check for python.exe
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

func getPythonVersionQuick(exe string) string {
	// Import from detector package would create circular dep, so inline a quick version check
	// We read the directory name as a heuristic first
	dir := filepath.Dir(exe)
	base := filepath.Base(dir)
	// Try to extract version from directory name like "cpython-3.13.9-..."
	if strings.HasPrefix(base, "cpython-") {
		parts := strings.SplitN(base, "-", 3)
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	// Fallback: try "Python313" -> "3.13"
	if strings.HasPrefix(base, "Python") || strings.HasPrefix(base, "python") {
		numPart := strings.TrimPrefix(strings.TrimPrefix(base, "Python"), "python")
		if len(numPart) >= 2 {
			return numPart[:1] + "." + numPart[1:]
		}
	}
	return ""
}

// FindDuplicates groups installations by major.minor version and identifies duplicates.
func FindDuplicates(installs []models.PythonInstallation) []models.DuplicateGroup {
	groups := map[string][]models.PythonInstallation{}
	for _, inst := range installs {
		if inst.Source == models.SourceVenv {
			continue // don't count venvs as duplicates
		}
		groups[inst.MajorMinor] = append(groups[inst.MajorMinor], inst)
	}

	var result []models.DuplicateGroup
	// Priority: uv > official > pyenv > conda > chocolatey > scoop > store > unknown
	sourcePriority := map[models.PythonSource]int{
		models.SourceUV:         0,
		models.SourceOfficial:   1,
		models.SourcePyenv:      2,
		models.SourceConda:      3,
		models.SourceChocolatey: 4,
		models.SourceScoop:      5,
		models.SourceStore:      6,
		models.SourceUnknown:    7,
	}

	for version, installs := range groups {
		if len(installs) < 2 {
			continue
		}
		// Find the one to keep (lowest priority number = best)
		bestIdx := 0
		bestPri := 99
		for i, inst := range installs {
			if p, ok := sourcePriority[inst.Source]; ok && p < bestPri {
				bestPri = p
				bestIdx = i
			}
		}
		keep := installs[bestIdx]
		result = append(result, models.DuplicateGroup{
			Version:       version,
			Installations: installs,
			RecommendKeep: &keep,
		})
	}
	return result
}

// GenerateRecommendations creates cleanup suggestions.
func GenerateRecommendations(installs []models.PythonInstallation, dupes []models.DuplicateGroup) []models.CleanupRecommendation {
	var recs []models.CleanupRecommendation

	// Mark duplicates for removal
	keepSet := map[string]bool{}
	for _, dg := range dupes {
		if dg.RecommendKeep != nil {
			keepSet[dg.RecommendKeep.Executable] = true
		}
	}

	for _, dg := range dupes {
		for _, inst := range dg.Installations {
			if keepSet[inst.Executable] {
				continue
			}
			recs = append(recs, models.CleanupRecommendation{
				Installation: inst,
				Action:       "Uninstall",
				Reason:       "Duplicate of " + dg.Version + " (keep " + string(dg.RecommendKeep.Source) + " version)",
				Risk:         models.RiskSafe,
				SpaceSaved:   inst.SizeBytes,
			})
		}
	}

	return recs
}

// FindOrphanedVenvs finds virtual environments whose base Python no longer exists.
func FindOrphanedVenvs(installs []models.PythonInstallation) []models.PythonInstallation {
	// Collect all real Python paths
	realPythons := map[string]bool{}
	for _, inst := range installs {
		if inst.Source != models.SourceVenv {
			realPythons[strings.ToLower(filepath.Clean(inst.Path))] = true
		}
	}

	var orphans []models.PythonInstallation
	for _, inst := range installs {
		if inst.Source != models.SourceVenv {
			continue
		}
		// Read pyvenv.cfg to find the home (base Python)
		cfgPath := filepath.Join(inst.Path, "pyvenv.cfg")
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if k, v, ok := strings.Cut(line, "="); ok {
				if strings.TrimSpace(k) == "home" {
					home := strings.TrimSpace(v)
					if _, err := os.Stat(filepath.Join(home, "python.exe")); err != nil {
						inst.IsOrphaned = true
						orphans = append(orphans, inst)
					}
				}
			}
		}
	}
	return orphans
}
