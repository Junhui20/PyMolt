package detector

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Junhui20/PyMolt/internal/models"
)

// TestDedupeInstallationsSymlinks reproduces the real-world layout that caused a
// single system Python to be reported three times: /bin is a symlink to
// /usr/bin, and python3 is a symlink to python3.X. Three different paths, one
// real interpreter — dedup must collapse them to one and keep the most specific
// source.
func TestDedupeInstallationsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX symlink layout; not applicable on Windows")
	}

	root := t.TempDir()
	usrbin := filepath.Join(root, "usr", "bin")
	if err := os.MkdirAll(usrbin, 0o755); err != nil {
		t.Fatal(err)
	}
	// Real interpreter file, plus a python3 -> python3.14 symlink next to it.
	realExe := filepath.Join(usrbin, "python3.14")
	if err := os.WriteFile(realExe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("python3.14", filepath.Join(usrbin, "python3")); err != nil {
		t.Fatal(err)
	}
	// bin -> usr/bin, mimicking the usrmerge layout on modern distros.
	if err := os.Symlink(filepath.Join(root, "usr", "bin"), filepath.Join(root, "bin")); err != nil {
		t.Fatal(err)
	}

	all := []models.PythonInstallation{
		// PATH discovery reaches it via /bin/python3 (symlinked dir + symlinked name).
		{Source: models.SourceUnknown, Version: "3.14.4", MajorMinor: "3.14",
			Path: filepath.Join(root, "bin"), Executable: filepath.Join(root, "bin", "python3"), IsDefault: true, InPath: true},
		// PATH discovery also reaches it via /usr/bin/python3.
		{Source: models.SourceUnknown, Version: "3.14.4", MajorMinor: "3.14",
			Path: usrbin, Executable: filepath.Join(usrbin, "python3"), InPath: true},
		// The system detector reports the versioned binary directly.
		{Source: models.SourceSystem, Version: "3.14.4", MajorMinor: "3.14",
			Path: usrbin, Executable: realExe},
	}

	got := dedupeInstallations(all)

	if len(got) != 1 {
		t.Fatalf("expected 1 installation after dedup, got %d: %+v", len(got), got)
	}
	if got[0].Source != models.SourceSystem {
		t.Errorf("expected the System source to win over Unknown, got %q", got[0].Source)
	}
	if !got[0].IsDefault {
		t.Error("expected IsDefault to be carried over from the deduped entry")
	}
	if !got[0].InPath {
		t.Error("expected InPath to be carried over from the deduped entry")
	}
}

// TestDedupeInstallationsKeepsVenvs ensures a virtual environment is never folded
// into the base interpreter it symlinks to — they are distinct installations.
func TestDedupeInstallationsKeepsVenvs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX symlink layout; not applicable on Windows")
	}

	root := t.TempDir()
	base := filepath.Join(root, "usr", "bin")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	baseExe := filepath.Join(base, "python3.14")
	if err := os.WriteFile(baseExe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// A venv whose interpreter symlinks back to the base interpreter.
	venvBin := filepath.Join(root, "proj", ".venv", "bin")
	if err := os.MkdirAll(venvBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(baseExe, filepath.Join(venvBin, "python3")); err != nil {
		t.Fatal(err)
	}

	all := []models.PythonInstallation{
		{Source: models.SourceSystem, Version: "3.14.4", MajorMinor: "3.14", Path: base, Executable: baseExe},
		{Source: models.SourceVenv, Version: "3.14.4", MajorMinor: "3.14",
			Path: filepath.Join(root, "proj", ".venv"), Executable: filepath.Join(venvBin, "python3")},
	}

	got := dedupeInstallations(all)

	if len(got) != 2 {
		t.Fatalf("expected base + venv to remain distinct (2), got %d: %+v", len(got), got)
	}
}
