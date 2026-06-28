package analyzer

import (
	"os"
	"path/filepath"

	"github.com/Junhui20/PyMolt/internal/models"
)

// IsExternallyManaged reports whether this interpreter ships a PEP 668
// "EXTERNALLY-MANAGED" marker, meaning a plain `pip install` into it is blocked
// (Debian/Ubuntu system Python, Homebrew Python, etc.). It checks the standard
// stdlib locations relative to the interpreter instead of running it, so it adds
// no subprocess to a scan. Best-effort: unusual layouts may be missed (a false
// negative just means no badge), but false positives are effectively impossible
// since the marker file only exists when a distributor places it.
func IsExternallyManaged(inst models.PythonInstallation) bool {
	if inst.MajorMinor == "" || inst.Path == "" {
		return false
	}
	pyDir := "python" + inst.MajorMinor // e.g. "python3.11"
	prefix := filepath.Dir(inst.Path)   // e.g. "/usr" from "/usr/bin"
	const marker = "EXTERNALLY-MANAGED"
	for _, c := range []string{
		filepath.Join(prefix, "lib", pyDir, marker),
		filepath.Join(prefix, "lib64", pyDir, marker),
		filepath.Join(inst.Path, "lib", pyDir, marker),
		filepath.Join(prefix, "lib", marker),
	} {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return true
		}
	}
	return false
}
