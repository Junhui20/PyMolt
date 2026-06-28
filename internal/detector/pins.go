package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// ProjectPin is a Python version pin found in a project directory.
type ProjectPin struct {
	Dir         string `json:"dir"`
	File        string `json:"file"`        // ".python-version", ".tool-versions", "mise.toml"
	Pinned      string `json:"pinned"`      // e.g. "3.11.4" or "3.11"
	ResolvedExe string `json:"resolvedExe"` // matching interpreter, or ""
	Installed   bool   `json:"installed"`
}

const pinScanMaxDepth = 4

// ScanProjectPins walks roots for Python version pins (.python-version,
// .tool-versions, mise.toml) and resolves each against the detected installs.
// When roots is empty it falls back to a few common project locations.
func ScanProjectPins(roots []string, installs []models.PythonInstallation) []ProjectPin {
	if len(roots) == 0 {
		roots = defaultPinRoots()
	}

	var pins []ProjectPin
	seenRoot := map[string]bool{}
	for _, root := range roots {
		if root == "" {
			continue
		}
		root = filepath.Clean(root)
		if seenRoot[root] {
			continue
		}
		seenRoot[root] = true
		walkPins(root, 0, installs, &pins)
	}

	// Dedup by directory + file in case roots overlap.
	seenPin := map[string]bool{}
	var uniq []ProjectPin
	for _, p := range pins {
		k := p.Dir + "|" + p.File
		if !seenPin[k] {
			seenPin[k] = true
			uniq = append(uniq, p)
		}
	}
	return uniq
}

func defaultPinRoots() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	var roots []string
	for _, d := range []string{"Projects", "projects", "repos", "dev", "Code", "src", "workspace"} {
		roots = append(roots, filepath.Join(home, d))
	}
	return roots
}

func walkPins(dir string, depth int, installs []models.PythonInstallation, out *[]ProjectPin) {
	if depth > pinScanMaxDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			if !skipDirs[e.Name()] {
				walkPins(filepath.Join(dir, e.Name()), depth+1, installs, out)
			}
			continue
		}
		if pin := parsePinFile(dir, e.Name()); pin != nil {
			resolvePin(pin, installs)
			*out = append(*out, *pin)
		}
	}
}

func parsePinFile(dir, name string) *ProjectPin {
	var version string
	switch name {
	case ".python-version":
		version = firstPinLine(filepath.Join(dir, name))
	case ".tool-versions":
		version = toolVersionsPython(filepath.Join(dir, name))
	case "mise.toml", ".mise.toml":
		version = miseTomlPython(filepath.Join(dir, name))
	default:
		return nil
	}
	if version == "" {
		return nil
	}
	return &ProjectPin{Dir: dir, File: name, Pinned: version}
}

func firstPinLine(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}

func toolVersionsPython(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) >= 2 && fields[0] == "python" {
			return fields[1]
		}
	}
	return ""
}

func miseTomlPython(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		i := strings.Index(line, "=")
		if i < 0 || strings.TrimSpace(line[:i]) != "python" {
			continue
		}
		val := strings.TrimSpace(line[i+1:])
		val = strings.Trim(val, "[]")
		val = strings.Split(val, ",")[0]
		val = strings.Trim(val, "\"' ")
		if val != "" {
			return val
		}
	}
	return ""
}

func resolvePin(pin *ProjectPin, installs []models.PythonInstallation) {
	want := strings.TrimPrefix(pin.Pinned, "python-")
	for i := range installs {
		if installs[i].Version == want {
			pin.Installed = true
			pin.ResolvedExe = installs[i].Executable
			return
		}
	}
	if mm := ExtractMajorMinor(want); mm != "" {
		for i := range installs {
			if installs[i].MajorMinor == mm {
				pin.Installed = true
				pin.ResolvedExe = installs[i].Executable
				return
			}
		}
	}
}
