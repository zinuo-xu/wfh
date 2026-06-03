package tracker

import (
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/store"
)

// GitTracker monitors git activity in the configured repository.
type GitTracker struct {
	cfg        *config.Config
	store      *store.Store
	logger     *log.Logger
	lastCommit string
	lastBranch string
	mu         sync.Mutex
}

// NewGitTracker creates a new Git activity tracker.
func NewGitTracker(cfg *config.Config, s *store.Store, logger *log.Logger) *GitTracker {
	return &GitTracker{
		cfg:    cfg,
		store:  s,
		logger: logger,
	}
}

// CheckActivity checks for new git activity and returns wether activity was detected.
func (g *GitTracker) CheckActivity() (bool, string, string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.cfg.RepoPath == "" {
		return false, "", ""
	}

	branch := g.currentBranch()
	commit := g.latestCommit()

	if commit == "" {
		return false, "", ""
	}

	// New commit detected
	if commit != g.lastCommit && g.lastCommit != "" {
		msg := g.commitMessage(commit)
		g.logEvent("commit", branch, msg)
		g.lastCommit = commit
		g.lastBranch = branch
		return true, branch, "new commit: " + truncate(msg, 80)
	}

	// Branch switch detected
	if branch != g.lastBranch && g.lastBranch != "" {
		g.logEvent("branch", branch, "switched from "+g.lastBranch)
		g.lastBranch = branch
		g.lastCommit = commit
		return true, branch, "switched to branch"
	}

	// First run - just record state
	if g.lastCommit == "" {
		g.lastCommit = commit
		g.lastBranch = branch
	}

	return false, branch, ""
}

// currentBranch returns the current git branch name.
func (g *GitTracker) currentBranch() string {
	cmd := exec.Command("git", "-C", g.cfg.RepoPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// latestCommit returns the latest commit hash.
func (g *GitTracker) latestCommit() string {
	cmd := exec.Command("git", "-C", g.cfg.RepoPath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// commitMessage returns the commit message for a given hash.
func (g *GitTracker) commitMessage(hash string) string {
	cmd := exec.Command("git", "-C", g.cfg.RepoPath, "log", "--format=%s", "-1", hash)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// logEvent records a git event to the store.
func (g *GitTracker) logEvent(action, branch, message string) {
	event := &store.GitEvent{
		Timestamp: time.Now(),
		Action:    action,
		Branch:    branch,
		Message:   message,
		RepoPath:  g.cfg.RepoPath,
	}
	if _, err := g.store.InsertGitEvent(event); err != nil {
		g.logger.Printf("Failed to log git event: %v", err)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
