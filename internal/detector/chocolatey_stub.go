//go:build !windows

package detector

import "github.com/Junhui20/PyMolt/internal/models"

// ChocolateyDetector is a no-op on non-Windows platforms.
type ChocolateyDetector struct{}

func (d ChocolateyDetector) Name() string                          { return "Chocolatey" }
func (d ChocolateyDetector) Detect() []models.PythonInstallation { return nil }
