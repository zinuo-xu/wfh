# Changelog

All notable changes to wfh are documented here.

## [0.1.0] — 2024-01-15

### Added

- Initial release of wfh
- Background daemon with start/stop/status
- Git activity tracking (commits, branches, merges)
- File change monitoring via fsnotify
- Editor detection (VS Code, Cursor, JetBrains, Neovim, Vim, Emacs, Zed)
- Browser/PR detection for code review tracking
- SQLite data layer with automatic migrations
- Daily activity reports with category breakdown
- Weekly digests with day-by-day comparison
- Beautiful terminal output with lipgloss styling
- Privacy-first architecture: all data local, no telemetry
- Configuration via `~/.config/wfh/config.json`
- Makefile for build and development
- CI/CD with GitHub Actions and GoReleaser
