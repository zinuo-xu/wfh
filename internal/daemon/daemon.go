// Package daemon manages the wfh background process lifecycle.
// The daemon runs as a lightweight background process tracking work activity.
package daemon

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/store"
	"github.com/zinuo-xu/wfh/internal/tracker"
)

// Daemon represents the background wfh process.
type Daemon struct {
	cfg      *config.Config
	store    *store.Store
	tracker  *tracker.ActivityEngine
	watcher  *FileWatcher
	logger   *log.Logger
	logFile  *os.File
	signalCh chan os.Signal
}

// New creates a new daemon instance.
func New(cfg *config.Config) (*Daemon, error) {
	// Ensure data directory exists
	if err := cfg.EnsureDataDir(); err != nil {
		return nil, fmt.Errorf("ensure data dir: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	logger := log.New(logFile, "[wfh] ", log.LstdFlags|log.Lshortfile)

	// Open database
	s, err := store.Open(cfg.DBFile)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("open store: %w", err)
	}

	// Create activity engine
	engine := tracker.NewActivityEngine(cfg, s, logger)

	// Create file watcher
	watcher := NewFileWatcher(cfg, s, logger)

	return &Daemon{
		cfg:      cfg,
		store:    s,
		tracker:  engine,
		watcher:  watcher,
		logger:   logger,
		logFile:  logFile,
		signalCh: make(chan os.Signal, 1),
	}, nil
}

// Start begins the daemon process.
func (d *Daemon) Start() error {
	d.logger.Println("Daemon starting...")

	// Write PID file
	if err := d.writePID(); err != nil {
		return fmt.Errorf("write pid: %w", err)
	}

	// Start activity tracker
	if err := d.tracker.Start(); err != nil {
		return fmt.Errorf("start tracker: %w", err)
	}

	// Start file watcher
	if err := d.watcher.Start(); err != nil {
		d.tracker.Stop()
		return fmt.Errorf("start watcher: %w", err)
	}

	d.logger.Println("Daemon started successfully")

	// Wait for shutdown signal
	signal.Notify(d.signalCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-d.signalCh
	d.logger.Printf("Received signal: %v", sig)

	return d.Stop()
}

// Stop gracefully shuts down the daemon.
func (d *Daemon) Stop() error {
	d.logger.Println("Daemon shutting down...")

	if d.watcher != nil {
		d.watcher.Stop()
	}
	if d.tracker != nil {
		d.tracker.Stop()
	}
	if d.store != nil {
		d.store.Close()
	}
	if d.logFile != nil {
		d.logFile.Close()
	}

	d.removePID()
	d.logger.Println("Daemon stopped")
	return nil
}

// IsRunning checks if the daemon process is currently running.
func IsRunning(cfg *config.Config) bool {
	pidFile := cfg.PidFile()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, signal 0 checks if process exists without sending a signal
	return process.Signal(syscall.Signal(0)) == nil
}

// writePID writes the current process ID to the PID file.
func (d *Daemon) writePID() error {
	pid := os.Getpid()
	pidDir := filepath.Dir(d.cfg.PidFile())
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(d.cfg.PidFile(), []byte(strconv.Itoa(pid)), 0644)
}

// removePID removes the PID file.
func (d *Daemon) removePID() {
	os.Remove(d.cfg.PidFile())
}
