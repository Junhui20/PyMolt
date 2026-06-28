//go:build windows

package detector

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// pep514Entry is one Company\Tag registration under SOFTWARE\Python (PEP 514).
type pep514Entry struct {
	Company     string
	Tag         string
	InstallPath string // directory containing python.exe
}

// enumeratePEP514 reads every Company\Tag\InstallPath under SOFTWARE\Python in
// HKLM, HKCU, and the WOW6432Node view — vendor-agnostic, per PEP 514. This is
// broader than the historical "PythonCore only" read, so distributions that
// register under their own company (incl. Microsoft's PyManager) are found.
func enumeratePEP514() []pep514Entry {
	var out []pep514Entry
	bases := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Python`},
		{registry.CURRENT_USER, `SOFTWARE\Python`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Python`},
	}
	for _, b := range bases {
		base, err := registry.OpenKey(b.root, b.path, registry.READ)
		if err != nil {
			continue
		}
		companies, err := base.ReadSubKeyNames(-1)
		base.Close()
		if err != nil {
			continue
		}
		for _, company := range companies {
			compPath := b.path + `\` + company
			comp, err := registry.OpenKey(b.root, compPath, registry.READ)
			if err != nil {
				continue
			}
			tags, err := comp.ReadSubKeyNames(-1)
			comp.Close()
			if err != nil {
				continue
			}
			for _, tag := range tags {
				ik, err := registry.OpenKey(b.root, compPath+`\`+tag+`\InstallPath`, registry.READ)
				if err != nil {
					continue
				}
				ip, _, err := ik.GetStringValue("")
				ik.Close()
				if err != nil || ip == "" {
					continue
				}
				out = append(out, pep514Entry{
					Company:     company,
					Tag:         tag,
					InstallPath: strings.TrimRight(ip, `\`),
				})
			}
		}
	}
	return out
}

// underLocalAppDataPython reports whether p is inside %LocalAppData%\Python,
// the location Microsoft's PyManager uses for its runtimes.
func underLocalAppDataPython(p string) bool {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return false
	}
	base = strings.ToLower(filepath.Clean(filepath.Join(base, "Python")))
	return strings.HasPrefix(strings.ToLower(filepath.Clean(p)), base)
}
