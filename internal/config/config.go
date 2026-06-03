// Package config provides configuration management for the wfh application.
// All configuration is stored locally; no telemetry or cloud sync.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appName    = "wfh"
	configDir  = ".config"
	defaultDB  = "wfh.db"
	defaultLog = "wfh.log"
)

// Config holds all application configuration.
type Config struct {
	// DataDir is the directory for database and log files.
	DataDir string `json:"data_dir"`

	// DBFile is the path to the SQLite database.
	DBFile string `json:"db_file"`

	// LogFile is the path to the daemon log file.
	LogFile string `json:"log_file"`

	// WatchDirs are additional directories to watch beyond the home repo.
	WatchDirs []string `json:"watch_dirs,omitempty"`

	// PollIntervalSec is how often to check activity (seconds).
	PollIntervalSec int `json:"poll_interval_sec"`

	// HeartbeatTimeoutSec is inactivity timeout before stopping tracking (seconds).
	HeartbeatTimeoutSec int `json:"heartbeat_timeout_sec"`

	// RepoPath is the git repository path to track.
	RepoPath string `json:"repo_path"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, configDir, appName)

	return &Config{
		DataDir:             dataDir,
		DBFile:              filepath.Join(dataDir, defaultDB),
		LogFile:             filepath.Join(dataDir, defaultLog),
		PollIntervalSec:     30,
		HeartbeatTimeoutSec: 300, // 5 minutes
		WatchDirs:           []string{},
	}
}

// Load reads configuration from disk, falling back to defaults.
func Load() (*Config, error) {
	cfg := DefaultConfig()
	path, err := configPath()
	if err != nil {
		return cfg, nil // Use defaults if no config path
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Save writes configuration to disk.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// EnsureDataDir creates the data directory if it doesn't exist.
func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0755)
}

// PidFile returns the path to the daemon PID file.
func (c *Config) PidFile() string {
	return filepath.Join(c.DataDir, "wfh.pid")
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, appName, "config.json"), nil
}

# feat: add progress bar for long-running operations (incremental change 1)

# test: fix flaky test with proper cleanup (incremental change 2)

# refactor: use generator expressions for memory efficiency (incremental change 3)

# chore: add security policy (incremental change 4)

# perf: reduce allocations in inner loop (incremental change 5)
