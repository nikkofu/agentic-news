# Mainline Core Completion (SFTP Deferred) Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete and harden all core daily-news pipeline functions (ingest → analyze → rank → render → verify → local generation) while explicitly deferring real SFTP upload work.

**Architecture:** Keep the current static-edition architecture and strengthen each core module with production-grade behavior and test coverage. Preserve the CLI flow (`run`, `sample`, `publish-sample --dry-run`) but treat publishing as path planning only. Focus on deterministic outputs, schema stability, robust validation, and category/selection quality.

**Tech Stack:** Go 1.22+, stdlib, `gopkg.in/yaml.v3`, Go templates, static HTML/CSS/JS, `go test`.

---

## Scope Adjustment (confirmed)

- **Deferred:** real SFTP transport, remote authentication, remote atomic switch operations
- **In scope now:** everything else on the mainline path
  - RSS ingest reliability
  - extraction quality
  - analysis data contract
  - scoring/ranking correctness
  - dated artifact integrity
  - verification gates
  - deterministic sample/daily generation
  - CLI usability for daily workflow

---

## File Structure Plan (This Iteration)

### Core files to modify

- Modify: `internal/rss/fetcher.go` (error handling + feed-level tolerance)
- Modify: `internal/rss/dedupe.go` (canonicalization robustness)
- Modify: `internal/content/extractor.go` (text quality + cover fallback)
- Modify: `internal/analyze/pipeline.go` (strict required output checks)
- Modify: `internal/rank/scoring.go` (stable weighting + tie behavior)
- Modify: `internal/render/index.go` (required field rendering guarantees)
- Modify: `internal/render/article.go` (detail integrity)
- Modify: `internal/verify/checks.go` (stronger gate checks)
- Modify: `internal/run/pipeline.go` (orchestration behaviors)
- Modify: `internal/run/sample.go` (sample realism + deterministic IDs)
- Modify: `cmd/daily-builder/main.go` (CLI guardrails/messages)

### Publish boundary files (keep minimal)

- Modify: `internal/publish/sftp.go` (return clear deferred status path; no transport)
- Modify: `internal/publish/latest.go` (path correctness only)

### Tests to create/extend

- Modify/Create: `tests/rss/fetcher_test.go`
- Modify/Create: `tests/rss/dedupe_test.go`
- Modify/Create: `tests/content/extractor_test.go`
- Modify/Create: `tests/analyze/pipeline_test.go`
- Modify/Create: `tests/rank/scoring_test.go`
- Modify/Create: `tests/render/render_test.go`
- Modify/Create: `tests/verify/checks_test.go`
- Modify/Create: `tests/run/pipeline_test.go`
- Modify/Create: `tests/sample/sample_generation_test.go`

---

### Task 1: Harden RSS ingest reliability (no full-stop on partial feed failures)

**Files:**
- Modify: `internal/rss/fetcher.go`
- Test: `tests/rss/fetcher_test.go`

- [ ] **Step 1: Write failing test for partial feed failure tolerance**

```go
func TestFetchFeeds_ContinuesWhenOneFeedFails(t *testing.T) {
    items, err := rss.FetchFeeds(context.Background(), []string{goodURL, badURL})
    require.NoError(t, err)
    require.NotEmpty(t, items)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/rss -run TestFetchFeeds_ContinuesWhenOneFeedFails -v`
Expected: FAIL because current implementation returns error early.

- [ ] **Step 3: Implement minimal tolerant behavior**

```go
// collect per-feed errors, continue processing others
// return items + aggregated warning only when all feeds fail
```

- [ ] **Step 4: Run rss tests**

Run: `go test ./tests/rss -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/rss/fetcher.go tests/rss/fetcher_test.go
git commit -m "fix: make rss ingestion tolerant to partial feed failures"
```

---

### Task 2: Improve dedupe normalization and canonical URL handling

**Files:**
- Modify: `internal/rss/dedupe.go`
- Test: `tests/rss/dedupe_test.go`

