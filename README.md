# agentic-news

An AI news butler for personal cognitive growth. It transforms high-density RSS streams into high-quality daily mobile briefings with deep analysis, personalized ranking, and actionable learning guidance.

## Overview

`agentic-news` is a single-user, highly personalized daily news pipeline:

- Ingest RSS from tech, public affairs, finance, and expert opinion sources
- Clean, deduplicate, and normalize content
- Run staged AI analysis (facts → deep insight → personal guidance)
- Rank and select 10–20 featured items every day
- Render a clean mobile H5 edition
- Publish to Nginx via SFTP with dated archives and `/latest` entry

Goal: not just aggregation, but continuous improvement in taste, cognition, and knowledge structure.

## Core Features (MVP)

- Daily automated morning edition (before 07:00)
- Chinese + English input, Chinese-first output
- Deep commentary with evidence-linked insights
- Source-tier confidence thresholds
- Strong personalization with dual-track learning:
  - Explicit feedback (like/neutral/disagree)
  - Implicit behavior (click/dwell/revisit/bookmark)
- Static H5 output with required fields:
  - Category
  - Summary
  - Score
  - Cover image
  - Detail page
  - Source + original link
  - Publish time

## Architecture (MVP)

- **Local Builder (Go)**: ingest → analyze → rank → render → verify → publish
- **Cloud Nginx**: static hosting only

## Output Structure

```text
/YYYY/MM/DD/index.html
/YYYY/MM/DD/articles/{id}.html
/YYYY/MM/DD/assets/...
/YYYY/MM/DD/data/daily.json
/YYYY/MM/DD/data/learning.json
/YYYY/MM/DD/attachments/...
/YYYY/MM/DD/meta.json
/latest/...
```

## Repository Status

🚧 In active build-out from approved spec and implementation plan.

- Spec: `docs/superpowers/specs/2026-03-19-agentic-news-butler-design.md`
- Plan: `docs/superpowers/plans/2026-03-19-agentic-news-mvp-implementation.md`

## Planned Project Layout

```text
cmd/daily-builder/
internal/config/
internal/rss/
internal/content/
internal/analyze/
internal/rank/
internal/profile/
internal/render/
internal/output/
internal/publish/
internal/verify/
internal/run/
web/templates/
web/static/
prompts/
config/
state/
output/
docs/
scripts/
tests/
```

## Quick Start (planned)

```bash
# configure yaml files in config/
# set env vars for SFTP credentials

go run ./cmd/daily-builder run --date today --mode morning
```

## Roadmap

- [x] Product design spec
- [x] Implementation plan
- [ ] Go bootstrap
- [ ] RSS ingest + dedupe
- [ ] AI analysis pipeline
- [ ] Ranking + personalization
- [ ] H5 rendering + artifact output
- [ ] SFTP publish + latest switch
- [ ] Verification gate + scheduler scripts

## License

Apache-2.0 — see [LICENSE](./LICENSE).
