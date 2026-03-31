package detector

import "github.com/Junhui20/PyMolt/internal/models"

// Detector finds Python installations from a specific source.
type Detector interface {
	Name() string
	Detect() []models.PythonInstallation
}
