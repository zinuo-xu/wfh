package tracker

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/store"
)

// EditorTracker monitors editor activity by watching recently modified files.
// It checks for common editor swap/lock files and modified source files.
type EditorTracker struct {
	cfg       *config.Config
	store     *store.Store
	logger    *log.Logger
	knownExts map[string]bool
	lastFiles map[string]time.Time
	mu        sync.Mutex
}

// NewEditorTracker creates a new editor activity tracker.
func NewEditorTracker(cfg *config.Config, s *store.Store, logger *log.Logger) *EditorTracker {
	return &EditorTracker{
		cfg:    cfg,
		store:  s,
		logger: logger,
		knownExts: map[string]bool{
			".go": true, ".py": true, ".js": true, ".ts": true,
			".jsx": true, ".tsx": true, ".rs": true, ".java": true,
			".kt": true, ".swift": true, ".c": true, ".h": true,
			".cpp": true, ".hpp": true, ".cs": true, ".rb": true,
			".php": true, ".vue": true, ".css": true, ".scss": true,
			".html": true, ".md": true, ".json": true, ".yaml": true,
			".yml": true, ".toml": true, ".sql": true, ".sh": true,
			".ps1": true, ".bat": true, ".proto": true, ".mod": true,
			".sum": true,
		},
		lastFiles: make(map[string]time.Time),
	}
}

// CheckActivity checks for recently modified files indicating editor activity.
func (e *EditorTracker) CheckActivity() (bool, string, string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cfg.RepoPath == "" {
		return false, "", ""
	}

	activeEditor := e.detectEditor()
	recentFiles := e.findRecentFiles()

	if len(recentFiles) > 0 {
		// Update last seen files
		for _, f := range recentFiles {
			info, err := os.Stat(f)
			if err == nil {
				e.lastFiles[f] = info.ModTime()
			}
		}
		return true, activeEditor, recentFiles[0]
	}

	return false, "", ""
}

// detectEditor attempts to detect the currently running editor.
func (e *EditorTracker) detectEditor() string {
	// Check common editor process names (cross-platform hints)
	editors := []struct {
		name string
		path string
	}{
		{"VS Code", ".vscode"},
		{"Cursor", ".cursor"},
		{"JetBrains", ".idea"},
		{"Neovim", ".nvim"},
		{"Vim", ".vim"},
		{"Emacs", ".emacs"},
		{"Sublime", ".sublime"},
		{"Zed", ".zed"},
	}

	home, _ := os.UserHomeDir()
	for _, ed := range editors {
		configDir := filepath.Join(home, ed.path)
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			return ed.name
		}
	}

	return "Unknown Editor"
}

// findRecentFiles finds source files modified within the poll interval.
func (e *EditorTracker) findRecentFiles() []string {
	var recent []string
	cutoff := time.Now().Add(-time.Duration(e.cfg.PollIntervalSec) * time.Second)

	filepath.Walk(e.cfg.RepoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}

		// Skip hidden directories and vendor/deps
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(info.Name())
		if !e.knownExts[ext] {
			return nil
		}

		// Skip editor swap/lock files
		if strings.HasSuffix(info.Name(), ".swp") || strings.HasSuffix(info.Name(), ".swx") ||
			strings.HasSuffix(info.Name(), "~") || strings.HasPrefix(info.Name(), ".#") {
			return nil
		}

		// Check if modified recently and not already seen
		if info.ModTime().After(cutoff) {
			if lastMod, seen := e.lastFiles[path]; !seen || info.ModTime().After(lastMod) {
				recent = append(recent, path)
			}
		}

		return nil
	})

	return recent
}
