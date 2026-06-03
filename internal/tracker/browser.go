package tracker

import (
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/store"
)

// BrowserTracker monitors browser activity, particularly code review and PR pages.
// Privacy-first: only captures URLs that match known code review patterns.
// No browsing history is collected.
type BrowserTracker struct {
	cfg    *config.Config
	store  *store.Store
	logger *log.Logger
	mu     sync.Mutex
	// Active PRs being reviewed (cached to avoid duplicate events)
	activePRs map[string]bool
}

// NewBrowserTracker creates a new browser activity tracker.
func NewBrowserTracker(cfg *config.Config, s *store.Store, logger *log.Logger) *BrowserTracker {
	return &BrowserTracker{
		cfg:       cfg,
		store:     s,
		logger:    logger,
		activePRs: make(map[string]bool),
	}
}

// CheckActivity checks for code-review-related browser activity.
func (b *BrowserTracker) CheckActivity() (bool, string, string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// On macOS, check Safari/Chrome tabs for PR pages
	// On Linux, check available window info
	// On Windows, check browser processes
	switch runtime.GOOS {
	case "darwin":
		return b.checkMacOS()
	case "linux":
		return b.checkLinux()
	case "windows":
		return b.checkWindows()
	}
	return false, "", ""
}

// checkMacOS uses AppleScript to get browser URLs (macOS only).
func (b *BrowserTracker) checkMacOS() (bool, string, string) {
	browsers := []string{"Google Chrome", "Safari", "Brave Browser", "Arc", "Firefox"}
	for _, browser := range browsers {
		script := `tell application "` + browser + `"
			if it is running then
				get URL of active tab of front window
			end if
		end tell`
		cmd := exec.Command("osascript", "-e", script)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		url := strings.TrimSpace(string(out))
		if prInfo := b.matchPRURL(url); prInfo != "" {
			return true, browser, prInfo
		}
	}
	return false, "", ""
}

// checkLinux attempts to read browser activity via available tools.
func (b *BrowserTracker) checkLinux() (bool, string, string) {
	// Try xdotool to get active window title
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	out, err := cmd.Output()
	if err == nil {
		title := strings.TrimSpace(string(out))
		if strings.Contains(title, "Pull Request") || strings.Contains(title, "PR") ||
			strings.Contains(title, "GitHub") || strings.Contains(title, "GitLab") {
			return true, "Browser", title
		}
	}
	return false, "", ""
}

// checkWindows checks for browser windows with code review tabs (Windows).
func (b *BrowserTracker) checkWindows() (bool, string, string) {
	// Check for browser processes
	cmd := exec.Command("powershell", "-Command",
		`Get-Process | Where-Object { $_.MainWindowTitle -ne "" } | Select-Object -ExpandProperty MainWindowTitle`)
	out, err := cmd.Output()
	if err != nil {
		return false, "", ""
	}

	titles := strings.Split(string(out), "\n")
	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		if b.matchWindowTitle(title) {
			return true, "Browser", title
		}
	}
	return false, "", ""
}

// matchPRURL checks if a URL matches known code review patterns.
func (b *BrowserTracker) matchPRURL(url string) string {
	patterns := []struct {
		host   string
		prefix string
	}{
		{"github.com", "/pull/"},
		{"gitlab.com", "/merge_requests/"},
		{"bitbucket.org", "/pull-requests/"},
		{"dev.azure.com", "/pullrequest/"},
	}

	for _, p := range patterns {
		if strings.Contains(url, p.host) && strings.Contains(url, p.prefix) {
			if !b.activePRs[url] {
				b.activePRs[url] = true
			}
			return url
		}
	}
	return ""
}

// matchWindowTitle checks window titles for code review indicators.
func (b *BrowserTracker) matchWindowTitle(title string) bool {
	keywords := []string{
		"Pull Request", "Merge Request", "PR",
		"GitHub", "GitLab", "Code Review",
	}
	titleLower := strings.ToLower(title)
	for _, kw := range keywords {
		if strings.Contains(titleLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
