#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CRON_LINE="30 5 * * * cd $PROJECT_DIR && ./scripts/run-morning.sh >> $PROJECT_DIR/state/morning.log 2>&1"

mkdir -p "$PROJECT_DIR/state"

( crontab -l 2>/dev/null | grep -v "run-morning.sh"; echo "$CRON_LINE" ) | crontab -

echo "Installed cron job: $CRON_LINE"
