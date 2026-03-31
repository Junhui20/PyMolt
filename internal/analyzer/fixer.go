package analyzer

import (
	"github.com/Junhui20/PyMolt/internal/models"
)

// FixIssue represents a detected problem with a suggested fix.
type FixIssue struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`    // "critical", "warning", "info"
	Title       string `json:"title"`
	Description string `json:"description"`
	FixAction   string `json:"fixAction"`   // what the fix button does
	FixLabel    string `json:"fixLabel"`     // button text
	SpaceSaved  int64  `json:"spaceSaved"`  // bytes reclaimable, 0 if N/A
	AutoFixable bool   `json:"autoFixable"` // can be fixed with one click
}

// FixReport is the full diagnostic report.
type FixReport struct {
	Issues       []FixIssue `json:"issues"`
	TotalIssues  int        `json:"totalIssues"`
	Critical     int        `json:"critical"`
	Warnings     int        `json:"warnings"`
	ReclaimBytes int64      `json:"reclaimBytes"`
	ScorePercent int        `json:"scorePercent"` // 0-100, 100 = perfect
}

// GenerateFixReport scans everything and produces a fix plan.
func GenerateFixReport(installs []models.PythonInstallation) *FixReport {
	report := &FixReport{}

	// 1. Check for duplicates
	dupes := FindDuplicates(installs)
	for _, d := range dupes {
		var removable int64
		for _, inst := range d.Installations {
			if d.RecommendKeep == nil || inst.Executable != d.RecommendKeep.Executable {
				removable += inst.SizeBytes
			}
		}
		report.Issues = append(report.Issues, FixIssue{
			ID:          "dup-" + d.Version,
			Severity:    "warning",
			Title:       "Duplicate Python " + d.Version,
			Description: "Python " + d.Version + " is installed from " + itoa(len(d.Installations)) + " different sources. Keep one, remove the rest.",
			FixAction:   "remove_duplicates",
			FixLabel:    "Remove duplicates",
			SpaceSaved:  removable,
			AutoFixable: true,
		})
		report.ReclaimBytes += removable
	}

	// 2. Check for orphaned venvs
	orphans := FindOrphanedVenvs(installs)
	for _, o := range orphans {
		report.Issues = append(report.Issues, FixIssue{
			ID:          "orphan-" + o.Path,
			Severity:    "warning",
			Title:       "Orphaned venv: " + o.Path,
			Description: "This virtual environment's base Python no longer exists. It cannot be used.",
			FixAction:   "delete_venv",
			FixLabel:    "Delete venv",
			SpaceSaved:  o.SizeBytes,
			AutoFixable: true,
		})
		report.ReclaimBytes += o.SizeBytes
	}

	// 3. Check PATH issues
	pathAnalysis := AnalyzePATH()
	if pathAnalysis.OrphanedCount > 0 {
		report.Issues = append(report.Issues, FixIssue{
			ID:          "path-orphaned",
			Severity:    "critical",
			Title:       itoa(pathAnalysis.OrphanedCount) + " orphaned PATH entries",
			Description: "PATH contains entries pointing to directories that no longer exist. This can cause 'python not found' errors.",
			FixAction:   "repair_path",
			FixLabel:    "Repair PATH",
			AutoFixable: true,
		})
	}

	if len(pathAnalysis.Conflicts) > 0 {
		for _, c := range pathAnalysis.Conflicts {
			if c != "" {
				report.Issues = append(report.Issues, FixIssue{
					ID:       "path-conflict",
					Severity: "warning",
					Title:    "PATH conflict",
					Description: c,
					FixAction: "repair_path",
					FixLabel:  "Repair PATH",
					AutoFixable: true,
				})
			}
		}
	}

	// 4. Check health
	healths := CheckAllHealth(installs)
	for _, h := range healths {
		if h.Overall == "Broken" {
			report.Issues = append(report.Issues, FixIssue{
				ID:          "broken-" + h.Version,
				Severity:    "critical",
				Title:       "Python " + h.Version + " is broken",
				Description: "Cannot execute. Consider removing it.",
				FixAction:   "uninstall",
				FixLabel:    "Remove",
				AutoFixable: false,
			})
		} else if h.Overall == "Warning" {
			for _, issue := range h.Issues {
				report.Issues = append(report.Issues, FixIssue{
					ID:          "health-" + h.Version,
					Severity:    "info",
					Title:       "Python " + h.Version + ": " + issue,
					Description: issue,
					FixAction:   "none",
					FixLabel:    "",
					AutoFixable: false,
				})
			}
		}
	}

	// 5. Check cache size
	pipSize, _ := GetPipCacheSize()
	uvSize, _ := GetUVCacheSize()
	totalCache := pipSize + uvSize
	if totalCache > 50*1024*1024 { // >50MB
		report.Issues = append(report.Issues, FixIssue{
			ID:          "cache",
			Severity:    "info",
			Title:       "Cache using " + models.FormatSize(totalCache),
			Description: "pip and uv caches can be safely cleaned to reclaim disk space.",
			FixAction:   "clean_cache",
			FixLabel:    "Clean caches",
			SpaceSaved:  totalCache,
			AutoFixable: true,
		})
		report.ReclaimBytes += totalCache
	}

	// Calculate score
	report.TotalIssues = len(report.Issues)
	for _, issue := range report.Issues {
		if issue.Severity == "critical" {
			report.Critical++
		} else if issue.Severity == "warning" {
			report.Warnings++
		}
	}

	// Score: start at 100, deduct for issues
	score := 100
	score -= report.Critical * 20
	score -= report.Warnings * 10
	score -= (report.TotalIssues - report.Critical - report.Warnings) * 3
	if score < 0 {
		score = 0
	}
	report.ScorePercent = score

	return report
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
