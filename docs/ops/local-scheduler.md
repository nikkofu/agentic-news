# Local Scheduler Setup

## Daily Morning Run

The project provides a local run script and cron installer:

- `scripts/run-morning.sh`
- `scripts/install-cron.sh`

### Manual dry-run

```bash
bash scripts/run-morning.sh --dry-run
```

Expected output includes:

```text
daily-builder run date=today mode=morning dry_run=true
```

### Install cron job

```bash
bash scripts/install-cron.sh
```

This installs a cron entry running at `05:30` daily and writes logs to:

- `state/morning.log`

### Notes

- Ensure Go is available in your cron environment.
- For production use, set AI and SFTP env vars in shell profile loaded by cron.
