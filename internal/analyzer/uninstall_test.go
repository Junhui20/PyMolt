package analyzer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Junhui20/PyMolt/internal/models"
)

func TestUnsafeToDelete_RejectsEmptyAndRelative(t *testing.T) {
	if unsafeToDelete("") == "" {
		t.Error(`unsafeToDelete("") should refuse an empty path`)
	}
	if unsafeToDelete(filepath.Join("relative", "venv")) == "" {
		t.Error("unsafeToDelete should refuse a non-absolute path")
	}
}

func TestUnsafeToDelete_RejectsProtectedAndHome(t *testing.T) {
	var protected string
	if runtime.GOOS == "windows" {
		protected = `C:\Windows`
	} else {
		protected = "/usr"
	}
	if unsafeToDelete(protected) == "" {
		t.Errorf("unsafeToDelete(%q) should refuse a protected system directory", protected)
	}

	if home := userHome(); home != "" {
		if unsafeToDelete(home) == "" {
			t.Errorf("unsafeToDelete(%q) should refuse the user's home directory", home)
		}
	}
}

func TestUnsafeToDelete_AllowsDeepNonProtected(t *testing.T) {
	var p string
	if runtime.GOOS == "windows" {
		p = `C:\Users\someone\projects\app\.venv`
	} else {
		p = "/home/someone/projects/app/.venv"
	}
	if reason := unsafeToDelete(p); reason != "" {
		t.Errorf("unsafeToDelete(%q) unexpectedly refused: %s", p, reason)
	}
}

func TestLooksLikePythonDir(t *testing.T) {
	base := t.TempDir()

	// Empty directory: not a Python dir.
	empty := filepath.Join(base, "empty")
	mkdir(t, empty)
	if looksLikePythonDir(empty) {
		t.Error("an empty directory must not look like a Python dir")
	}

	// A folder with ONLY a planted pyvenv.cfg (the attack vector): not enough.
	planted := filepath.Join(base, "documents")
	mkdir(t, planted)
	writeFile(t, filepath.Join(planted, "pyvenv.cfg"), "version = 3.11\n")
	if looksLikePythonDir(planted) {
		t.Error("a folder with only a stray pyvenv.cfg must not be treated as a venv")
	}

	// A real venv layout: pyvenv.cfg + a bin/ (POSIX) or Scripts/ (Windows) dir.
	venv := filepath.Join(base, "venv")
	mkdir(t, venv)
	writeFile(t, filepath.Join(venv, "pyvenv.cfg"), "version = 3.11\n")
	scriptsDir := "bin"
	if runtime.GOOS == "windows" {
		scriptsDir = "Scripts"
	}
	mkdir(t, filepath.Join(venv, scriptsDir))
	if !looksLikePythonDir(venv) {
		t.Error("a pyvenv.cfg alongside a scripts directory should look like a venv")
	}

	// A directory containing a real interpreter binary.
	install := filepath.Join(base, "py")
	exeRel := filepath.Join("bin", "python3")
	if runtime.GOOS == "windows" {
		exeRel = "python.exe"
	}
	exe := filepath.Join(install, exeRel)
	mkdir(t, filepath.Dir(exe))
	writeFile(t, exe, "")
	if !looksLikePythonDir(install) {
		t.Error("a directory containing a python interpreter should look like a Python dir")
	}
}

// TestUninstall_DeletesWindowsStyleOrphanedVenv covers a real case: a venv created
// on Windows (Scripts/ layout, base interpreter under a C:\... path) synced onto a
// Linux machine, where it is orphaned and has no usable interpreter. The "Delete
// venv" auto-fix must still remove it — looksLikePythonDir accepts pyvenv.cfg next
// to a Scripts/ directory regardless of host OS.
func TestUninstall_DeletesWindowsStyleOrphanedVenv(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "project", ".venv311")
	mkdir(t, filepath.Join(dir, "Scripts"))
	mkdir(t, filepath.Join(dir, "Lib"))
	writeFile(t, filepath.Join(dir, "pyvenv.cfg"),
		"home = C:\\Users\\Someone\\AppData\\Roaming\\uv\\python\\cpython-3.11.14-windows-x86_64-none\n")

	res := Uninstall(models.PythonInstallation{Source: models.SourceVenv, Path: dir, SizeBytes: 1234})
	if !res.Success {
		t.Fatalf("expected delete to succeed, got failure: %s", res.Message)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, stat err = %v", dir, err)
	}
}

func mkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
}

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}
