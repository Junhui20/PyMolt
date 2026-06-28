package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// UninstallResult reports what happened.
type UninstallResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	SpaceFreed int64  `json:"spaceFreed"`
}

// Uninstall removes a Python installation based on its source.
func Uninstall(inst models.PythonInstallation) *UninstallResult {
	switch inst.Source {
	case models.SourceVenv:
		return deleteDirectory(inst)
	case models.SourceUV:
		return uninstallUV(inst)
	case models.SourcePyenv:
		return deleteDirectory(inst)
	case models.SourceOfficial:
		return uninstallOfficial(inst)
	case models.SourceChocolatey:
		return uninstallChocolatey(inst)
	case models.SourceScoop:
		return uninstallScoop(inst)
	case models.SourceHomebrew:
		return uninstallHomebrew(inst)
	case models.SourcePyManager:
		return uninstallPyManager(inst)
	default:
		return deleteDirectory(inst)
	}
}

// deleteDirectory removes an installation directory after verifying it is not a
// protected system or home location. The path comes from a prior scan (not raw
// user input), but this guard is the last line of defense for the app's most
// destructive operation, so it validates explicitly rather than trusting input.
func deleteDirectory(inst models.PythonInstallation) *UninstallResult {
	if reason := unsafeToDelete(inst.Path); reason != "" {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Refusing to delete %s: %s", inst.Path, reason)}
	}

	// Defense in depth: only delete a directory that actually contains a Python
	// interpreter (or a venv layout). This stops a mis-detected or planted folder
	// — e.g. a Documents folder with a stray pyvenv.cfg — from ever being wiped.
	if !looksLikePythonDir(inst.Path) {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Refusing to delete %s: it does not contain a Python interpreter", inst.Path)}
	}

	size := inst.SizeBytes
	if err := os.RemoveAll(inst.Path); err != nil {
		return &UninstallResult{Success: false, Message: fmt.Sprintf("Failed to delete %s: %v", inst.Path, err)}
	}
	return &UninstallResult{Success: true, Message: fmt.Sprintf("Deleted %s", inst.Path), SpaceFreed: size}
}

// unsafeToDelete returns a human-readable reason if the path must not be deleted,
// or "" if deletion is allowed. It refuses non-absolute paths, the filesystem
// root, protected system directories, the user's home directory, and any
// directory that is an ancestor of one of those (deleting which would also wipe
// the protected location).
func unsafeToDelete(path string) string {
	if path == "" {
		return "empty path"
	}
	if !filepath.IsAbs(path) {
		return "path is not absolute"
	}

	target := normalizePath(path)

	// Block the filesystem/volume root itself ("/" or e.g. "C:\").
	if target == normalizePath(filepath.VolumeName(path)+string(filepath.Separator)) {
		return "this is the filesystem root"
	}

	protected := platformProtectedPaths()
	if home := userHome(); home != "" {
		protected = append(protected, home)
	}

	for _, p := range protected {
		if p == "" {
			continue
		}
		np := normalizePath(p)
		if target == np {
			return "this is a protected directory"
		}
		// Refuse if target is an ancestor of a protected path.
		if strings.HasPrefix(np, target+string(filepath.Separator)) {
			return "deleting this would remove a protected directory"
		}
	}

	return ""
}

// normalizePath cleans a path and lower-cases it on case-insensitive platforms
// (Windows) so comparisons are reliable.
func normalizePath(p string) string {
	p = filepath.Clean(p)
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}
	return p
}

// userHome returns the user's home directory, or "" if it cannot be determined.
func userHome() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

// looksLikePythonDir reports whether path has the hallmarks of a Python
// installation or virtual environment: an interpreter binary, or a pyvenv.cfg
// next to a bin/ (POSIX) or Scripts/ (Windows) directory. A lone pyvenv.cfg is
// deliberately NOT enough, so a planted file can't turn an arbitrary folder into
// a deletable "venv".
func looksLikePythonDir(path string) bool {
	for _, rel := range []string{
		filepath.Join("bin", "python"),
		filepath.Join("bin", "python3"),
		filepath.Join("Scripts", "python.exe"),
		"python.exe",
		"python3.exe",
	} {
		if fi, err := os.Stat(filepath.Join(path, rel)); err == nil && !fi.IsDir() {
			return true
		}
	}
	if fi, err := os.Stat(filepath.Join(path, "pyvenv.cfg")); err == nil && !fi.IsDir() {
		for _, d := range []string{"bin", "Scripts"} {
			if di, err := os.Stat(filepath.Join(path, d)); err == nil && di.IsDir() {
				return true
			}
		}
	}
	return false
}

// uninstallUV, uninstallOfficial, uninstallChocolatey, uninstallScoop, uninstallHomebrew,
// uninstallPyManager, platformProtectedPaths, and the cache functions
// are defined in uninstall_windows.go / uninstall_unix.go
