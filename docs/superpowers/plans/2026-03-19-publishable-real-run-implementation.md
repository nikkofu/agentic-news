# Publishable Real-Run Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade `run --dry-run` from a simulated pipeline to a real RSS-driven pipeline that can still generate a publishable local edition by mixing `standard` and `brief` cards when some items degrade.

**Architecture:** Keep the existing local-only mainline and deferred publish boundary, but replace simulated candidate generation with real RSS ingest, extraction, downgrade-aware analysis, type-aware ranking, and publishability verification. Standard cards represent fully extracted and analyzed items; brief cards represent honest fallback items that keep the edition publishable without pretending they are full deep-analysis pieces.

**Tech Stack:** Go 1.22+, stdlib, current RSS/content/analyze/render packages, Go templates, `go test`.

---

## File Structure Plan

### Core files to modify

- Modify: `internal/model/types.go`
- Modify: `internal/content/extractor.go`
- Modify: `internal/rank/scoring.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `internal/verify/checks.go`
- Modify: `internal/run/pipeline.go`
- Modify: `cmd/daily-builder/main.go`
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/article.tmpl`
- Modify: `README.md`
- Modify: `docs/ops/local-scheduler.md`

### Tests to modify/create

- Modify: `tests/content/extractor_test.go`
- Modify: `tests/rank/scoring_test.go`
- Modify: `tests/render/render_test.go`
- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/verify/checks_test.go`

### Existing files to reference

- Reference: `docs/superpowers/specs/2026-03-19-publishable-real-run-design.md`
- Reference: `internal/rss/fetcher.go`
- Reference: `internal/rss/dedupe.go`
- Reference: `internal/analyze/pipeline.go`

---

### Task 1: Add card-type model fields and publishability verification rules

**Files:**
- Modify: `internal/model/types.go`
- Modify: `internal/verify/checks.go`
- Test: `tests/verify/checks_test.go`

- [ ] **Step 1: Write failing tests for card-type-aware verification**

```go
func TestDailyEdition_FailsWhenStandardCardCountBelowThreshold(t *testing.T) {
    // 10 featured items but only 2 standard cards should fail
}

func TestDailyEdition_FailsWhenBriefCardMissingFallbackReason(t *testing.T) {
    // brief item without fallback_reason should fail
}
```

- [ ] **Step 2: Run verify tests to confirm red**

Run: `go test ./tests/verify -run 'TestDailyEdition_(FailsWhenStandardCardCountBelowThreshold|FailsWhenBriefCardMissingFallbackReason)' -v`
Expected: FAIL because current verifier only checks file existence and featured count.

- [ ] **Step 3: Add card-type fields and minimum publishability checks**

```go
type DailyPick struct {
    ID             string
    CardType       string
    FallbackReason string
    // existing fields...
}

// verify:
// - featured >= 10
// - standard_count >= 3
// - brief cards require fallback_reason + source_url + summary
// - standard cards require viewpoint + summary + source fields
```

- [ ] **Step 4: Run full verify package**

Run: `go test ./tests/verify -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/model/types.go internal/verify/checks.go tests/verify/checks_test.go
git commit -m "feat: enforce publishable edition card-type verification rules"
```

---

### Task 2: Render standard and brief cards honestly

**Files:**
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/article.tmpl`
- Test: `tests/render/render_test.go`

- [ ] **Step 1: Write failing tests for brief-card rendering**

```go
func TestRenderDailyOutput_LabelsBriefCardsAndShowsFallbackReason(t *testing.T) {
    // index should show brief marker
    // article should show fallback reason and omit fake AI viewpoint copy
}
```

- [ ] **Step 2: Run render tests to confirm red**

Run: `go test ./tests/render -run TestRenderDailyOutput_LabelsBriefCardsAndShowsFallbackReason -v`
Expected: FAIL because current render path treats every card the same.

- [ ] **Step 3: Add template data shaping for card-type-specific output**

```go
type featuredCardData struct {
    CardType       string
    FallbackReason string
    ShowViewpoint  bool
    BadgeLabel     string
}

// standard => normal viewpoint presentation
// brief => brief badge + fallback reason + original link only
```

- [ ] **Step 4: Run render package**

Run: `go test ./tests/render -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/index.go internal/render/article.go web/templates/index.tmpl web/templates/article.tmpl tests/render/render_test.go
git commit -m "feat: distinguish standard and brief cards in rendered output"
```

---

### Task 3: Make ranking prefer standard cards before brief cards

**Files:**
- Modify: `internal/rank/scoring.go`
- Test: `tests/rank/scoring_test.go`

- [ ] **Step 1: Write failing test for card-type priority**

```go
func TestRankItems_PrioritizesStandardCardsBeforeBriefCards(t *testing.T) {
    items := []model.DailyPick{
        {ID: "brief-1", CardType: "brief", ScoreFinal: 99},
        {ID: "standard-1", CardType: "standard", ScoreFinal: 70},
    }
    got := rank.RankItems(items)
    if got[0].ID != "standard-1" { t.Fatal("expected standard card first") }
}
```

- [ ] **Step 2: Run rank tests to confirm red**

Run: `go test ./tests/rank -run TestRankItems_PrioritizesStandardCardsBeforeBriefCards -v`
Expected: FAIL because current ordering ignores card type.

- [ ] **Step 3: Implement type-aware deterministic ranking**

