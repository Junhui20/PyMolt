//go:build !darwin

package detector

import "github.com/Junhui20/PyMolt/internal/models"

// HomebrewDetector is a no-op on non-macOS platforms.
type HomebrewDetector struct{}

func (d HomebrewDetector) Name() string                          { return "Homebrew" }
func (d HomebrewDetector) Detect() []models.PythonInstallation { return nil }
