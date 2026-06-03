package reporter

import (
	"fmt"
	"strings"
	"time"

	"github.com/zinuo-xu/wfh/internal/store"
)

// WeeklyReport holds the weekly activity summary data.
type WeeklyReport struct {
	WeekOf        string
	TotalDuration string
	DailyBreak    []DayEntry
	CategoryBreak []CategoryEntry
	CommitCount   int
	EventCount    int
	ActiveDays    int
}

// DayEntry shows activity for a single day in the weekly report.
type DayEntry struct {
	DayName       string
	Date          string
	TotalDuration string
	Bar           string
	CommitCount   int
}

// GenerateWeeklyReport creates a summary for the current week (Mon-Sun).
func (r *Reporter) GenerateWeeklyReport() (*WeeklyReport, error) {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()-int(weekday-time.Monday), 0, 0, 0, 0, now.Location())

	summary, err := r.store.GetWeeklySummary()
	if err != nil {
		return nil, fmt.Errorf("get weekly summary: %w", err)
	}

	report := &WeeklyReport{
		WeekOf:      monday.Format("Jan 2, 2006"),
		CommitCount: summary.CommitCount,
		EventCount:  summary.EventCount,
		ActiveDays:  summary.DayCount,
	}

	// Build daily breakdown
	maxDuration := time.Duration(0)
	dailyDurations := make(map[string]time.Duration)
	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		if day.After(now) {
			break
		}
		dateKey := day.Format("2006-01-02")
		dailyReport, err := r.GenerateDailyReport(dateKey)
		if err != nil {
			continue
		}
		dur := parseDuration(dailyReport.TotalDuration)
		if dur > maxDuration {
			maxDuration = dur
		}
		dailyDurations[day.Format("Mon")] = dur
	}

	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		if day.After(now) {
			break
		}
		dayName := day.Format("Mon")
		dur := dailyDurations[dayName]
		barLen := 0
		if maxDuration > 0 {
			barLen = int((float64(dur) / float64(maxDuration)) * 20)
		}
		bar := strings.Repeat("|", barLen)
		if bar == "" && dur > 0 {
			bar = "."
		}

		// Count commits for this day
		commitCount := 0
		dateKey := day.Format("2006-01-02")
		gitEvents, _ := r.store.GetGitEventsByRange(
			time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()),
			time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, day.Location()),
		)
		for _, e := range gitEvents {
			if e.Action == "commit" {
				commitCount++
			}
		}

		report.DailyBreak = append(report.DailyBreak, DayEntry{
			DayName:       dayName,
			Date:          day.Format("Jan 2"),
			TotalDuration: formatDuration(dur),
			Bar:           bar,
			CommitCount:   commitCount,
		})
	}

	// Category breakdown
	total := parseDuration(report.TotalDuration)
	for cat, dur := range summary.CategoryBreak {
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

	report.TotalDuration = formatDuration(total)

	return report, nil
}
