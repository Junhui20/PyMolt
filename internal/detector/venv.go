package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Junhui20/PyMolt/internal/config"
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

	cfg := config.Load()

	maxDepth := d.MaxDepth
	if maxDepth == 0 {
		maxDepth = 4
	}

	searchDirs := d.SearchDirs
	if len(searchDirs) == 0 {
		if cfg.FullHomeScan {
			searchDirs = []string{HomeDir()}
			maxDepth = 6
		} else {
			searchDirs = defaultVenvSearchDirs()
		}
	}

	// Always include centralized venv storage locations (Poetry, Pipenv, etc.)
	searchDirs = append(searchDirs, centralVenvDirs()...)

	// Append user-configured custom paths.
	searchDirs = append(searchDirs, cfg.VenvScanPaths...)

	// Deduplicate (a custom path may overlap with defaults).
	searchDirs = dedupPaths(searchDirs)

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

// centralVenvDirs returns directories where tools store venvs centrally
// (not inside user projects). These are always scanned regardless of mode.
func centralVenvDirs() []string {
	home := HomeDir()
	var dirs []string

	// Pipenv / virtualenvwrapper: ~/.virtualenvs or WORKON_HOME
	if wh := os.Getenv("WORKON_HOME"); wh != "" {
		dirs = append(dirs, wh)
	}
	dirs = append(dirs, filepath.Join(home, ".virtualenvs"))

	switch runtime.GOOS {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")  // C:\Users\X\AppData\Local
		roaming := os.Getenv("APPDATA")      // C:\Users\X\AppData\Roaming

		// Poetry
		if local != "" {
			dirs = append(dirs, filepath.Join(local, "pypoetry", "Cache", "virtualenvs"))
		}
		if roaming != "" {
			dirs = append(dirs, filepath.Join(roaming, "pypoetry", "virtualenvs"))
		}
		// Hatch
		if local != "" {
			dirs = append(dirs, filepath.Join(local, "hatch", "env", "virtual"))
		}
		// PDM
		if local != "" {
			dirs = append(dirs, filepath.Join(local, "pdm", "venvs"))
		}
		// Pipenv (Windows)
		if local != "" {
			dirs = append(dirs, filepath.Join(local, "pipenv", "virtualenvs"))
		}

	case "darwin":
		// Poetry
		dirs = append(dirs, filepath.Join(home, "Library", "Caches", "pypoetry", "virtualenvs"))
		// Hatch
		dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "hatch", "env", "virtual"))
		// PDM
		dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "pdm", "venvs"))

	default: // Linux
		cacheDir := os.Getenv("XDG_CACHE_HOME")
		if cacheDir == "" {
			cacheDir = filepath.Join(home, ".cache")
		}
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(home, ".local", "share")
		}

		// Poetry
		dirs = append(dirs, filepath.Join(cacheDir, "pypoetry", "virtualenvs"))
		// Hatch
		dirs = append(dirs, filepath.Join(dataDir, "hatch", "env", "virtual"))
		// PDM
		dirs = append(dirs, filepath.Join(dataDir, "pdm", "venvs"))
		// Pipenv (Linux)
		dirs = append(dirs, filepath.Join(dataDir, "virtualenvs"))
	}

	return dirs
}

// skipDirs contains directory names to always skip during scanning.
var skipDirs = map[string]bool{
	"node_modules": true,
	"__pycache__":  true,
	".git":         true,
	"AppData":      true,
	"Library":      true,
	".cache":       true,
	".local":       true,
	".npm":         true,
	".cargo":       true,
	".rustup":      true,
	"go":           true,
	".gradle":      true,
	".m2":          true,
	".nuget":       true,
	".docker":      true,
	".vscode":      true,
	"OneDrive":     true,
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
		if skipDirs[name] {
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
			case "version_info":
				// uv uses "version_info = 3.11.14"
				// virtualenv uses "version_info = 3.11.14.final.0"
				if version == "" {
					version = cleanVersionInfo(v)
				}
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

// cleanVersionInfo normalizes version_info values.
// "3.11.14" → "3.11.14"
// "3.11.14.final.0" → "3.11.14" (virtualenv format)
func cleanVersionInfo(v string) string {
	parts := strings.SplitN(v, ".", 4)
	if len(parts) >= 4 {
		// Check if 4th part is non-numeric (e.g. "final")
		if len(parts[3]) > 0 && (parts[3][0] < '0' || parts[3][0] > '9') {
			return parts[0] + "." + parts[1] + "." + parts[2]
		}
	}
	return v
}

// dedupPaths removes duplicate paths (case-insensitive on Windows).
func dedupPaths(paths []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, p := range paths {
		key := filepath.Clean(p)
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}
