#!/usr/bin/env bash
set -euo pipefail

DATE="today"
MODE="morning"
DRY_RUN="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --date)
      DATE="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN="true"
      shift
      ;;
    *)
      echo "Unknown arg: $1"
      exit 1
      ;;
  esac
done

cmd=(go run ./cmd/daily-builder run --date "$DATE" --mode "$MODE")
if [[ "$DRY_RUN" == "true" ]]; then
  cmd+=(--dry-run)
fi

"${cmd[@]}"
