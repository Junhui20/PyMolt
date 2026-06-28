package analyzer

import (
	"path/filepath"
	"testing"

	"github.com/Junhui20/PyMolt/internal/models"
)

func TestIsExternallyManaged(t *testing.T) {
	base := t.TempDir()
	binDir := filepath.Join(base, "bin")
	libDir := filepath.Join(base, "lib", "python3.11")
	mkdir(t, binDir) // helper defined in uninstall_test.go (same package)
	mkdir(t, libDir)

	inst := models.PythonInstallation{Path: binDir, MajorMinor: "3.11"}

	if IsExternallyManaged(inst) {
		t.Error("should not be externally managed before the marker exists")
	}

	writeFile(t, filepath.Join(libDir, "EXTERNALLY-MANAGED"), "")
	if !IsExternallyManaged(inst) {
		t.Error("should detect the EXTERNALLY-MANAGED marker")
	}

	// A typical venv (no marker) must never be flagged.
	venvBin := filepath.Join(base, "venv", "bin")
	mkdir(t, venvBin)
	venv := models.PythonInstallation{Path: venvBin, MajorMinor: "3.11"}
	if IsExternallyManaged(venv) {
		t.Error("a venv without the marker must not be flagged")
	}
}
