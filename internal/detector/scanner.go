package detector

import (
	"path/filepath"
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

	// Deduplicate by executable path
	seen := make(map[string]bool)
	var unique []models.PythonInstallation
	for _, inst := range all {
		key := strings.ToLower(filepath.Clean(inst.Executable))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, inst)
	}

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