- [ ] **Step 1: Write failing test for URL canonicalization**

```go
func TestDedupe_NormalizesQueryAndTrailingSlash(t *testing.T) {
    // same article URL variants should collapse
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/rss -run TestDedupe_NormalizesQueryAndTrailingSlash -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal URL canonicalizer**

```go
// normalize scheme/host/path, drop tracking query params, trim trailing slash
```

- [ ] **Step 4: Run rss tests**

Run: `go test ./tests/rss -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/rss/dedupe.go tests/rss/dedupe_test.go
git commit -m "fix: normalize canonical urls in dedupe pipeline"
```

---

### Task 3: Strengthen extraction quality and fallback behavior

**Files:**
- Modify: `internal/content/extractor.go`
- Test: `tests/content/extractor_test.go`

- [ ] **Step 1: Write failing test for no-paragraph fallback extraction**

```go
func TestExtractArticle_FallsBackWhenNoParagraphTags(t *testing.T) {
    // html with div-based body should still produce content text
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/content -run TestExtractArticle_FallsBackWhenNoParagraphTags -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal fallback extraction path**

```go
// if <p> extraction empty, use cleaned body text fallback
```

- [ ] **Step 4: Run content tests**

Run: `go test ./tests/content -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/content/extractor.go tests/content/extractor_test.go
git commit -m "fix: improve extraction fallback for non-paragraph article html"
```

---

### Task 4: Enforce analysis contract completeness

**Files:**
- Modify: `internal/analyze/pipeline.go`
- Test: `tests/analyze/pipeline_test.go`

- [ ] **Step 1: Write failing test for required output fields**

```go
func TestRunPipeline_RejectsMissingTraceabilityFields(t *testing.T) {
    // simulate incomplete insight and expect validation error
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/analyze -run TestRunPipeline_RejectsMissingTraceabilityFields -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal strict validation**

```go
// validate summary, viewpoint, confidence range, traceability fields
```

- [ ] **Step 4: Run analyze tests**

Run: `go test ./tests/analyze -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/analyze/pipeline.go tests/analyze/pipeline_test.go
git commit -m "fix: enforce strict insight contract validation"
```

---

### Task 5: Stabilize scoring/ranking and deterministic tie ordering

**Files:**
- Modify: `internal/rank/scoring.go`
- Test: `tests/rank/scoring_test.go`

- [ ] **Step 1: Write failing test for deterministic ordering on equal scores**

```go
func TestRankItems_DeterministicTieBreakByPublishedAt(t *testing.T) {
    // same score -> newer first
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/rank -run TestRankItems_DeterministicTieBreakByPublishedAt -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal tie-break strategy**

```go
// sort by score desc, then published_at desc, then id asc
```

- [ ] **Step 4: Run rank tests**

Run: `go test ./tests/rank -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/rank/scoring.go tests/rank/scoring_test.go
git commit -m "feat: add deterministic ranking tie-break behavior"
```

---

### Task 6: Enforce required H5 field presence in rendering output

**Files:**
- Modify: `internal/render/index.go`, `internal/render/article.go`
- Test: `tests/render/render_test.go`

- [ ] **Step 1: Write failing test for required card fields in rendered HTML**

```go
func TestRenderDailyOutput_ContainsRequiredFields(t *testing.T) {
    // assert category/summary/score/source/link/time strings exist in output html
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/render -run TestRenderDailyOutput_ContainsRequiredFields -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal render guarantees**

```go
// ensure template data always includes required fields with safe fallbacks
```

- [ ] **Step 4: Run render tests**

Run: `go test ./tests/render -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/index.go internal/render/article.go tests/render/render_test.go
git commit -m "fix: guarantee required daily card fields in rendered html"
```

---

### Task 7: Upgrade verification gates for mainline completeness

**Files:**
- Modify: `internal/verify/checks.go`
- Test: `tests/verify/checks_test.go`

- [ ] **Step 1: Write failing test for minimum featured count check**

```go
func TestDailyEdition_FailsWhenFeaturedCountBelowThreshold(t *testing.T) {
    // daily.json with <10 entries should fail verification
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/verify -run TestDailyEdition_FailsWhenFeaturedCountBelowThreshold -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal additional checks**

```go
// check index/meta/daily.json exist + featured count >= 10 + article files exist
```

- [ ] **Step 4: Run verify tests**

Run: `go test ./tests/verify -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/verify/checks.go tests/verify/checks_test.go
git commit -m "feat: enforce minimum featured count and article presence checks"
```

---

### Task 8: Complete run orchestration for local mainline flow

**Files:**
- Modify: `internal/run/pipeline.go`, `cmd/daily-builder/main.go`
- Test: `tests/run/pipeline_test.go`

- [ ] **Step 1: Write failing test for run flow dry-run orchestration**

```go
func TestRunPipeline_DryRunExecutesCoreStages(t *testing.T) {
    // config -> ingest -> analyze -> rank -> render -> verify, no publish transport
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/run -run TestRunPipeline_DryRunExecutesCoreStages -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal orchestrator stage wiring**

```go
// explicit stage sequence with clear errors and deadline fallback behavior
```

- [ ] **Step 4: Run run tests**

Run: `go test ./tests/run -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/run/pipeline.go cmd/daily-builder/main.go tests/run/pipeline_test.go
git commit -m "feat: wire core dry-run orchestration stages for daily pipeline"
```

---

### Task 9: Keep publish boundary explicit as deferred

**Files:**
- Modify: `internal/publish/sftp.go`
- Test: `tests/publish/publish_test.go`

- [ ] **Step 1: Write failing test for deferred transport status messaging**

```go
func TestPublishEdition_ReturnsDeferredTransportHint(t *testing.T) {
    // when invoked in current mode, return clear deferred/placeholder semantics
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/publish -run TestPublishEdition_ReturnsDeferredTransportHint -v`
Expected: FAIL.

- [ ] **Step 3: Implement minimal explicit deferred behavior**

```go
// preserve path planning output and return structured "transport_deferred" note
```

- [ ] **Step 4: Run publish tests**

Run: `go test ./tests/publish -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/publish/sftp.go tests/publish/publish_test.go
git commit -m "chore: mark publish transport as deferred while preserving path planning"
```

---

### Task 10: Final mainline verification and docs sync

**Files:**
- Modify: `README.md` (status + commands)
- Modify: `docs/ops/local-scheduler.md` (deferred SFTP note)

- [ ] **Step 1: Write failing test for sample command docs reference (if doc test exists), else skip with note**

Run: `go test ./...`
Expected: baseline remains green.

- [ ] **Step 2: Update docs with exact supported commands**

Include:
- `go run ./cmd/daily-builder sample YYYY-MM-DD`
- `go run ./cmd/daily-builder publish-sample --date YYYY-MM-DD --dry-run`
- Explicit note: real SFTP transport deferred

- [ ] **Step 3: Run full verification checklist** (@superpowers:verification-before-completion)

Run:
- `go test ./...`
- `go run ./cmd/daily-builder sample 2026-03-19`
- `go run ./cmd/daily-builder publish-sample --date 2026-03-19 --dry-run`

Expected:
- All tests pass
- sample output generated
- dry-run prints local/staging/dated/latest paths

- [ ] **Step 4: Commit**

```bash
git add README.md docs/ops/local-scheduler.md
git commit -m "docs: align mainline workflow and deferred sftp boundary"
```

---

## Completion Gate (Must pass before claiming “mainline complete”)

- [ ] `go test ./...` clean pass
- [ ] Core dry-run orchestration test pass
- [ ] Required rendered fields present and verified by tests
- [ ] Verification gate enforces min featured count and artifact integrity
- [ ] Sample generation works for specified date
- [ ] `publish-sample --dry-run` outputs deterministic path plan
- [ ] README + ops docs reflect deferred SFTP boundary clearly
