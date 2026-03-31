package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// VenvDetector finds virtual environments.
type VenvDetector struct {
	SearchDirs []string
	MaxDepth   int
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
		searchDirs = defaultVenvSearchDirs()
	}

	for _, root := range searchDirs {
		d.scanDir(root, 0, maxDepth, &results)
	}
	return results
}

func defaultVenvSearchDirs() []string {
	home := HomeDir()
	dirs := []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Projects"),
		filepath.Join(home, "repos"),
		filepath.Join(home, "dev"),
		filepath.Join(home, "Desktop"),
	}

	if runtime.GOOS != "windows" {
		dirs = append(dirs,
			filepath.Join(home, "projects"),
			filepath.Join(home, "src"),
			filepath.Join(home, "code"),
		)
	}

	return dirs
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
		if name == "node_modules" || name == "__pycache__" || name == ".git" || name == "AppData" || name == "Library" {
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
	if home != "" {
		for _, name := range pythonExeNames() {
			if _, err := os.Stat(filepath.Join(home, name)); err == nil {
				break
			}
		}
	}

	scriptsDir := filepath.Join(venvDir, venvScriptsDir())

	return &models.PythonInstallation{
		Version:    version,
		MajorMinor: ExtractMajorMinor(version),
		Path:       venvDir,
		Executable: FindExecutable(scriptsDir),
		Source:     models.SourceVenv,
		SizeBytes:  DirSize(venvDir),
		InPath:     false,
	}
}
