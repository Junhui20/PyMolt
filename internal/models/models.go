package models

import (
	"fmt"
	"path/filepath"
)

// PythonSource identifies how a Python installation was installed.
type PythonSource string

const (
	SourceOfficial   PythonSource = "Official Installer"
	SourcePyenv      PythonSource = "pyenv-win"
	SourceConda      PythonSource = "Conda/Miniconda"
	SourceUV         PythonSource = "uv"
	SourceChocolatey PythonSource = "Chocolatey"
	SourceScoop      PythonSource = "Scoop"
	SourceStore      PythonSource = "Microsoft Store"
	SourceVenv       PythonSource = "Virtual Environment"
	SourceUnknown    PythonSource = "Unknown"
)

// RiskLevel indicates how safe a cleanup action is.
type RiskLevel string

const (
	RiskSafe      RiskLevel = "Safe"
	RiskCaution   RiskLevel = "Caution"
	RiskDangerous RiskLevel = "Dangerous"
)

// PythonInstallation represents a single Python found on the system.
type PythonInstallation struct {
	Version      string       // e.g. "3.13.9"
	MajorMinor   string       // e.g. "3.13"
	Path         string       // directory containing python.exe
	Executable   string       // full path to python.exe
	Source       PythonSource // how it was installed
	SizeBytes    int64        // disk usage
	InPath       bool         // present in system/user PATH
	IsDefault    bool         // the one invoked by bare `python` command
	Architecture string       // "64-bit" or "32-bit"
	IsOrphaned   bool         // venv whose base Python no longer exists
}

// DisplaySize returns a human-readable size string.
func (p PythonInstallation) DisplaySize() string {
	return FormatSize(p.SizeBytes)
}

// BaseName returns the directory name (e.g. "Python313").
func (p PythonInstallation) BaseName() string {
	return filepath.Base(p.Path)
}

// DuplicateGroup is a set of installations with the same major.minor version.
type DuplicateGroup struct {
	Version       string                // e.g. "3.13"
	Installations []PythonInstallation
	RecommendKeep *PythonInstallation   // which one to keep
}

// CleanupRecommendation is a suggested action for one installation.
type CleanupRecommendation struct {
	Installation PythonInstallation
	Action       string    // "Uninstall", "Delete", "Remove from PATH", "Keep"
	Reason       string
	Risk         RiskLevel
	SpaceSaved   int64
}

// ScanResult holds everything found during a scan.
type ScanResult struct {
	Installations   []PythonInstallation
	Duplicates      []DuplicateGroup
	Recommendations []CleanupRecommendation
	TotalSize       int64
	ReclaimableSize int64
	ScanDurationMs  int64
}

// FormatSize converts bytes to a human-readable string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
