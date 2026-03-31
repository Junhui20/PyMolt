package internal

import (
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Junhui20/PyMolt/internal/analyzer"
	"github.com/Junhui20/PyMolt/internal/detector"
	"github.com/Junhui20/PyMolt/internal/models"
)

// App is the main backend bound to Wails frontend.
type App struct {
	mu       sync.RWMutex
	lastScan *models.ScanResult
}

func NewApp() *App { return &App{} }

// --- Scan ---

type ScanResult struct {
	Installations   []models.PythonInstallation   `json:"installations"`
	Duplicates      []models.DuplicateGroup        `json:"duplicates"`
	Recommendations []models.CleanupRecommendation `json:"recommendations"`
	OrphanedVenvs   []models.PythonInstallation    `json:"orphanedVenvs"`
	TotalSize       int64                          `json:"totalSize"`
	ReclaimableSize int64                          `json:"reclaimableSize"`
	ScanDurationMs  int64                          `json:"scanDurationMs"`
}

func (a *App) Scan() *ScanResult {
	scanner := detector.NewScanner()
	raw := scanner.Scan()
	dupes := analyzer.FindDuplicates(raw.Installations)
	recs := analyzer.GenerateRecommendations(raw.Installations, dupes)
	orphans := analyzer.FindOrphanedVenvs(raw.Installations)

	var reclaimable int64
	for _, r := range recs {
		reclaimable += r.SpaceSaved
	}
	for _, o := range orphans {
		reclaimable += o.SizeBytes
	}

	a.mu.Lock()
	a.lastScan = raw
	a.mu.Unlock()

	return &ScanResult{
		Installations: raw.Installations, Duplicates: dupes, Recommendations: recs,
		OrphanedVenvs: orphans, TotalSize: raw.TotalSize, ReclaimableSize: reclaimable,
		ScanDurationMs: raw.ScanDurationMs,
	}
}

// findInstallation looks up executable in lastScan under the read lock.
// Returns the matching installation and true, or zero value and false.
func (a *App) findInstallation(executable string) (models.PythonInstallation, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.lastScan == nil {
		return models.PythonInstallation{}, false
	}
	for _, inst := range a.lastScan.Installations {
		if inst.Executable == executable {
			return inst, true
		}
	}
	return models.PythonInstallation{}, false
}

// --- PATH ---

func (a *App) GetPATHAnalysis() *analyzer.PathAnalysis { return analyzer.AnalyzePATH() }

func (a *App) RepairPATH() string {
	_, err := analyzer.BackupPATH()
	if err != nil {
		return "Failed to backup PATH: " + err.Error()
	}
	n, err := analyzer.RemoveOrphanedPaths()
	if err != nil {
		return "Failed: " + err.Error()
	}
	if n == 0 {
		return "No orphaned entries found"
	}
	return "Removed " + strconv.Itoa(n) + " orphaned PATH entries (backup saved)"
}

func (a *App) SetDefaultPython(executable string) string {
	inst, ok := a.findInstallation(executable)
	if !ok {
		if a.lastScan == nil {
			return "Scan first"
		}
		return "Installation not found"
	}
	if err := analyzer.SetDefaultPython(inst.Path); err != nil {
		return "Error: " + err.Error()
	}
	return "Set Python " + inst.Version + " as default (restart terminal to take effect)"
}

// --- Health ---

func (a *App) GetHealthCheck() []*analyzer.HealthStatus {
	a.mu.RLock()
	scan := a.lastScan
	a.mu.RUnlock()
	if scan == nil {
		return nil
	}
	return analyzer.CheckAllHealth(scan.Installations)
}

// --- Packages ---

func (a *App) GetPackages(executable string) []analyzer.Package {
	if _, ok := a.findInstallation(executable); !ok {
		return nil
	}
	pkgs, err := analyzer.ListPackages(executable)
	if err != nil {
		return nil
	}
	return pkgs
}

func (a *App) GetOutdatedPackages(executable string) []analyzer.OutdatedPackage {
	if _, ok := a.findInstallation(executable); !ok {
		return nil
	}
	pkgs, err := analyzer.ListOutdated(executable)
	if err != nil {
		return nil
	}
	return pkgs
}

func (a *App) InstallPackage(executable, packageName string) string {
	if _, ok := a.findInstallation(executable); !ok {
		if a.lastScan == nil {
			return "Scan first"
		}
		return "Installation not found"
	}
	msg, err := analyzer.InstallPackage(executable, packageName)
	if err != nil {
		return "Error: " + err.Error()
	}
	return msg
}

func (a *App) UninstallPackage(executable, packageName string) string {
	if _, ok := a.findInstallation(executable); !ok {
		if a.lastScan == nil {
			return "Scan first"
		}
		return "Installation not found"
	}
	msg, err := analyzer.UninstallPackage(executable, packageName)
	if err != nil {
		return "Error: " + err.Error()
	}
	return msg
}

func (a *App) ExportRequirements(executable string) string {
	content, err := analyzer.ExportRequirements(executable)
	if err != nil {
		return "Error: " + err.Error()
	}
	return content
}

// --- Terminal ---

func (a *App) OpenTerminal(executable string) string {
	inst, ok := a.findInstallation(executable)
	if !ok {
		if a.lastScan == nil {
			return "Scan first"
		}
		return "Not found"
	}
	if err := analyzer.OpenTerminal(inst); err != nil {
		return "Error: " + err.Error()
	}
	return "OK"
}

// --- Uninstall ---

