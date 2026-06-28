package detector

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
)

// Scanner orchestrates all detectors and produces a unified result.
type Scanner struct {
	detectors []Detector
	OnStatus  func(msg string)
}

// NewScanner creates a scanner with all built-in detectors.
func NewScanner() *Scanner {
	return &Scanner{
		detectors: []Detector{
			OfficialDetector{},
			PyManagerDetector{},
			HomebrewDetector{},
			UVDetector{},
			PyenvDetector{},
			CondaDetector{},
			ChocolateyDetector{},
			ScoopDetector{},
			StoreDetector{},
			VenvDetector{},
			AsdfDetector{},
			MiseDetector{},
			PipxDetector{},
			IDEDetector{},
			WhichDetector{}, // last: catches anything others missed
		},
	}
}

// Scan runs all detectors and returns merged results.
func (s *Scanner) Scan() *models.ScanResult {
	start := time.Now()

	var mu sync.Mutex
	var all []models.PythonInstallation
	var wg sync.WaitGroup

	for _, det := range s.detectors {
		wg.Add(1)
		go func(d Detector) {
			defer wg.Done()
			found := d.Detect()
			mu.Lock()
			all = append(all, found...)
			mu.Unlock()
		}(det)
	}

	wg.Wait()

	unique := dedupeInstallations(all)

	// Mark default Python
	unique = markDefault(unique)

	// Calculate total size
	var totalSize int64
	for _, inst := range unique {
		totalSize += inst.SizeBytes
	}

	return &models.ScanResult{
		Installations:  unique,
		TotalSize:      totalSize,
		ScanDurationMs: time.Since(start).Milliseconds(),
	}
}

// dedupeInstallations collapses entries that refer to the same physical
// interpreter. Resolving symlinks is essential: on most Linux systems /bin is a
// symlink to /usr/bin and python3 a symlink to python3.X, so /usr/bin/python3,
// /bin/python3 and /usr/bin/python3.14 are the SAME file reached three ways.
// Keying on the literal path counted them as three installs and produced a bogus
// "duplicate Python — remove the rest" recommendation that, if followed, would
// damage the system interpreter. Virtual environments are keyed by their own
// directory instead, because a venv's interpreter symlinks back to its base and
// must stay distinct.
func dedupeInstallations(all []models.PythonInstallation) []models.PythonInstallation {
	seen := make(map[string]int) // dedupe key -> index into unique
	var unique []models.PythonInstallation
	for _, inst := range all {
		key := dedupeKey(inst)
		if key == "" {
			continue
		}
		if idx, ok := seen[key]; ok {
			// Same physical interpreter already recorded. Keep whichever entry has
			// the more specific source (a named installer / System beats the generic
			// PATH-discovery "Unknown"), and carry over identifying flags.
			if sourceRank(inst.Source) > sourceRank(unique[idx].Source) {
				inst.IsDefault = inst.IsDefault || unique[idx].IsDefault
				inst.InPath = inst.InPath || unique[idx].InPath
				unique[idx] = inst
			} else {
				unique[idx].IsDefault = unique[idx].IsDefault || inst.IsDefault
				unique[idx].InPath = unique[idx].InPath || inst.InPath
			}
			continue
		}
		seen[key] = len(unique)
		unique = append(unique, inst)
	}
	return unique
}

// dedupeKey returns a stable identity for an installation. Real interpreters are
// identified by their canonical (symlink-resolved) executable so the same file
// reached by several paths counts once; venvs are identified by their directory
// so they are never folded into the base interpreter they link to.
func dedupeKey(inst models.PythonInstallation) string {
	if inst.Source == models.SourceVenv {
		return "venv:" + canonicalPath(inst.Path)
	}
	return canonicalPath(inst.Executable)
}

// canonicalPath resolves symlinks to a real filesystem path so equivalent paths
// compare equal. It falls back to a lexical clean when the path can't be
// resolved (e.g. it no longer exists), and lower-cases on case-insensitive
// platforms (Windows).
func canonicalPath(path string) string {
	if path == "" {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = filepath.Clean(path)
	}
	if runtime.GOOS == "windows" {
		resolved = strings.ToLower(resolved)
	}
	return resolved
}

// sourceRank ranks how specific/trustworthy a source label is, so that when two
// detectors report the same physical interpreter the more informative label
// wins. Generic PATH discovery ("Unknown") is lowest.
func sourceRank(s models.PythonSource) int {
	switch s {
	case models.SourceUnknown:
		return 0
	case models.SourceSystem:
		return 1
	default:
		return 2 // a specific installer/manager (Official, uv, pyenv, venv, ...)
	}
}

func markDefault(installs []models.PythonInstallation) []models.PythonInstallation {
	defaultVersion := GetPythonVersion("python3")
	if defaultVersion == "" {
		defaultVersion = GetPythonVersion("python")
	}
	if defaultVersion == "" {
		return installs
	}
	for i := range installs {
		if installs[i].Version == defaultVersion && installs[i].InPath {
			installs[i].IsDefault = true
			break
		}
	}
	return installs
}