```go
// compare:
// 1) card_type priority: standard before brief
// 2) score desc
// 3) published_at desc
// 4) id asc
```

- [ ] **Step 4: Run rank package**

Run: `go test ./tests/rank -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/rank/scoring.go tests/rank/scoring_test.go
git commit -m "feat: rank standard cards ahead of brief fallbacks"
```

---

### Task 4: Expose extraction quality so run can downgrade items

**Files:**
- Modify: `internal/content/extractor.go`
- Test: `tests/content/extractor_test.go`

- [ ] **Step 1: Write failing tests for extraction quality classification**

```go
func TestExtractArticle_FlagsWeakFallbackContent(t *testing.T) {
    // body fallback with very thin content should return downgrade metadata
}

func TestExtractArticle_FlagsStrongParagraphExtractionAsStandardEligible(t *testing.T) {
    // normal article html should remain standard-eligible
}
```

- [ ] **Step 2: Run content tests to confirm red**

Run: `go test ./tests/content -run 'TestExtractArticle_(FlagsWeakFallbackContent|FlagsStrongParagraphExtractionAsStandardEligible)' -v`
Expected: FAIL because current extractor returns only `Article` with no quality status.

- [ ] **Step 3: Add extraction status output**

```go
type ExtractionStatus struct {
    StandardEligible bool
    UsedFallbackText bool
    FallbackReason   string
}

func ExtractArticle(ctx context.Context, item model.RawItem) (model.Article, ExtractionStatus, error) {
    // keep strict fetch behavior
    // classify text quality for standard vs brief decision
}
```

- [ ] **Step 4: Run content package**

Run: `go test ./tests/content -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/content/extractor.go tests/content/extractor_test.go
git commit -m "feat: expose extraction quality for brief-card downgrade decisions"
```

---

### Task 5: Replace simulated run flow with real RSS-driven card construction

**Files:**
- Modify: `internal/run/pipeline.go`
- Reference: `internal/rss/fetcher.go`
- Reference: `internal/rss/dedupe.go`
- Reference: `internal/analyze/pipeline.go`
- Test: `tests/run/pipeline_test.go`

- [ ] **Step 1: Write failing tests for real run behavior**

```go
func TestRunPipeline_ProducesPublishableMixedEditionFromRealRSS(t *testing.T) {
    // fetch real raw items (stubbed via hooks), extract, analyze
    // standard cards should be preferred
    // brief cards should backfill when needed
}

func TestRunPipeline_FailsWhenPublishabilityThresholdCannotBeMet(t *testing.T) {
    // fewer than 3 standard cards or fewer than 10 total should fail
}
```

- [ ] **Step 2: Run run tests to confirm red**

Run: `go test ./tests/run -run 'TestRunPipeline_(ProducesPublishableMixedEditionFromRealRSS|FailsWhenPublishabilityThresholdCannotBeMet)' -v`
Expected: FAIL because current run path still uses synthetic ingest/analyze defaults.

- [ ] **Step 3: Replace synthetic defaults with real flow and downgrade rules**

```go
type DryRunHooks struct {
    FetchFeeds func(ctx context.Context, urls []string) ([]model.RawItem, error)
    Dedupe     func([]model.RawItem) []model.RawItem
    Extract    func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error)
    Analyze    func(ctx context.Context, article model.Article, profile profile.UserProfile) (model.Insight, error)
}

// default run path:
// 1) collect rss urls from config
// 2) fetch + dedupe raw items
// 3) build standard or brief DailyPick per item
// 4) rank
// 5) select top 10 while preserving standard-first ordering
// 6) render + verify
```

- [ ] **Step 4: Run run package**

Run: `go test ./tests/run -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/run/pipeline.go tests/run/pipeline_test.go
git commit -m "feat: wire real rss run pipeline with standard and brief card output"
```

---

### Task 6: Sync CLI/docs and run final verification

**Files:**
- Modify: `cmd/daily-builder/main.go`
- Modify: `README.md`
- Modify: `docs/ops/local-scheduler.md`

- [ ] **Step 1: Update CLI/user-facing wording for publishable mixed-card runs**

Include:
- real RSS-driven `run --dry-run`
- publishability thresholds (`10` featured, `3` standard)
- brief-card downgrade behavior

- [ ] **Step 2: Run full test suite**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 3: Run manual command verification**

Run:
- `go run ./cmd/daily-builder run --date 2026-03-19 --mode morning --dry-run`
- `go run ./cmd/daily-builder publish-sample --date 2026-03-19 --dry-run`

Expected:
- `run` reports a local edition path and whether fallback was used
- `publish-sample --dry-run` still prints deterministic path planning

- [ ] **Step 4: Commit**

```bash
git add cmd/daily-builder/main.go README.md docs/ops/local-scheduler.md
git commit -m "docs: describe publishable real-run workflow and brief-card fallback behavior"
```

---

## Completion Gate

- [ ] `go test ./...` clean pass
- [ ] `run --dry-run` uses real RSS fetch + dedupe rather than synthetic candidates
- [ ] Edition can publish with mixed `standard + brief` cards
- [ ] Verification enforces `featured >= 10`
- [ ] Verification enforces `standard >= 3`
- [ ] Brief cards render with visible downgrade semantics
- [ ] `publish-sample --dry-run` remains deterministic and deferred
