package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Junhui20/PyMolt/internal/models"
)

func TestScanProjectPins(t *testing.T) {
	root := t.TempDir()

	mkProj := func(sub, file, content string) {
		d := filepath.Join(root, sub)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, file), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mkProj("projA", ".python-version", "3.11.4\n")
	mkProj("projB", ".tool-versions", "nodejs 20.0.0\npython 3.9.1\n")
	mkProj("projC", "mise.toml", "[tools]\npython = \"3.12\"\n")
	mkProj(filepath.Join("node_modules", "dep"), ".python-version", "2.7\n")   // must be pruned
	mkProj(filepath.Join("a", "b", "c", "d", "e"), ".python-version", "2.6\n") // beyond max depth

	installs := []models.PythonInstallation{
		{Version: "3.11.4", MajorMinor: "3.11", Executable: "/usr/bin/python3.11"},
		{Version: "3.12.1", MajorMinor: "3.12", Executable: "/usr/bin/python3.12"},
	}
	pins := ScanProjectPins([]string{root}, installs)

	byBase := map[string]ProjectPin{}
	for _, p := range pins {
		byBase[filepath.Base(p.Dir)] = p
	}

	if p, ok := byBase["projA"]; !ok || !p.Installed || p.ResolvedExe == "" {
		t.Errorf("projA (.python-version 3.11.4) should resolve to an install: %+v (ok=%v)", p, ok)
	}
	if p, ok := byBase["projB"]; !ok || p.Pinned != "3.9.1" || p.Installed {
		t.Errorf("projB (3.9.1) should be found and unsatisfied: %+v (ok=%v)", p, ok)
	}
	if p, ok := byBase["projC"]; !ok || p.Pinned != "3.12" || !p.Installed {
		t.Errorf("projC (mise 3.12) should resolve via major.minor: %+v (ok=%v)", p, ok)
	}
	for _, p := range pins {
		if p.Pinned == "2.7" {
			t.Error("a pin inside node_modules should have been pruned")
		}
		if p.Pinned == "2.6" {
			t.Error("a pin beyond the max scan depth should be ignored")
		}
	}
}
