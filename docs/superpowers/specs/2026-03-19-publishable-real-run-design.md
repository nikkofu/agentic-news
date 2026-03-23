# Publishable Real-Run Design Spec

- **Date:** 2026-03-19
- **Project:** agentic-news
- **Scope:** Upgrade the local `run` pipeline from simulated orchestration to a real RSS-driven, publishable local edition with explicit downgrade semantics and no real SFTP transport.

## 1. Problem Statement

The current mainline can generate local sample editions and run a dry pipeline, but the `run` flow still relies on simulated candidate generation. That is not sufficient for a daily product that must produce a publishable edition from real RSS input.

The next milestone is to make `run` consume real RSS feeds and still generate a **publishable local edition** even when extraction or analysis quality is uneven. The system should avoid all-or-nothing failure for single-item issues, but it must still enforce a minimum quality bar for the final edition.

This iteration does **not** implement real SFTP upload or remote `/latest` switching. The publish boundary remains local artifact generation plus explicit deferred transport status.

## 2. Confirmed Product Decisions

1. `run` must use real RSS source input rather than simulated candidates.
2. The system should prefer successful full-analysis cards, but it must still produce a publishable edition when some items degrade.
3. If article extraction fails or analysis fails for a single item, the item should degrade rather than abort the full run.
4. The edition must remain honest about card quality; degraded items must not be presented as full deep-analysis items.
5. A publishable edition must contain at least:
   - `10` featured cards total
   - `3` standard cards minimum
6. If the pipeline cannot meet that threshold, the run must fail with a clear reason rather than claiming success.
7. Real SFTP transport remains out of scope for this iteration.

## 3. Card Model

Two card types are introduced for featured output:

### 3.1 Standard card

A standard card represents a fully processed item:

- RSS metadata available
- article extraction produced acceptable body text
- analysis pipeline completed and passed contract validation
- render output can safely show deep summary, viewpoint, and related AI fields

The card carries:

- `card_type = standard`
- full summary / score / source / time fields
- deep-analysis fields such as viewpoint

### 3.2 Brief card

A brief card represents a degraded but still publishable item:

- RSS metadata is available
- extraction failed, quality was too low, or analysis failed / returned incomplete output

The card carries:

- `card_type = brief`
- `fallback_reason`
- title, summary, source, original link, category, publish time
- no fake or placeholder deep-analysis treatment that would misrepresent quality

Brief cards are a deliberate fallback path, not an error case.

## 4. Pipeline Architecture

The real `run` pipeline becomes:

1. Load config
2. Fetch RSS feeds
3. Dedupe raw items
4. Build candidate set from real RSS items
5. Attempt article extraction per item
6. Attempt AI analysis for items with sufficient extracted quality
7. Convert each candidate into either:
   - standard card
   - brief card
8. Rank cards with type-aware ordering
9. Select featured cards
10. Render local edition
11. Verify publishability

### 4.1 Failure handling philosophy

- **Feed-level failure:** continue if other feeds still produce usable items
- **Item-level extraction failure:** downgrade that item to brief
- **Item-level analysis failure:** downgrade that item to brief
- **Edition-level failure:** only fail if final publishability thresholds are not met

This keeps the pipeline resilient while preserving an explicit minimum quality boundary.

## 5. Downgrade Rules

An item should become a brief card when any of the following are true:

- article fetch fails
- extracted content is too short or too noisy
- extraction falls back to weak text and does not meet the standard-card threshold
- analysis contract validation fails
- required analysis fields are missing

Brief cards should derive fields from the best available inputs in this order:

1. RSS title
2. RSS description / raw content
3. extracted partial text if available
4. original source metadata

The minimum brief-card contract is:

- title
- summary
- source name
- source URL
- published time
- category
- `card_type`
- `fallback_reason`

## 6. Ranking Rules

Ranking becomes type-aware:

1. Standard cards always rank ahead of brief cards
2. Within the same card type, use deterministic ordering:
   - score descending
   - published time descending
   - ID ascending

Selection logic:

- Prefer standard cards first
- If standard cards are fewer than 10, fill the remaining slots with brief cards
- Do not allow brief cards to displace standard cards when standard cards are available

This ensures degraded content is used only as publishability padding, not as a replacement for high-quality content.

## 7. Rendering Rules

Rendering must visually and structurally distinguish the two card types.

### 7.1 Standard cards

Show the normal card experience:

- category
- title
- summary
- score
- source
- publish time
- viewpoint / deep-analysis cues

### 7.2 Brief cards

Show a lighter fallback presentation:

- visible label such as `Brief` / `Fallback`
- summary from RSS metadata or fallback text
- source, original link, publish time
- downgrade reason on the detail page

Brief cards must **not** pretend to include deep AI insight if the analysis stage did not succeed.

## 8. Verification Rules

The verification gate should move from file-existence checks to publishability checks.

The edition passes verification only if:

- required files exist
- `daily.json` parses successfully
- `featured_count >= 10`
- `standard_count >= 3`
- every featured item has a detail page
- every standard card contains the required standard-card fields
- every brief card contains the required brief-card fields

The gate should fail with actionable reasons, for example:

- insufficient featured count
- insufficient standard-card count
- missing article page
- missing fallback reason on brief card

## 9. File-Level Design

### Core model

- Modify: `internal/model/types.go`
  - add `CardType`
  - add `FallbackReason`

### Extraction

- Modify: `internal/content/extractor.go`
  - expose enough quality status for `run` to decide standard vs brief

### Analysis

- Keep: `internal/analyze/pipeline.go`
  - retain strict validation
  - let `run` convert failures into brief-card downgrade decisions

### Ranking

- Modify: `internal/rank/scoring.go`
  - enforce standard-first ordering
  - keep deterministic tie-breaking

### Rendering

- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/article.tmpl`
  - distinguish standard vs brief presentation

### Verification

- Modify: `internal/verify/checks.go`
  - enforce publishability thresholds and per-card-type required fields

### Run orchestration

- Modify: `internal/run/pipeline.go`
  - replace simulated candidate generation with real RSS-driven flow
  - downgrade item-level failures to brief cards
  - fail only when publishability thresholds cannot be met

## 10. Non-Goals

This design does not include:

- real SFTP upload
- remote staging upload
- remote atomic `/latest` switching
- login or multi-user profile collection
- full feedback-learning loop

## 11. Success Criteria

This design is successful when:

1. `go run ./cmd/daily-builder run --date YYYY-MM-DD --mode morning --dry-run` uses real RSS inputs
2. extraction / analysis failures degrade individual items rather than collapsing the whole run
3. the system can generate a mixed `standard + brief` edition
4. the verification gate enforces `10` featured items minimum and `3` standard cards minimum
5. the edition clearly communicates which items are degraded
6. the run fails cleanly when publishability thresholds cannot be satisfied

## 12. Recommended Implementation Order

1. Add card-type model fields and publishability verification rules
2. Update rendering for standard vs brief cards
3. Update ranking to be type-aware
4. Replace simulated run orchestration with the real RSS-driven pipeline

This order minimizes integration risk by locking the publishability contract before the orchestration changes.
