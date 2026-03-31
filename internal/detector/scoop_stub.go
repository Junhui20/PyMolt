//go:build !windows

package detector

import "github.com/Junhui20/PyMolt/internal/models"

// ScoopDetector is a no-op on non-Windows platforms.
type ScoopDetector struct{}

func (d ScoopDetector) Name() string                          { return "Scoop" }
func (d ScoopDetector) Detect() []models.PythonInstallation { return nil }
