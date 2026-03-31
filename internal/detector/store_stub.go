//go:build !windows

package detector

import "github.com/Junhui20/PyMolt/internal/models"

// StoreDetector is a no-op on non-Windows platforms.
type StoreDetector struct{}

func (d StoreDetector) Name() string                          { return "Microsoft Store" }
func (d StoreDetector) Detect() []models.PythonInstallation { return nil }
