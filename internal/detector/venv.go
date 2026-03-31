package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// VenvDetector finds virtual environments.
type VenvDetector struct {
	SearchDirs []string // directories to scan for venvs
	MaxDepth   int      // how deep to search (default 4)
}

func (d VenvDetector) Name() string { return "Virtual Environments" }

func (d VenvDetector) Detect() []models.PythonInstallation {
	var results []models.PythonInstallation

	maxDepth := d.MaxDepth
	if maxDepth == 0 {
		maxDepth = 4
	}

	searchDirs := d.SearchDirs
	if len(searchDirs) == 0 {
		home := os.Getenv("USERPROFILE")
		searchDirs = []string{
			filepath.Join(home, "Documents"),
			filepath.Join(home, "Projects"),
			filepath.Join(home, "repos"),
			filepath.Join(home, "dev"),
			filepath.Join(home, "Desktop"),
		}
	}

	for _, root := range searchDirs {
		d.scanDir(root, 0, maxDepth, &results)
	}
	return results
}

func (d VenvDetector) scanDir(dir string, depth, maxDepth int, results *[]models.PythonInstallation) {
	if depth > maxDepth {
		return
	}

	cfgPath := filepath.Join(dir, "pyvenv.cfg")
	if _, err := os.Stat(cfgPath); err == nil {
		inst := d.parseVenvConfig(dir, cfgPath)
		if inst != nil {
			*results = append(*results, *inst)
		}
		return // don't recurse into a venv
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip common non-project dirs
		if name == "node_modules" || name == "__pycache__" || name == ".git" || name == "AppData" {
			continue
		}
		d.scanDir(filepath.Join(dir, name), depth+1, maxDepth, results)
	}
}

func (d VenvDetector) parseVenvConfig(venvDir, cfgPath string) *models.PythonInstallation {
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var version, home string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if k, v, ok := strings.Cut(line, "="); ok {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			switch k {
			case "version":
				version = v
			case "home":
				home = v
			}
		}
	}

	if version == "" {
		return nil
	}

	// Check if base Python still exists
	baseExists := false
	if home != "" {
		if _, err := os.Stat(filepath.Join(home, "python.exe")); err == nil {
			baseExists = true
		}
	}

	_ = baseExists // will be used for orphan detection in analyzer

	return &models.PythonInstallation{
		Version:    version,
		MajorMinor: ExtractMajorMinor(version),
		Path:       venvDir,
		Executable: FindExecutable(filepath.Join(venvDir, "Scripts")),
		Source:     models.SourceVenv,
		SizeBytes:  DirSize(venvDir),
		InPath:     false,
	}
}
