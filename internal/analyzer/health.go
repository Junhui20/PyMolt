package analyzer

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
)

// HealthStatus for a single Python installation.
type HealthStatus struct {
	Executable string   `json:"executable"`
	Version    string   `json:"version"`
	PythonOK   bool     `json:"pythonOk"`
	PipOK      bool     `json:"pipOk"`
	SslOK      bool     `json:"sslOk"`
	SiteOK     bool     `json:"siteOk"`
	Overall    string   `json:"overall"`
	Issues     []string `json:"issues"`
}

// CheckHealth runs diagnostics on a single Python installation.
func CheckHealth(inst models.PythonInstallation) *HealthStatus {
	hs := &HealthStatus{
		Executable: inst.Executable,
		Version:    inst.Version,
	}

	if runPythonCmd(inst.Executable, "--version") {
		hs.PythonOK = true
	} else {
		hs.Issues = append(hs.Issues, "python executable cannot run")
	}

	if !hs.PythonOK {
		hs.Overall = "Broken"
		return hs
	}

	if runPythonCmd(inst.Executable, "-m", "pip", "--version") {
		hs.PipOK = true
	} else {
		hs.Issues = append(hs.Issues, "pip is not available")
	}

	if runPythonCmd(inst.Executable, "-c", "import ssl; print(ssl.OPENSSL_VERSION)") {
		hs.SslOK = true
	} else {
		hs.Issues = append(hs.Issues, "SSL module is broken (cannot install packages from PyPI)")
	}

	if runPythonCmd(inst.Executable, "-c", "import site; print(site.getsitepackages())") {
		hs.SiteOK = true
	} else {
		hs.Issues = append(hs.Issues, "site-packages is not accessible")
	}

	switch {
	case hs.PythonOK && hs.PipOK && hs.SslOK && hs.SiteOK:
		hs.Overall = "Healthy"
	case hs.PythonOK:
		hs.Overall = "Warning"
	default:
		hs.Overall = "Broken"
	}

	return hs
}

// CheckAllHealth runs diagnostics on all installations.
func CheckAllHealth(installs []models.PythonInstallation) []*HealthStatus {
	var results []*HealthStatus
	for _, inst := range installs {
		if inst.Source == models.SourceVenv {
			if inst.Executable == "" {
				continue
			}
		}
		results = append(results, CheckHealth(inst))
	}
	return results
}

func runPythonCmd(exe string, args ...string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, args...)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