func (a *App) UninstallPython(executable string) *analyzer.UninstallResult {
	inst, ok := a.findInstallation(executable)
	if !ok {
		if a.lastScan == nil {
			return &analyzer.UninstallResult{Success: false, Message: "Scan first"}
		}
		return &analyzer.UninstallResult{Success: false, Message: "Not found"}
	}
	return analyzer.Uninstall(inst)
}

// --- Cache ---

type CacheInfo struct {
	PipSize int64  `json:"pipSize"`
	PipPath string `json:"pipPath"`
	UVSize  int64  `json:"uvSize"`
	UVPath  string `json:"uvPath"`
}

func (a *App) GetCacheInfo() *CacheInfo {
	pipSize, pipPath := analyzer.GetPipCacheSize()
	uvSize, uvPath := analyzer.GetUVCacheSize()
	return &CacheInfo{PipSize: pipSize, PipPath: pipPath, UVSize: uvSize, UVPath: uvPath}
}

func (a *App) CleanPipCache() *analyzer.UninstallResult { return analyzer.CleanPipCache() }
func (a *App) CleanUVCache() *analyzer.UninstallResult  { return analyzer.CleanUVCache() }

// --- Create Venv ---

func (a *App) CreateVenv(pythonExe, targetDir, name string) *analyzer.CreateVenvResult {
	return analyzer.CreateVenv(pythonExe, targetDir, name)
}

// --- Marketplace ---

func (a *App) LoadCatalog() []analyzer.MarketplacePackage {
	pkgs, _ := analyzer.LoadCatalog()
	return pkgs
}

func (a *App) GetCatalogCategories() []string {
	pkgs, _ := analyzer.LoadCatalog()
	return analyzer.GetCategories(pkgs)
}

func (a *App) SearchMarketplace(query string) []analyzer.MarketplacePackage {
	pkgs, _ := analyzer.SearchPyPI(query)
	return pkgs
}

func (a *App) GetPyPIDetail(name string) *analyzer.PyPIPackageDetail {
	d, _ := analyzer.FetchPyPIDetail(name)
	return d
}

// --- Python Versions ---

func (a *App) ListPythonVersions() []analyzer.PythonVersion {
	v, _ := analyzer.ListPythonVersions()
	return v
}

func (a *App) InstallPythonVersion(version string) string {
	msg, err := analyzer.InstallPythonVersion(version)
	if err != nil {
		return "Error: " + err.Error()
	}
	return msg
}

func (a *App) UninstallPythonVersion(version string) string {
	msg, err := analyzer.UninstallPythonVersion(version)
	if err != nil {
		return "Error: " + err.Error()
	}
	return msg
}

// --- Add to PATH ---

func (a *App) AddToPATH(executable string) string {
	inst, ok := a.findInstallation(executable)
	if !ok {
		return "Installation not found"
	}
	if err := analyzer.AddToPATH(inst.Path); err != nil {
		return "Error: " + err.Error()
	}
	return "Added " + inst.Path + " to PATH (restart terminal to take effect)"
}

// --- Fix My Python ---

func (a *App) GetFixReport() *analyzer.FixReport {
	a.mu.RLock()
	scan := a.lastScan
	a.mu.RUnlock()
	if scan == nil {
		return &analyzer.FixReport{ScorePercent: 100}
	}
	return analyzer.GenerateFixReport(scan.Installations)
}

func (a *App) ExecuteFix(fixAction string) string {
	switch fixAction {
	case "repair_path":
		return a.RepairPATH()
	case "clean_cache":
		r1 := analyzer.CleanPipCache()
		r2 := analyzer.CleanUVCache()
		return r1.Message + "; " + r2.Message
	default:
		return "Use the specific action buttons for this fix"
	}
}

// --- Update Check ---

func (a *App) CheckForUpdate() *analyzer.UpdateInfo {
	return analyzer.CheckForUpdate()
}

// --- Export / Import ---

type ExportData struct {
	Installations []ExportInstallation `json:"installations"`
	ExportedAt    string               `json:"exportedAt"`
	Platform      string               `json:"platform"`
}

type ExportInstallation struct {
	Version    string   `json:"version"`
	Source     string   `json:"source"`
	Path       string   `json:"path"`
	IsDefault  bool     `json:"isDefault"`
	Packages   []string `json:"packages"`
}

func (a *App) ExportEnvironment() string {
	a.mu.RLock()
	scan := a.lastScan
	a.mu.RUnlock()
	if scan == nil {
		return `{"error":"Scan first"}`
	}

	var installs []ExportInstallation
	for _, inst := range scan.Installations {
		pkgs, _ := analyzer.ListPackages(inst.Executable)
		var pkgNames []string
		for _, p := range pkgs {
			pkgNames = append(pkgNames, p.Name+"=="+p.Version)
		}
		installs = append(installs, ExportInstallation{
			Version:   inst.Version,
			Source:    string(inst.Source),
			Path:     inst.Path,
			IsDefault: inst.IsDefault,
			Packages: pkgNames,
		})
	}

	data := ExportData{
		Installations: installs,
		ExportedAt:    time.Now().Format(time.RFC3339),
		Platform:      runtime.GOOS,
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(out)
}

func (a *App) ImportEnvironment(jsonData string) string {
	var data ExportData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "Invalid JSON: " + err.Error()
	}

	var summary []string
	for _, inst := range data.Installations {
		summary = append(summary, inst.Source+" Python "+inst.Version+" ("+strconv.Itoa(len(inst.Packages))+" packages)")
	}

	return "Found " + strconv.Itoa(len(data.Installations)) + " installations:\n" + strings.Join(summary, "\n")
}
