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
dry-run daily pipeline generated at output/YYYY/MM/DD featured=10 fallback=<true|false>
```

### Install cron job

```bash
bash scripts/install-cron.sh
```

This installs a cron entry running at `05:30` daily and writes logs to:

- `state/morning.log`

### Notes

- Ensure Go is available in your cron environment.
- Current mainline flow uses real RSS input from `config/rss_sources.yaml` and generates/verifies local artifacts only.
- The default repository config points at a local RSS fixture and local HTML fixtures so dry-runs work offline.
- The builder reads `state/feedback/profile_snapshot.json` when present and uses it for ranking plus explanation copy on the next run.
- If `state/feedback/profile_snapshot.json` is missing, the builder still succeeds with deterministic scoring fallback and safe generic explanation copy.
- Same-day feedback state is materialized under `state/feedback/`:
  - `state/feedback/events/YYYY-MM.jsonl`
  - `state/feedback/profile_snapshot.json`
  - `state/feedback/learning_snapshot.json`
- A successful publishable local edition currently requires `10` featured cards with at least `3` `standard` cards; the remainder may be `brief` fallback cards.
- Real SFTP transport is explicitly deferred; use `go run ./cmd/daily-builder publish-sample --date YYYY-MM-DD --dry-run` to inspect remote path planning.
- For a later transport-enabled iteration, set AI and SFTP env vars in the shell profile loaded by cron.

## Local Feedback API

Run the co-located feedback service alongside the scheduler when you want same-day profile refresh:

```bash
go run ./cmd/feedback-api
```

Optional overrides:

```bash
AGENTIC_NEWS_FEEDBACK_ADDR=127.0.0.1:18081 AGENTIC_NEWS_STATE_DIR=/tmp/agentic-news-feedback go run ./cmd/feedback-api
```

Operational notes:

- Keep the feedback API and `daily-builder run` pointed at the same state root when you want the next run to consume fresh snapshots.
- Same-day panel refresh comes from the feedback API; cron-driven builder runs only consume the saved snapshots on the next generation pass.
