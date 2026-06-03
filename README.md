# wfh — Work From Home Activity Tracker

> **Privacy-first**: All data stored locally. Zero telemetry. No cloud sync.
> Your activity data never leaves your machine.

wfh is a CLI tool that tracks your development activity throughout the day.
It monitors git commits, file changes, and editor usage to generate
personal productivity reports — all without sending data anywhere.

## Features

- **Git Activity Tracking** — detects commits, branch switches, and merges
- **File Change Monitoring** — watches your repository for file modifications
- **Editor Detection** — identifies active editor sessions
- **Daily Reports** — see how you spent your time each day
- **Weekly Digests** — get a birds-eye view of your week
- **Background Daemon** — runs silently, tracks continuously
- **Privacy Focused** — no accounts, no cloud, no telemetry
- **Beautiful Terminal Output** — styled with lipgloss

## Installation

### From source

```bash
git clone https://github.com/zinuo-xu/wfh.git
cd wfh
make build
sudo make install
```

### Using Go install

```bash
go install github.com/zinuo-xu/wfh/cmd/wfh@latest
```

### macOS (Homebrew)

```bash
brew tap zinuo-xu/wfh
brew install wfh
```

### Linux / macOS script

```bash
curl -sfL https://raw.githubusercontent.com/zinuo-xu/wfh/main/scripts/install.sh | sh
```

## Quick Start

```bash
# Start the tracking daemon
wfh start

# Check daemon status
wfh status

# View today's activity
wfh today

# View weekly digest
wfh week

# Watch a specific directory
wfh watch /path/to/your/project

# Stop the daemon
wfh stop
```

## Commands

| Command   | Description                                |
|-----------|--------------------------------------------|
| `start`   | Start the background tracking daemon        |
| `stop`    | Stop the daemon                             |
| `status`  | Show daemon running status and daily totals |
| `today`   | Display today's activity summary            |
| `week`    | Show weekly activity digest                 |
| `report`  | Generate detailed report for any date       |
| `watch`   | Watch a directory for file changes          |
| `config`  | View or update configuration                |

## Configuration

Configuration is stored at `~/.config/wfh/config.json`:

```json
{
  "data_dir": "~/.config/wfh",
  "db_file": "~/.config/wfh/wfh.db",
  "log_file": "~/.config/wfh/wfh.log",
  "repo_path": "/path/to/your/project",
  "poll_interval_sec": 30,
  "heartbeat_timeout_sec": 300
}
```

## Data Privacy

wfh is built with privacy as its core principle:

- **All data is stored locally** on your machine in `~/.config/wfh/`
- **No telemetry** — no usage data, no crash reports are sent anywhere
- **No cloud sync** — your activity data never leaves your computer
- **No accounts required** — just install and run
- **Full data control** — delete your data by removing `~/.config/wfh/`
- **Open source** — inspect every line of code; no secrets, no surprises

## Architecture

```
wfh/
├── cmd/wfh/         # CLI entry point (Cobra)
├── internal/
│   ├── config/      # Configuration management
│   ├── daemon/      # Background process lifecycle
│   │   ├── daemon.go     # Process management
│   │   ├── watcher.go    # fsnotify file watcher
│   │   └── heartbeat.go  # Active time tracking
│   ├── tracker/     # Activity detection engines
│   │   ├── tracker.go    # Coordinator
│   │   ├── git.go        # Git activity
│   │   ├── editor.go     # Editor detection
│   │   └── browser.go    # Code review detection
│   ├── store/       # SQLite data layer
│   └── reporter/    # Report generation (lipgloss)
└── scripts/         # Installation helpers
```

## License

MIT — see [LICENSE](LICENSE)
