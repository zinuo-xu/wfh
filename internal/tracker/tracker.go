// Package tracker monitors development activity sources and records them.
// All data stays local - zero telemetry, no cloud dependencies.
package tracker

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/store"
)

// ActivityEngine coordinates activity tracking across multiple sources.
type ActivityEngine struct {
	cfg     *config.Config
	store   *store.Store
	git     *GitTracker
	editor  *EditorTracker
	browser *BrowserTracker

	active     bool
	activeMu   sync.RWMutex
	lastActive time.Time
	stopCh     chan struct{}

	logger *log.Logger
}

// NewActivityEngine creates a new activity tracking engine.
func NewActivityEngine(cfg *config.Config, s *store.Store, logger *log.Logger) *ActivityEngine {
	return &ActivityEngine{
		cfg:        cfg,
		store:      s,
		git:        NewGitTracker(cfg, s, logger),
		editor:     NewEditorTracker(cfg, s, logger),
		browser:    NewBrowserTracker(cfg, s, logger),
		lastActive: time.Now(),
		stopCh:     make(chan struct{}),
		logger:     logger,
	}
}

// Start begins all tracking subsystems.
func (e *ActivityEngine) Start() error {
	e.activeMu.Lock()
	if e.active {
		e.activeMu.Unlock()
		return fmt.Errorf("tracker already running")
	}
	e.active = true
	e.activeMu.Unlock()

	e.logger.Println("Activity engine started")

	// Start periodic polling for all sources
	go e.pollLoop()

	return nil
}

// Stop gracefully stops all tracking.
func (e *ActivityEngine) Stop() error {
	e.activeMu.Lock()
	defer e.activeMu.Unlock()

	if !e.active {
		return nil
	}
	e.active = false
	close(e.stopCh)
	e.logger.Println("Activity engine stopped")
	return nil
}

// IsActive returns whether the tracking engine is running.
func (e *ActivityEngine) IsActive() bool {
	e.activeMu.RLock()
	defer e.activeMu.RUnlock()
	return e.active
}

// LastActiveTime returns when activity was last detected.
func (e *ActivityEngine) LastActiveTime() time.Time {
	e.activeMu.RLock()
	defer e.activeMu.RUnlock()
	return e.lastActive
}

// RecordActivity records a discrete activity event.
func (e *ActivityEngine) RecordActivity(category, source, detail string, duration int64) error {
	record := &store.ActivityRecord{
		StartTime: time.Now().Add(-time.Duration(duration) * time.Second),
		EndTime:   time.Now(),
		Duration:  duration,
		Category:  category,
		Source:    source,
		Detail:    detail,
		RepoPath:  e.cfg.RepoPath,
	}

	_, err := e.store.InsertActivity(record)
	if err != nil {
		return fmt.Errorf("record activity: %w", err)
	}

	e.activeMu.Lock()
	e.lastActive = time.Now()
	e.activeMu.Unlock()

	return nil
}

// pollLoop runs periodic checks on all tracking sources.
func (e *ActivityEngine) pollLoop() {
	ticker := time.NewTicker(time.Duration(e.cfg.PollIntervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.pollAll()
		}
	}
}

// pollAll checks all activity sources for new activity.
func (e *ActivityEngine) pollAll() {
	// Check git activity
	if gitActive, branch, detail := e.git.CheckActivity(); gitActive {
		e.RecordActivity("coding", "git", fmt.Sprintf("branch:%s %s", branch, detail), int64(e.cfg.PollIntervalSec))
	}

	// Check editor activity
	if editorActive, name, file := e.editor.CheckActivity(); editorActive {
		e.RecordActivity("coding", "editor", fmt.Sprintf("%s editing %s", name, file), int64(e.cfg.PollIntervalSec))
	}

	// Check browser/PR activity
	if browserActive, url, title := e.browser.CheckActivity(); browserActive {
		e.RecordActivity("reviewing", "browser", fmt.Sprintf("%s - %s", title, url), int64(e.cfg.PollIntervalSec))
	}
}
