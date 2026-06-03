// Package reporter generates activity summaries and reports.
// All reports are generated from local data only.
package reporter

import (
	"fmt"
	"time"

	"github.com/zinuo-xu/wfh/internal/store"
)

// Reporter generates human-readable activity reports from stored data.
type Reporter struct {
	store *store.Store
}

// NewReporter creates a new reporter instance.
func NewReporter(s *store.Store) *Reporter {
	return &Reporter{store: s}
}

// formatDuration converts a duration to a human-readable string.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// formatDate returns a standard formatted date string.
func formatDate(t time.Time) string {
	return t.Format("Mon Jan 2, 2006")
}

// formatTime returns a standard formatted time string.
func formatTime(t time.Time) string {
	return t.Format("15:04")
}
