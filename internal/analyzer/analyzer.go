package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// PathEntry represents one entry in the PATH related to Python.
type PathEntry struct {
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	HasPython bool   `json:"hasPython"`
	Version   string `json:"version"`
	Source    string `json:"source"`
	Priority  int    `json:"priority"`
	Orphaned  bool   `json:"orphaned"`
	Shadowed  bool   `json:"shadowed"`
}

// PathAnalysis is the full PATH report.
type PathAnalysis struct {
	Entries       []PathEntry `json:"entries"`
	DefaultPython string      `json:"defaultPython"`
	Conflicts     []string    `json:"conflicts"`
	OrphanedCount int         `json:"orphanedCount"`
}

// AnalyzePATH is implemented in pathanalysis_windows.go / pathanalysis_unix.go

// getPythonVersionQuick does a fast version check without running the executable.
func getPythonVersionQuick(exe string) string {
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
			continue
		}
		groups[inst.MajorMinor] = append(groups[inst.MajorMinor], inst)
	}

	var result []models.DuplicateGroup
	sourcePriority := map[models.PythonSource]int{
		models.SourceUV:         0,
		models.SourceOfficial:   1,
		models.SourceHomebrew:   2,
		models.SourcePyenv:      3,
		models.SourceConda:      4,
		models.SourceSystem:     5,
		models.SourceChocolatey: 6,
		models.SourceScoop:      7,
		models.SourceStore:      8,
		models.SourceUnknown:    9,
	}

	for version, installs := range groups {
		if len(installs) < 2 {
			continue
		}
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
		cfgPath := filepath.Join(inst.Path, "pyvenv.cfg")
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if k, v, ok := strings.Cut(line, "="); ok {
				if strings.TrimSpace(k) == "home" {
					home := strings.TrimSpace(v)
					// Check for python executable in the home dir
					found := false
					for _, name := range []string{"python3", "python", "python.exe", "python3.exe"} {
						if _, err := os.Stat(filepath.Join(home, name)); err == nil {
							found = true
							break
						}
					}
					if !found {
						inst.IsOrphaned = true
						orphans = append(orphans, inst)
					}
				}
			}
		}
	}
	return orphans
}
