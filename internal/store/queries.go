package store

import (
	"database/sql"
	"fmt"
	"time"
)

// InsertActivity logs a new activity period.
func (s *Store) InsertActivity(a *ActivityRecord) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO activity_log (start_time, end_time, duration_sec, category, source, detail, repo_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.StartTime.Format(time.RFC3339),
		nullTime(a.EndTime),
		a.Duration,
		a.Category,
		a.Source,
		a.Detail,
		a.RepoPath,
	)
	if err != nil {
		return 0, fmt.Errorf("insert activity: %w", err)
	}
	return result.LastInsertId()
}

// InsertGitEvent records a git activity event.
func (s *Store) InsertGitEvent(e *GitEvent) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO git_events (timestamp, action, branch, message, repo_path)
		VALUES (?, ?, ?, ?, ?)`,
		e.Timestamp.Format(time.RFC3339),
		e.Action,
		e.Branch,
		e.Message,
		e.RepoPath,
	)
	if err != nil {
		return 0, fmt.Errorf("insert git event: %w", err)
	}
	return result.LastInsertId()
}

// GetActivityByRange returns activity records within a time range.
func (s *Store) GetActivityByRange(start, end time.Time) ([]ActivityRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, start_time, end_time, duration_sec, category, source, detail, repo_path
		FROM activity_log
		WHERE start_time >= ? AND start_time <= ?
		ORDER BY start_time DESC`,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("query activity: %w", err)
	}
	defer rows.Close()

	return scanActivityRows(rows)
}

// GetGitEventsByRange returns git events within a time range.
func (s *Store) GetGitEventsByRange(start, end time.Time) ([]GitEvent, error) {
	rows, err := s.db.Query(`
		SELECT id, timestamp, action, branch, message, repo_path
		FROM git_events
		WHERE timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp DESC`,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("query git events: %w", err)
	}
	defer rows.Close()

	return scanGitEventRows(rows)
}

// GetTodaysActivity returns all activity records for today.
func (s *Store) GetTodaysActivity() ([]ActivityRecord, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return s.GetActivityByRange(start, now)
}

// GetWeeksActivity returns all activity records for the current week.
func (s *Store) GetWeeksActivity() ([]ActivityRecord, error) {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day()-int(weekday-time.Monday), 0, 0, 0, 0, now.Location())
	return s.GetActivityByRange(start, now)
}

// GetDailySummary returns aggregated summary for a specific date.
func (s *Store) GetDailySummary(date string) (*Summary, error) {
	activities, err := s.GetActivityByRange(
		parseDate(date),
		parseDate(date).Add(24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	sum := &Summary{
		CategoryBreak: make(map[string]time.Duration),
	}

	for _, a := range activities {
		d := time.Duration(a.Duration) * time.Second
		sum.TotalDuration += d
		sum.CategoryBreak[a.Category] += d
	}

	// Count git events for the day
	gitEvents, err := s.GetGitEventsByRange(
		parseDate(date),
		parseDate(date).Add(24*time.Hour),
	)
	if err == nil {
		for _, e := range gitEvents {
			if e.Action == "commit" {
				sum.CommitCount++
			}
		}
	}

	return sum, nil
}

// GetWeeklySummary returns aggregated summary for the current week.
func (s *Store) GetWeeklySummary() (*Summary, error) {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-int(weekday-time.Monday), 0, 0, 0, 0, now.Location())

	activities, err := s.GetActivityByRange(weekStart, now)
	if err != nil {
		return nil, err
	}

	sum := &Summary{
		CategoryBreak: make(map[string]time.Duration),
	}

	seenDays := make(map[string]bool)
	for _, a := range activities {
		d := time.Duration(a.Duration) * time.Second
		sum.TotalDuration += d
		sum.CategoryBreak[a.Category] += d
		day := a.StartTime.Format("2006-01-02")
		if !seenDays[day] {
			seenDays[day] = true
			sum.DayCount++
		}
	}

	gitEvents, err := s.GetGitEventsByRange(weekStart, now)
	if err == nil {
		for _, e := range gitEvents {
			sum.EventCount++
			if e.Action == "commit" {
				sum.CommitCount++
			}
		}
	}

	return sum, nil
}

// PurgeOlderThan removes activity data older than the given duration.
func (s *Store) PurgeOlderThan(d time.Duration) (int64, error) {
	cutoff := time.Now().Add(-d).Format(time.RFC3339)

	result, err := s.db.Exec(`DELETE FROM activity_log WHERE start_time < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge activities: %w", err)
	}
	actRows, _ := result.RowsAffected()

	result, err = s.db.Exec(`DELETE FROM git_events WHERE timestamp < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge git events: %w", err)
	}
	gitRows, _ := result.RowsAffected()

	return actRows + gitRows, nil
}

func scanActivityRows(rows *sql.Rows) ([]ActivityRecord, error) {
	var records []ActivityRecord
	for rows.Next() {
		var r ActivityRecord
		var startStr, endStr sql.NullString
		err := rows.Scan(&r.ID, &startStr, &endStr, &r.Duration, &r.Category, &r.Source, &r.Detail, &r.RepoPath)
		if err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		r.StartTime, _ = time.Parse(time.RFC3339, startStr.String)
		if endStr.Valid {
			r.EndTime, _ = time.Parse(time.RFC3339, endStr.String)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func scanGitEventRows(rows *sql.Rows) ([]GitEvent, error) {
	var events []GitEvent
	for rows.Next() {
		var e GitEvent
		var ts string
		err := rows.Scan(&e.ID, &ts, &e.Action, &e.Branch, &e.Message, &e.RepoPath)
		if err != nil {
			return nil, fmt.Errorf("scan git event: %w", err)
		}
		e.Timestamp, _ = time.Parse(time.RFC3339, ts)
		events = append(events, e)
	}
	return events, rows.Err()
}

func nullTime(t time.Time) *string {
	if t.IsZero() {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func parseDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now()
	}
	return t
}
