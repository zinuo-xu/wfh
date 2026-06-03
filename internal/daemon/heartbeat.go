package daemon

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Heartbeat tracks active/inactive periods based on file system activity.
// It records "active" time spans when recent file changes are detected
// and pauses when no changes occur within the heartbeat timeout.
type Heartbeat struct {
	cfg              *config.Config
	logger           *log.Logger
	active           bool
	activeStart      time.Time
	totalActiveToday time.Duration
	mu               sync.Mutex
	stopCh           chan struct{}
}

// NewHeartbeat creates a new heartbeat tracker.
func NewHeartbeat(cfg *config.Config, logger *log.Logger) *Heartbeat {
	return &Heartbeat{
		cfg:    cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start begins the heartbeat monitoring loop.
func (h *Heartbeat) Start() {
	go h.loop()
}

// Stop stops the heartbeat monitoring.
func (h *Heartbeat) Stop() {
	close(h.stopCh)
}

// IsActive returns whether the user is currently considered active.
func (h *Heartbeat) IsActive() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.active
}

// TotalActiveToday returns the total active duration for today.
func (h *Heartbeat) TotalActiveToday() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	dur := h.totalActiveToday
	if h.active {
		dur += time.Since(h.activeStart)
	}
	return dur
}

// Ping records a heartbeat event, marking the user as active.
func (h *Heartbeat) Ping() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.active {
		h.active = true
		h.activeStart = time.Now()
		h.logger.Println("Activity started")
	}
}

// loop periodically checks for file modifications to determine active state.
func (h *Heartbeat) loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.checkFileActivity()
		}
	}
}

// checkFileActivity looks for recent file modifications in the repo.
func (h *Heartbeat) checkFileActivity() {
	if h.cfg.RepoPath == "" {
		return
	}

	recentMod := false
	cutoff := time.Now().Add(-time.Duration(h.cfg.HeartbeatTimeoutSec) * time.Second)

	// Walk recent files (limit depth for performance)
	filepath.Walk(h.cfg.RepoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || recentMod {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if info.ModTime().After(cutoff) {
			recentMod = true
		}
		return nil
	})

	h.mu.Lock()
	defer h.mu.Unlock()

	if recentMod {
		if !h.active {
			h.active = true
			h.activeStart = time.Now()
			h.logger.Println("Activity resumed (file changes detected)")
		}
	} else {
		if h.active {
			// Check timeout
			if time.Since(h.activeStart) > time.Duration(h.cfg.HeartbeatTimeoutSec)*time.Second {
				h.totalActiveToday += time.Since(h.activeStart)
				h.active = false
				h.logger.Println("Activity stopped (timeout)")
			}
		}
	}
}
