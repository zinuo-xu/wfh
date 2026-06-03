// Package store provides the SQLite data layer for wfh.
// All data is stored locally. No data ever leaves the machine.
package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database connection and provides access methods.
type Store struct {
	db   *sql.DB
	mu   sync.Mutex
}

// ActivityRecord represents a tracked activity period.
type ActivityRecord struct {
	ID        int64     `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Duration  int64     `json:"duration_sec"`
	Category  string    `json:"category"`   // "coding", "debugging", "reviewing", "writing"
	Source    string    `json:"source"`     // "git", "editor", "browser", "file"
	Detail    string    `json:"detail"`     // branch name, file path, PR URL
	RepoPath  string    `json:"repo_path"`
}

// GitEvent represents a git activity event.
type GitEvent struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`     // "commit", "branch", "merge", "push"
	Branch    string    `json:"branch"`
	Message   string    `json:"message"`
	RepoPath  string    `json:"repo_path"`
}

// Summary holds aggregated activity data.
type Summary struct {
	TotalDuration time.Duration
	CategoryBreak map[string]time.Duration
	EventCount    int
	CommitCount   int
	DayCount      int
}

// Open creates or opens the SQLite database and runs migrations.
func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?cache=shared&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for use in queries.
func (s *Store) DB() *sql.DB {
	return s.db
}
