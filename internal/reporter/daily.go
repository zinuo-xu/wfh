package reporter

import (
	"fmt"
	"strings"
	"time"

	"github.com/zinuo-xu/wfh/internal/store"
)

// DailyReport holds the daily activity summary data.
type DailyReport struct {
	Date          string
	TotalDuration string
	CategoryBreak []CategoryEntry
	GitEvents     []store.GitEvent
	ActivityCount int
	CommitCount   int
	ByHour        []int // Activities per hour (0-23)
}

// CategoryEntry shows time spent in a category.
type CategoryEntry struct {
	Name     string
	Duration string
	Percent  float64
	Bar      string
}

// GenerateDailyReport creates a report for the specified date.
func (r *Reporter) GenerateDailyReport(date string) (*DailyReport, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		// Default to today
		now := time.Now()
		date = now.Format("2006-01-02")
		parsedDate = now
	}

	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	activities, err := r.store.GetActivityByRange(startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("get activities: %w", err)
	}

	gitEvents, err := r.store.GetGitEventsByRange(startOfDay, endOfDay)
	if err != nil {
		gitEvents = []store.GitEvent{}
	}

	report := &DailyReport{
		Date:          date,
		ActivityCount: len(activities),
		GitEvents:     gitEvents,
	}

	// Calculate totals and category breakdown
	categoryDurations := make(map[string]time.Duration)
	hourlyActivity := make([]int, 24)

	for _, a := range activities {
		d := time.Duration(a.Duration) * time.Second
		categoryDurations[a.Category] += d
		report.TotalDuration = formatDuration(parseDuration(report.TotalDuration) + d)
		hour := a.StartTime.Hour()
		hourlyActivity[hour]++
	}

	report.ByHour = hourlyActivity

	total := parseDuration(report.TotalDuration)
	for cat, dur := range categoryDurations {
		percent := 0.0
		if total > 0 {
			percent = (float64(dur) / float64(total)) * 100
		}
		bar := strings.Repeat("|", int(percent/10))
		if bar == "" && percent > 0 {
			bar = "."
		}
		report.CategoryBreak = append(report.CategoryBreak, CategoryEntry{
			Name:     cat,
			Duration: formatDuration(dur),
			Percent:  percent,
			Bar:      bar,
		})
	}

	for _, e := range gitEvents {
		if e.Action == "commit" {
			report.CommitCount++
		}
	}

	return report, nil
}

// parseDuration parses a formatted duration string back to a duration.
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	var hours, minutes int
	fmt.Sscanf(s, "%dh %dm", &hours, &minutes)
	if hours == 0 {
		fmt.Sscanf(s, "%dm", &minutes)
	}
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
}
