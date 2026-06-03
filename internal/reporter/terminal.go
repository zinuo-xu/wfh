package reporter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TerminalRenderer renders reports with lipgloss styling for the terminal.
type TerminalRenderer struct {
	// Color scheme - all customizable
	primary   lipgloss.Style
	secondary lipgloss.Style
	accent    lipgloss.Style
	success   lipgloss.Style
	warning   lipgloss.Style
	dim       lipgloss.Style
	header    lipgloss.Style
	bar       lipgloss.Style
	divider   string
}

// NewTerminalRenderer creates a new terminal renderer with default styles.
func NewTerminalRenderer() *TerminalRenderer {
	primary := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)

	secondary := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EC4899"))

	accent := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#06B6D4"))

	success := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981"))

	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B"))

	dim := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 2)

	bar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED"))

	divider := strings.Repeat("─", 50)

	return &TerminalRenderer{
		primary:   primary,
		secondary: secondary,
		accent:    accent,
		success:   success,
		warning:   warning,
		dim:       dim,
		header:    header,
		bar:       bar,
		divider:   divider,
	}
}

// RenderDailyReport renders a daily activity report.
func (tr *TerminalRenderer) RenderDailyReport(report *DailyReport) string {
	var b strings.Builder

	b.WriteString(tr.header.Render(fmt.Sprintf(" Daily Report — %s ", report.Date)))
	b.WriteString("\n\n")

	// Total time
	b.WriteString(tr.primary.Render("Total Active Time: "))
	b.WriteString(tr.accent.Render(report.TotalDuration))
	b.WriteString("\n")

	// Commits
	b.WriteString(tr.primary.Render("Commits: "))
	b.WriteString(fmt.Sprintf("%d", report.CommitCount))
	b.WriteString("\n")

	// Activities
	b.WriteString(tr.primary.Render("Activity Periods: "))
	b.WriteString(fmt.Sprintf("%d", report.ActivityCount))
	b.WriteString("\n\n")

	// Category breakdown
	if len(report.CategoryBreak) > 0 {
		b.WriteString(tr.secondary.Render("Category Breakdown:"))
		b.WriteString("\n")
		for _, cat := range report.CategoryBreak {
			label := fmt.Sprintf("  %-15s", cat.Name)
			b.WriteString(label)
			b.WriteString(tr.accent.Render(fmt.Sprintf("%8s", cat.Duration)))
			b.WriteString(fmt.Sprintf(" (%5.1f%%) ", cat.Percent))
			b.WriteString(tr.bar.Render(cat.Bar))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Hourly activity heatmap
	b.WriteString(tr.secondary.Render("Hourly Activity:"))
	b.WriteString("\n")
	for i := 0; i < 24; i += 4 {
		var line strings.Builder
		for j := 0; j < 4 && i+j < 24; j++ {
			h := i + j
			count := 0
			if h < len(report.ByHour) {
				count = report.ByHour[h]
			}
			if count > 0 {
				line.WriteString(tr.success.Render(fmt.Sprintf("%02d:00 █ ", h)))
			} else {
				line.WriteString(tr.dim.Render(fmt.Sprintf("%02d:00 · ", h)))
			}
		}
		b.WriteString("  " + line.String() + "\n")
	}
	b.WriteString("\n")

	// Recent git events
	if len(report.GitEvents) > 0 {
		b.WriteString(tr.secondary.Render("Recent Git Activity:"))
		b.WriteString("\n")
		maxEvents := 5
		if len(report.GitEvents) < maxEvents {
			maxEvents = len(report.GitEvents)
		}
		for i := 0; i < maxEvents; i++ {
			e := report.GitEvents[i]
			icon := "✓"
			if e.Action == "branch" {
				icon = "⎇"
			} else if e.Action == "merge" {
				icon = "◈"
			}
			b.WriteString(fmt.Sprintf("  %s %s %s", icon, tr.dim.Render(e.Timestamp.Format("15:04")), e.Message))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(tr.divider)
	b.WriteString("\n")

	return b.String()
}

// RenderWeeklyReport renders a weekly activity digest.
func (tr *TerminalRenderer) RenderWeeklyReport(report *WeeklyReport) string {
	var b strings.Builder

	b.WriteString(tr.header.Render(fmt.Sprintf(" Weekly Report — Week of %s ", report.WeekOf)))
	b.WriteString("\n\n")

	// Summary stats
	b.WriteString(tr.primary.Render("Total Active Time: "))
	b.WriteString(tr.accent.Render(report.TotalDuration))
	b.WriteString("\n")

	b.WriteString(tr.primary.Render("Active Days: "))
	b.WriteString(fmt.Sprintf("%d", report.ActiveDays))
	b.WriteString("\n")

	b.WriteString(tr.primary.Render("Total Commits: "))
	b.WriteString(fmt.Sprintf("%d", report.CommitCount))
	b.WriteString("\n")

	b.WriteString(tr.primary.Render("Total Events: "))
	b.WriteString(fmt.Sprintf("%d", report.EventCount))
	b.WriteString("\n\n")

	// Daily breakdown
	b.WriteString(tr.secondary.Render("Daily Breakdown:"))
	b.WriteString("\n")
	for _, day := range report.DailyBreak {
		label := fmt.Sprintf("  %-4s %-12s", day.DayName, day.Date)
		b.WriteString(label)
		b.WriteString(tr.accent.Render(fmt.Sprintf("%10s", day.TotalDuration)))
		b.WriteString(" ")
		b.WriteString(tr.bar.Render(day.Bar))
		if day.CommitCount > 0 {
			b.WriteString(tr.dim.Render(fmt.Sprintf(" (%d commits)", day.CommitCount)))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Category breakdown
	if len(report.CategoryBreak) > 0 {
		b.WriteString(tr.secondary.Render("Category Breakdown:"))
		b.WriteString("\n")
		for _, cat := range report.CategoryBreak {
			label := fmt.Sprintf("  %-15s", cat.Name)
			b.WriteString(label)
			b.WriteString(tr.accent.Render(fmt.Sprintf("%8s", cat.Duration)))
			b.WriteString(fmt.Sprintf(" (%5.1f%%) ", cat.Percent))
			b.WriteString(tr.bar.Render(cat.Bar))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(tr.divider)
	b.WriteString("\n")
	b.WriteString(tr.dim.Render("  All data stored locally. No telemetry. No cloud."))
	b.WriteString("\n")

	return b.String()
}

// RenderStatus renders the daemon status.
func (tr *TerminalRenderer) RenderStatus(running bool, pid int, activeSince string, totalToday string) string {
	var b strings.Builder

	b.WriteString(tr.header.Render(" wfh Status "))
	b.WriteString("\n\n")

	if running {
		b.WriteString(tr.success.Render("● Running"))
		b.WriteString("\n")
	} else {
		b.WriteString(tr.warning.Render("○ Stopped"))
		b.WriteString("\n")
	}

	if pid > 0 {
		b.WriteString(tr.primary.Render("PID: "))
		b.WriteString(fmt.Sprintf("%d", pid))
		b.WriteString("\n")
	}

	if activeSince != "" {
		b.WriteString(tr.primary.Render("Active Since: "))
		b.WriteString(activeSince)
		b.WriteString("\n")
	}

	if totalToday != "" {
		b.WriteString(tr.primary.Render("Active Today: "))
		b.WriteString(tr.accent.Render(totalToday))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tr.divider)
	b.WriteString("\n")
	b.WriteString(tr.dim.Render("  Privacy-first: all data local, no telemetry, no cloud."))
	b.WriteString("\n")

	return b.String()
}
