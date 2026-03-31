package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Junhui20/PyMolt/internal/analyzer"
	"github.com/Junhui20/PyMolt/internal/detector"
	"github.com/Junhui20/PyMolt/internal/models"
)

// Run handles CLI commands. Returns true if a CLI command was handled.
func Run(args []string) bool {
	if len(args) < 2 {
		return false
	}

	switch args[1] {
	case "scan":
		cmdScan()
	case "fix":
		cmdFix()
	case "versions":
		cmdVersions()
	case "health":
		cmdHealth()
	case "path":
		cmdPath()
	case "cache":
		cmdCache()
	case "help", "--help", "-h":
		cmdHelp()
	default:
		return false
	}
	return true
}

func cmdHelp() {
	fmt.Println(`PyMolt - Python environment manager

Usage:
  pymolt              Open GUI
  pymolt scan         Scan all Python installations
  pymolt fix          Auto-detect and fix issues
  pymolt versions     List installed/available Python versions
  pymolt health       Run health checks
  pymolt path         Analyze PATH entries
  pymolt cache        Show cache sizes
  pymolt help         Show this help`)
}

func cmdScan() {
	fmt.Println("Scanning for Python installations...")
	s := detector.NewScanner()
	result := s.Scan()

	fmt.Printf("\nFound %d installations (%s total) in %dms\n\n",
		len(result.Installations), models.FormatSize(result.TotalSize), result.ScanDurationMs)

	for _, inst := range result.Installations {
		def := ""
		if inst.IsDefault {
			def = " [DEFAULT]"
		}
		fmt.Printf("  %-20s  %-10s  %-6s  %8s  %s%s\n",
			inst.Source, inst.Version, inst.Architecture,
			models.FormatSize(inst.SizeBytes), inst.Path, def)
	}
}

func cmdFix() {
	fmt.Println("Scanning...")
	s := detector.NewScanner()
	raw := s.Scan()

	dupes := analyzer.FindDuplicates(raw.Installations)
	orphans := analyzer.FindOrphanedVenvs(raw.Installations)
	pathAnalysis := analyzer.AnalyzePATH()

	issues := 0

	// Duplicates
	if len(dupes) > 0 {
		fmt.Printf("\n[!] %d duplicate version group(s):\n", len(dupes))
		for _, d := range dupes {
			fmt.Printf("    Python %s found in %d sources\n", d.Version, len(d.Installations))
			for _, inst := range d.Installations {
				keep := ""
				if d.RecommendKeep != nil && inst.Executable == d.RecommendKeep.Executable {
					keep = " <- KEEP"
				}
				fmt.Printf("      - %s: %s%s\n", inst.Source, inst.Path, keep)
			}
		}
		issues += len(dupes)
	}

	// Orphaned venvs
	if len(orphans) > 0 {
		fmt.Printf("\n[!] %d orphaned virtual environment(s):\n", len(orphans))
		for _, o := range orphans {
			fmt.Printf("    %s (%s) - base Python missing\n", o.Path, models.FormatSize(o.SizeBytes))
		}
		issues += len(orphans)
	}

	// PATH issues
	if pathAnalysis.OrphanedCount > 0 {
		fmt.Printf("\n[!] %d orphaned PATH entry/entries:\n", pathAnalysis.OrphanedCount)
		for _, e := range pathAnalysis.Entries {
			if e.Orphaned {
				fmt.Printf("    %s (does not exist)\n", e.Path)
			}
		}
		issues += pathAnalysis.OrphanedCount
	}

	// Cache
	pipSize, _ := analyzer.GetPipCacheSize()
	uvSize, _ := analyzer.GetUVCacheSize()
	if pipSize+uvSize > 100*1024*1024 { // >100MB
		fmt.Printf("\n[!] Cache using %s (pip: %s, uv: %s)\n",
			models.FormatSize(pipSize+uvSize), models.FormatSize(pipSize), models.FormatSize(uvSize))
		issues++
	}

	if issues == 0 {
		fmt.Println("\nAll clean! No issues found.")
	} else {
		fmt.Printf("\n%d issue(s) found. Run the GUI for interactive fixes: pymolt\n", issues)
	}
}

func cmdVersions() {
	vers, err := analyzer.ListPythonVersions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	fmt.Println("Installed:")
	for _, v := range vers {
		if v.Installed {
			fmt.Printf("  Python %-10s  %s\n", v.Version, v.Path)
		}
	}
	fmt.Println("\nAvailable:")
	for _, v := range vers {
		if !v.Installed {
			fmt.Printf("  Python %s\n", v.Version)
		}
	}
}

func cmdHealth() {
	fmt.Println("Running health checks...")
	s := detector.NewScanner()
	raw := s.Scan()
	results := analyzer.CheckAllHealth(raw.Installations)

	for _, h := range results {
		status := h.Overall
		checks := []string{}
		if h.PythonOK {
			checks = append(checks, "python:OK")
		} else {
			checks = append(checks, "python:FAIL")
		}
		if h.PipOK {
			checks = append(checks, "pip:OK")
		} else {
			checks = append(checks, "pip:FAIL")
		}
		if h.SslOK {
			checks = append(checks, "ssl:OK")
		} else {
			checks = append(checks, "ssl:FAIL")
		}
		if h.SiteOK {
			checks = append(checks, "site:OK")
		} else {
			checks = append(checks, "site:FAIL")
		}
		fmt.Printf("  %-10s  %-8s  %s\n", h.Version, status, strings.Join(checks, "  "))
	}
}

func cmdPath() {
	p := analyzer.AnalyzePATH()
	if len(p.Entries) == 0 {
		fmt.Println("No Python-related PATH entries found.")
		return
	}
	for i, e := range p.Entries {
		status := ""
		if e.Orphaned {
			status = " [ORPHANED]"
		} else if e.HasPython && i == 0 {
			status = " [DEFAULT]"
		}
		src := fmt.Sprintf("[%s]", e.Source)
		fmt.Printf("  %d. %-8s %s%s\n", i+1, src, e.Path, status)
	}
	if len(p.Conflicts) > 0 {
		fmt.Println("\nIssues:")
		for _, c := range p.Conflicts {
			fmt.Printf("  - %s\n", c)
		}
	}
}

func cmdCache() {
	pipSize, pipPath := analyzer.GetPipCacheSize()
	uvSize, uvPath := analyzer.GetUVCacheSize()
	fmt.Printf("  pip cache:  %8s  %s\n", models.FormatSize(pipSize), pipPath)
	fmt.Printf("  uv cache:   %8s  %s\n", models.FormatSize(uvSize), uvPath)
	fmt.Printf("  Total:      %8s\n", models.FormatSize(pipSize+uvSize))
}
