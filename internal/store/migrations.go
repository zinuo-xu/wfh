package store

import "fmt"

// migrate runs database schema migrations.
func (s *Store) migrate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check current schema version
	var version int
	err := s.db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		version = 0
	}

	// Run migrations sequentially
	migrations := []struct {
		version int
		sql     string
	}{
		{1, schemaV1},
		{2, schemaV2},
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback()

	for _, m := range migrations {
		if m.version > version {
			if _, err := tx.Exec(m.sql); err != nil {
				return fmt.Errorf("migration v%d: %w", m.version, err)
			}
			if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
				return fmt.Errorf("set version %d: %w", m.version, err)
			}
		}
	}

	return tx.Commit()
}

const schemaV1 = `
CREATE TABLE IF NOT EXISTS activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    start_time TEXT NOT NULL,
    end_time TEXT,
    duration_sec INTEGER NOT NULL DEFAULT 0,
    category TEXT NOT NULL DEFAULT 'coding',
    source TEXT NOT NULL DEFAULT 'file',
    detail TEXT NOT NULL DEFAULT '',
    repo_path TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_activity_start ON activity_log(start_time);
CREATE INDEX IF NOT EXISTS idx_activity_category ON activity_log(category);
CREATE INDEX IF NOT EXISTS idx_activity_source ON activity_log(source);

CREATE TABLE IF NOT EXISTS git_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    action TEXT NOT NULL,
    branch TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    repo_path TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_git_timestamp ON git_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_git_action ON git_events(action);
`

const schemaV2 = `
CREATE TABLE IF NOT EXISTS config_store (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS daily_summaries (
    date TEXT PRIMARY KEY,
    total_seconds INTEGER NOT NULL DEFAULT 0,
    commit_count INTEGER NOT NULL DEFAULT 0,
    category_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`
