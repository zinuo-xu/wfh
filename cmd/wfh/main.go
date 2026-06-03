package main

import (
	"fmt"
	"os"

	"github.com/zinuo-xu/wfh/internal/config"
	"github.com/zinuo-xu/wfh/internal/daemon"
	"github.com/zinuo-xu/wfh/internal/reporter"
	"github.com/zinuo-xu/wfh/internal/store"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "wfh",
	Short: "wfh — Work From Home activity tracker",
	Long: `wfh tracks your development activity locally.

Privacy-first: all data is stored on your machine. No telemetry.
No cloud sync. No data ever leaves your computer.

wfh monitors git activity, file changes, and editor usage to
provide daily and weekly productivity reports.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("load config: %w", err)
		}
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the wfh daemon in the background",
	RunE: func(cmd *cobra.Command, args []string) error {
		if daemon.IsRunning(cfg) {
			return fmt.Errorf("wfh daemon is already running")
		}

		d, err := daemon.New(cfg)
		if err != nil {
			return fmt.Errorf("create daemon: %w", err)
		}

		fmt.Println("Starting wfh daemon...")
		return d.Start()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the wfh daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !daemon.IsRunning(cfg) {
			return fmt.Errorf("wfh daemon is not running")
		}

		fmt.Println("Stopping wfh daemon...")
		d, err := daemon.New(cfg)
		if err != nil {
			return err
		}
		return d.Stop()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		running := daemon.IsRunning(cfg)

		s, err := store.Open(cfg.DBFile)
		if err != nil {
			return err
		}
		defer s.Close()

		today, _ := s.GetTodaysActivity()
		var totalSec int64
		for _, a := range today {
			totalSec += a.Duration
		}
		totalDur := fmt.Sprintf("%dh %dm", int(totalSec/3600), int(totalSec%3600/60))

		var activeSince string
		if len(today) > 0 {
			activeSince = today[0].StartTime.Format("15:04")
		}

		renderer := reporter.NewTerminalRenderer()
		fmt.Print(renderer.RenderStatus(running, 0, activeSince, totalDur))
		return nil
	},
}

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's activity summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open(cfg.DBFile)
		if err != nil {
			return err
		}
		defer s.Close()

		now := cfgNow()
		date := now.Format("2006-01-02")

		r := reporter.NewReporter(s)
		report, err := r.GenerateDailyReport(date)
		if err != nil {
			return err
		}

		renderer := reporter.NewTerminalRenderer()
		fmt.Print(renderer.RenderDailyReport(report))
		return nil
	},
}

var weekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show weekly activity digest",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open(cfg.DBFile)
		if err != nil {
			return err
		}
		defer s.Close()

		r := reporter.NewReporter(s)
		report, err := r.GenerateWeeklyReport()
		if err != nil {
			return err
		}

		renderer := reporter.NewTerminalRenderer()
		fmt.Print(renderer.RenderWeeklyReport(report))
		return nil
	},
}

var reportCmd = &cobra.Command{
	Use:   "report [date]",
	Short: "Generate a detailed activity report for a specific date",
	Long: `Generate a report for a given date (YYYY-MM-DD).
If no date is provided, today's report is shown.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Open(cfg.DBFile)
		if err != nil {
			return err
		}
		defer s.Close()

		date := cfgNow().Format("2006-01-02")
		if len(args) > 0 {
			date = args[0]
		}

		r := reporter.NewReporter(s)
		report, err := r.GenerateDailyReport(date)
		if err != nil {
			return err
		}

		renderer := reporter.NewTerminalRenderer()
		fmt.Print(renderer.RenderDailyReport(report))
		return nil
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch [directory]",
	Short: "Watch a directory for file changes (one-shot)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cfg.RepoPath = args[0]
		}

		if cfg.RepoPath == "" {
			return fmt.Errorf("no directory specified and no repo_path in config")
		}

		s, err := store.Open(cfg.DBFile)
		if err != nil {
			return err
		}
		defer s.Close()

		fmt.Printf("Watching: %s\n", cfg.RepoPath)
		fmt.Println("Press Ctrl+C to stop")

		d, err := daemon.New(cfg)
		if err != nil {
			return err
		}
		return d.Start()
	},
}

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "Get or set configuration values",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Show current config
			fmt.Printf("DataDir:             %s\n", cfg.DataDir)
			fmt.Printf("DBFile:              %s\n", cfg.DBFile)
			fmt.Printf("LogFile:             %s\n", cfg.LogFile)
			fmt.Printf("RepoPath:            %s\n", cfg.RepoPath)
			fmt.Printf("PollIntervalSec:     %d\n", cfg.PollIntervalSec)
			fmt.Printf("HeartbeatTimeoutSec: %d\n", cfg.HeartbeatTimeoutSec)
			return nil
		}

		if len(args) == 2 {
			return fmt.Errorf("setting config values not yet implemented in CLI; edit %s directly", cfg.PidFile())
		}

		return nil
	},
}

// cfgNow returns the current time.
func cfgNow() time.Time {
	return time.Now()
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(todayCmd)
	rootCmd.AddCommand(weekCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(configCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
