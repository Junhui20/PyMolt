package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Config holds user preferences that persist between sessions.
type Config struct {
	VenvScanPaths []string `json:"venvScanPaths"`
	FullHomeScan  bool     `json:"fullHomeScan"`
}

var (
	mu       sync.RWMutex
	cached   *Config
	filePath string
)

func init() {
	filePath = filepath.Join(configDir(), "config.json")
}

// configDir returns the OS-appropriate config directory for PyMolt.
func configDir() string {
	switch runtime.GOOS {
	case "windows":
		if d := os.Getenv("APPDATA"); d != "" {
			return filepath.Join(d, "pymolt")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), ".config", "pymolt")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "pymolt")
	default:
		if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
			return filepath.Join(d, "pymolt")
		}
		return filepath.Join(os.Getenv("HOME"), ".config", "pymolt")
	}
}

// Load reads the config from disk, returning defaults if the file doesn't exist.
func Load() *Config {
	mu.RLock()
	if cached != nil {
		c := *cached
		mu.RUnlock()
		return &c
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// Double-check after acquiring write lock.
	if cached != nil {
		c := *cached
		return &c
	}

	cfg := &Config{}
	data, err := os.ReadFile(filePath)
	if err == nil {
		_ = json.Unmarshal(data, cfg)
	}
	cached = cfg
	c := *cfg
	return &c
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	cached = cfg
	return os.WriteFile(filePath, data, 0o644)
}

// AddVenvPath adds a scan directory and persists the change.
func AddVenvPath(dir string) error {
	cfg := Load()
	dir = filepath.Clean(dir)
	for _, p := range cfg.VenvScanPaths {
		if filepath.Clean(p) == dir {
			return nil // already present
		}
	}
	cfg.VenvScanPaths = append(cfg.VenvScanPaths, dir)
	return Save(cfg)
}

// RemoveVenvPath removes a scan directory and persists the change.
func RemoveVenvPath(dir string) error {
	cfg := Load()
	dir = filepath.Clean(dir)
	var updated []string
	for _, p := range cfg.VenvScanPaths {
		if filepath.Clean(p) != dir {
			updated = append(updated, p)
		}
	}
	cfg.VenvScanPaths = updated
	return Save(cfg)
}

// SetFullHomeScan toggles full home directory scanning.
func SetFullHomeScan(enabled bool) error {
	cfg := Load()
	cfg.FullHomeScan = enabled
	return Save(cfg)
}
