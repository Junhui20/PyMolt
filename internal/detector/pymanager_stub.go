//go:build !windows

package detector

import "github.com/Junhui20/PyMolt/internal/models"

// PyManagerDetector is a no-op on non-Windows platforms (PyManager is Windows-only).
type PyManagerDetector struct{}

func (d PyManagerDetector) Name() string { return "PyManager" }

func (d PyManagerDetector) Detect() []models.PythonInstallation { return nil }
