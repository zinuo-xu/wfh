package daemon

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/store"
)

// FileWatcher monitors the filesystem for changes using fsnotify.
// It detects file modifications, creations, and deletions in tracked directories.
type FileWatcher struct {
	cfg      *config.Config
	store    *store.Store
	logger   *log.Logger
	watcher  *fsnotify.Watcher
	stopCh   chan struct{}
	done     sync.WaitGroup
	events   uint64
	lastLog  time.Time
}

// NewFileWatcher creates a new filesystem watcher.
func NewFileWatcher(cfg *config.Config, s *store.Store, logger *log.Logger) *FileWatcher {
	return &FileWatcher{
		cfg:     cfg,
		store:   s,
		logger:  logger,
		stopCh:  make(chan struct{}),
		lastLog: time.Now(),
	}
}

// Start begins watching directories for changes.
func (fw *FileWatcher) Start() error {
	var err error
	fw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	watchedDirs := fw.buildWatchList()
	for _, dir := range watchedDirs {
		if err := fw.watcher.Add(dir); err != nil {
			fw.logger.Printf("Warning: could not watch %s: %v", dir, err)
		} else {
			fw.logger.Printf("Watching: %s", dir)
		}
	}

	fw.done.Add(1)
	go fw.eventLoop()

	return nil
}

// Stop stops the file watcher.
func (fw *FileWatcher) Stop() {
	close(fw.stopCh)
	if fw.watcher != nil {
		fw.watcher.Close()
	}
	fw.done.Wait()
}

// buildWatchList compiles the list of directories to watch.
func (fw *FileWatcher) buildWatchList() []string {
	dirs := []string{}

	if fw.cfg.RepoPath != "" {
		dirs = append(dirs, fw.cfg.RepoPath)
	}

	dirs = append(dirs, fw.cfg.WatchDirs...)

	return dirs
}

// eventLoop processes filesystem events.
func (fw *FileWatcher) eventLoop() {
	defer fw.done.Done()

	for {
		select {
		case <-fw.stopCh:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Printf("Watcher error: %v", err)
		}
	}
}

// handleEvent processes a single filesystem event.
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Ignore hidden files and common VCS/editor temp files
	name := filepath.Base(event.Name)
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "#") ||
		strings.HasSuffix(name, "~") || strings.HasSuffix(name, ".swp") ||
		strings.HasSuffix(name, ".swx") {
		return
	}

	// Extract file extension for categorisation
	ext := filepath.Ext(event.Name)
	category := classifyExtension(ext)

	// Log activity
	duration := int64(30) // Default 30-second activity block
	fw.store.InsertActivity(&store.ActivityRecord{
		StartTime: time.Now().Add(-30 * time.Second),
		EndTime:   time.Now(),
		Duration:  duration,
		Category:  category,
		Source:    "file",
		Detail:    event.Name,
		RepoPath:  fw.cfg.RepoPath,
	})

	fw.events++
	if time.Since(fw.lastLog) > 30*time.Second {
		fw.logger.Printf("Processed %d file events in the last 30s", fw.events)
		fw.events = 0
		fw.lastLog = time.Now()
	}
}

// classifyExtension maps file extensions to activity categories.
func classifyExtension(ext string) string {
	switch ext {
	case ".go", ".rs", ".py", ".js", ".ts", ".jsx", ".tsx":
		return "coding"
	case ".java", ".kt", ".swift", ".c", ".h", ".cpp", ".hpp", ".cs":
		return "coding"
	case ".rb", ".php", ".vue", ".css", ".scss", ".html":
		return "coding"
	case ".md", ".txt", ".rst", ".adoc":
		return "writing"
	case ".json", ".yaml", ".yml", ".toml", ".xml":
		return "configuring"
	case ".sql", ".db", ".sqlite":
		return "data"
	case ".sh", ".ps1", ".bat", ".Makefile", ".mk":
		return "scripting"
	case ".mod", ".sum":
		return "dependencies"
	default:
		return "coding"
	}
}
