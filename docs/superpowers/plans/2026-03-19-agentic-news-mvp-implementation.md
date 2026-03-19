# Agentic News Butler MVP Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single-user, highly personalized daily RSS-to-H5 pipeline in Go that generates and publishes a mobile-first morning edition (before 07:00) with deep AI analysis and learning guidance.

**Architecture:** A local Go pipeline runs on schedule: ingest RSS, clean and dedupe content, generate multi-stage AI insights, rank top 10–20 personalized picks, render static H5 + JSON artifacts by date (`YYYY/MM/DD`), then publish to Nginx via SFTP with staging + atomic switch to `/latest`. Frontend is static and data-driven from generated JSON.

**Tech Stack:** Go 1.22+, standard library + selected libs (RSS parse, HTML extraction, SFTP), Go templates, static HTML/CSS/JS, local cron/launchd.

---

## File Structure Plan

### New files/directories to create

- Create: `go.mod`
- Create: `go.sum`
- Create: `cmd/daily-builder/main.go` (CLI entry)
- Create: `internal/config/config.go` (load/validate yaml + env)
- Create: `internal/config/types.go` (typed config models)
- Create: `internal/rss/fetcher.go` (RSS pull + parse)
- Create: `internal/rss/dedupe.go` (URL/content dedupe)
- Create: `internal/content/extractor.go` (article extraction/cleanup)
- Create: `internal/content/language.go` (zh/en detection)
- Create: `internal/model/types.go` (core domain structs)
- Create: `internal/analyze/client.go` (LLM API client wrapper)
- Create: `internal/analyze/prompts.go` (template loading/rendering)
- Create: `internal/analyze/pipeline.go` (A/B/C analysis stages)
- Create: `internal/rank/scoring.go` (weighted score + reweighting)
- Create: `internal/profile/profile.go` (explicit + implicit feedback model)
- Create: `internal/render/templates.go` (template setup)
- Create: `internal/render/index.go` (daily index render)
- Create: `internal/render/article.go` (detail page render)
- Create: `internal/output/writer.go` (date-path output writer)
- Create: `internal/publish/sftp.go` (upload + promote)
- Create: `internal/publish/latest.go` (`/latest` refresh)
- Create: `internal/verify/checks.go` (pre-publish checks)
- Create: `internal/run/pipeline.go` (orchestrator)
- Create: `web/templates/index.tmpl`
- Create: `web/templates/article.tmpl`
- Create: `web/static/styles.css`
- Create: `web/static/app.js`
- Create: `prompts/extract_facts.tmpl`
- Create: `prompts/deep_analysis.tmpl`
- Create: `prompts/personal_advisor.tmpl`
- Create: `prompts/final_editor.tmpl`
- Create: `config/rss_sources.yaml`
- Create: `config/scoring.yaml`
- Create: `config/ai.yaml`
- Create: `state/.gitkeep`
- Create: `output/.gitkeep`
- Create: `tests/rss/fetcher_test.go`
- Create: `tests/rss/dedupe_test.go`
- Create: `tests/content/extractor_test.go`
- Create: `tests/analyze/pipeline_test.go`
- Create: `tests/rank/scoring_test.go`
- Create: `tests/render/render_test.go`
- Create: `tests/verify/checks_test.go`
- Create: `tests/run/pipeline_test.go`
- Create: `scripts/run-morning.sh`
- Create: `scripts/install-cron.sh`
- Create: `docs/ops/local-scheduler.md`

### Existing files to reference

- Read/Reference: `docs/superpowers/specs/2026-03-19-agentic-news-butler-design.md`

---

### Task 1: Bootstrap Go project and config contracts

**Files:**
- Create: `go.mod`, `cmd/daily-builder/main.go`
- Create: `internal/config/types.go`, `internal/config/config.go`
- Create: `config/rss_sources.yaml`, `config/scoring.yaml`, `config/ai.yaml`
- Test: `tests/run/pipeline_test.go` (bootstrap test)

- [ ] **Step 1: Write failing test for config loading** (@superpowers:test-driven-development)

```go
func TestLoadConfig_ValidMinimalConfig(t *testing.T) {
    cfg, err := config.LoadConfig("testdata/config")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if cfg.AI.QualityMode != "high" {
        t.Fatalf("expected high quality mode")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tests/run -run TestLoadConfig_ValidMinimalConfig -v`
Expected: FAIL with missing `LoadConfig` implementation.

- [ ] **Step 3: Implement minimal config structs + loader**

```go
func LoadConfig(dir string) (Config, error) {
    var cfg Config
    // load ai/scoring/rss yaml, then validate required fields
    return cfg, nil
}
```

- [ ] **Step 4: Run focused test then full package tests**

Run: `go test ./tests/run -run TestLoadConfig_ValidMinimalConfig -v && go test ./...`
Expected: PASS for new config tests.

- [ ] **Step 5: Commit bootstrap slice**

```bash
git add go.mod cmd/daily-builder/main.go internal/config config tests/run config
git commit -m "feat: bootstrap Go CLI and typed config loading"
```

---

### Task 2: Implement RSS ingestion and deduplication

**Files:**
- Create: `internal/rss/fetcher.go`, `internal/rss/dedupe.go`
- Create: `internal/model/types.go`
- Test: `tests/rss/fetcher_test.go`, `tests/rss/dedupe_test.go`

- [ ] **Step 1: Write failing fetcher parser tests**

```go
func TestFetchFeeds_ParsesItems(t *testing.T) {
    items, err := rss.FetchFeeds(context.Background(), []string{testFeedURL})
    if err != nil { t.Fatal(err) }
    if len(items) == 0 { t.Fatal("expected items") }
}
```

- [ ] **Step 2: Run fetcher test and confirm fail**

Run: `go test ./tests/rss -run TestFetchFeeds_ParsesItems -v`
Expected: FAIL due to unimplemented fetch logic.

- [ ] **Step 3: Implement fetch + parse + timeout handling**

```go
func FetchFeeds(ctx context.Context, urls []string) ([]model.RawItem, error) {
    // pull each rss, parse, normalize timestamps, return merged items
}
```

- [ ] **Step 4: Write failing dedupe tests**

```go
func TestDedupe_RemovesCanonicalDuplicates(t *testing.T) {
    got := rss.Dedupe(items)
    if len(got) != 1 { t.Fatalf("expected 1 unique item") }
}
```

- [ ] **Step 5: Implement canonical-url + title-hash dedupe**

```go
func Dedupe(items []model.RawItem) []model.RawItem {
    // by canonical URL first, fallback to normalized title hash
}
```

- [ ] **Step 6: Run rss tests and commit**

Run: `go test ./tests/rss -v`
Expected: PASS.

```bash
git add internal/rss internal/model tests/rss
git commit -m "feat: add RSS ingestion and deduplication pipeline"
```

---

### Task 3: Build content extraction and language classification

**Files:**
- Create: `internal/content/extractor.go`, `internal/content/language.go`
- Test: `tests/content/extractor_test.go`

- [ ] **Step 1: Write failing extraction test**

```go
func TestExtractArticle_ReturnsCleanTextAndCover(t *testing.T) {
    article, err := content.ExtractArticle(ctx, raw)
    if err != nil { t.Fatal(err) }
    if article.ContentText == "" { t.Fatal("expected text") }
}
```

- [ ] **Step 2: Run test and verify fail**

Run: `go test ./tests/content -run TestExtractArticle_ReturnsCleanTextAndCover -v`
Expected: FAIL due to missing extractor.

- [ ] **Step 3: Implement minimal extractor and zh/en detector**

```go
func ExtractArticle(ctx context.Context, item model.RawItem) (model.Article, error) {
    // fetch target url content, strip boilerplate, pick cover image, detect language
}
```

- [ ] **Step 4: Run content tests**

Run: `go test ./tests/content -v`
Expected: PASS.

- [ ] **Step 5: Commit extraction slice**

```bash
git add internal/content tests/content internal/model
git commit -m "feat: implement article extraction and language detection"
```

---

### Task 4: Implement staged AI analysis pipeline

**Files:**
- Create: `internal/analyze/client.go`, `internal/analyze/prompts.go`, `internal/analyze/pipeline.go`
- Create: `prompts/extract_facts.tmpl`, `prompts/deep_analysis.tmpl`, `prompts/personal_advisor.tmpl`, `prompts/final_editor.tmpl`
- Test: `tests/analyze/pipeline_test.go`

- [ ] **Step 1: Write failing test for stage outputs**

```go
func TestAnalyzeArticle_ReturnsRequiredInsightFields(t *testing.T) {
    insight, err := analyze.RunPipeline(ctx, article, profile)
    if err != nil { t.Fatal(err) }
    if insight.SummaryDeep == "" || insight.Viewpoint == "" {
        t.Fatal("missing required insight fields")
    }
}
```

- [ ] **Step 2: Run analyze test to confirm fail**

Run: `go test ./tests/analyze -run TestAnalyzeArticle_ReturnsRequiredInsightFields -v`
Expected: FAIL due to unimplemented analysis pipeline.

- [ ] **Step 3: Implement prompt renderer + staged calls (A/B/C)**

```go
func RunPipeline(ctx context.Context, a model.Article, p profile.UserProfile) (model.Insight, error) {
    // stage A extract facts
    // stage B deep analysis
    // stage C personalized advisor
    // validate required JSON fields before return
}
```

- [ ] **Step 4: Add traceability metadata fields and schema guard**

Run: `go test ./tests/analyze -v`
Expected: PASS and trace fields present.

- [ ] **Step 5: Commit analysis slice**

```bash
git add internal/analyze prompts tests/analyze internal/model
git commit -m "feat: add staged AI analysis with prompt templates"
```

---

### Task 5: Implement scoring, source-tier confidence rules, and personalization

**Files:**
- Create: `internal/rank/scoring.go`, `internal/profile/profile.go`
- Modify: `internal/model/types.go`
- Test: `tests/rank/scoring_test.go`

- [ ] **Step 1: Write failing ranking formula test**

```go
func TestScoreItem_AppliesBaseWeights(t *testing.T) {
    score := rank.ScoreItem(input, rank.DefaultWeights())
    if score <= 0 { t.Fatal("expected positive score") }
}
```

- [ ] **Step 2: Run scoring test and confirm fail**

Run: `go test ./tests/rank -run TestScoreItem_AppliesBaseWeights -v`
Expected: FAIL due to missing ScoreItem.

- [ ] **Step 3: Implement weighted scoring + domain dynamic reweighting**

```go
func ScoreItem(s model.ScoreSignals, w Weights, category string) float64 {
    // apply base weighted sum then category adjustments
}
```

- [ ] **Step 4: Write and implement source-tier threshold tests (A/B/C)**

Run: `go test ./tests/rank -v`
Expected: PASS including tier confidence behavior.

- [ ] **Step 5: Implement explicit/implicit profile updates with deterministic merge**

Run: `go test ./...`
Expected: PASS for profile + ranking behavior.

- [ ] **Step 6: Commit ranking slice**

```bash
git add internal/rank internal/profile internal/model tests/rank
git commit -m "feat: add personalized ranking and source-tier confidence policy"
```

---

### Task 6: Render mobile H5 pages and date-based output artifacts

**Files:**
- Create: `internal/render/templates.go`, `internal/render/index.go`, `internal/render/article.go`, `internal/output/writer.go`
- Create: `web/templates/index.tmpl`, `web/templates/article.tmpl`, `web/static/styles.css`, `web/static/app.js`
- Test: `tests/render/render_test.go`

- [ ] **Step 1: Write failing render test for index + detail pages**

```go
func TestRenderDailyOutput_WritesIndexAndArticlePages(t *testing.T) {
    outDir := t.TempDir()
    err := render.DailyEdition(outDir, dailyData)
    if err != nil { t.Fatal(err) }
    assertFileExists(t, filepath.Join(outDir, "2026/03/19/index.html"))
}
```

- [ ] **Step 2: Run render test and verify fail**

Run: `go test ./tests/render -run TestRenderDailyOutput_WritesIndexAndArticlePages -v`
Expected: FAIL due to missing renderer/writer.

- [ ] **Step 3: Implement templates + output writer**

```go
func DailyEdition(baseDir string, daily model.DailyEdition) error {
    // create YYYY/MM/DD tree, render index/details, copy static assets, write data json/meta
}
```

- [ ] **Step 4: Ensure required card fields render (category/summary/score/cover/source/link/time)**

Run: `go test ./tests/render -v`
Expected: PASS with field assertions.

- [ ] **Step 5: Commit rendering slice**

```bash
git add internal/render internal/output web/templates web/static tests/render
git commit -m "feat: render mobile daily H5 and date-based artifacts"
```

---

### Task 7: Add SFTP publish with staging and latest switch

**Files:**
- Create: `internal/publish/sftp.go`, `internal/publish/latest.go`
- Test: `tests/verify/checks_test.go` (publish preconditions), `tests/run/pipeline_test.go` (publish flow)

- [ ] **Step 1: Write failing test for publish path mapping**

```go
func TestBuildRemotePaths_UsesDateAndLatestTargets(t *testing.T) {
    paths := publish.BuildRemotePaths(date)
    if paths.Staging == "" || paths.Latest == "" { t.Fatal("missing paths") }
}
```

- [ ] **Step 2: Run test and confirm fail**

Run: `go test ./tests/run -run TestBuildRemotePaths_UsesDateAndLatestTargets -v`
Expected: FAIL due to missing publish helpers.

- [ ] **Step 3: Implement SFTP uploader (staging upload + validate + promote)**

```go
func PublishEdition(ctx context.Context, localDir string, cfg config.SFTPConfig) error {
    // upload to /_staging/YYYYMMDD, validate, move to /YYYY/MM/DD, refresh /latest
}
```

- [ ] **Step 4: Add retryable error surfaces and idempotent behavior**

Run: `go test ./tests/run -v`
Expected: PASS for publish path and flow tests.

- [ ] **Step 5: Commit publish slice**

```bash
git add internal/publish tests/run tests/verify
git commit -m "feat: implement staged SFTP publishing and latest refresh"
```

---

### Task 8: Add pre-publish verification gate and orchestrator pipeline

**Files:**
- Create: `internal/verify/checks.go`, `internal/run/pipeline.go`
- Modify: `cmd/daily-builder/main.go`
- Test: `tests/verify/checks_test.go`, `tests/run/pipeline_test.go`

- [ ] **Step 1: Write failing verification gate test**

```go
func TestVerifyDailyEdition_RejectsMissingIndex(t *testing.T) {
    err := verify.DailyEdition(dir)
    if err == nil { t.Fatal("expected verification error") }
}
```

- [ ] **Step 2: Run verify test and confirm fail**

Run: `go test ./tests/verify -run TestVerifyDailyEdition_RejectsMissingIndex -v`
Expected: FAIL due to missing verifier.

- [ ] **Step 3: Implement checks (files, links, json schema, min picks>=10)**

```go
func DailyEdition(dir string) error {
    // run all checks and aggregate failures
}
```

- [ ] **Step 4: Implement orchestrator run flow + deadline guard (06:50 fallback)**

Run: `go test ./tests/run -v && go test ./tests/verify -v`
Expected: PASS and fallback path covered.

- [ ] **Step 5: Commit verify/orchestration slice**

```bash
git add internal/verify internal/run cmd/daily-builder tests/verify tests/run
git commit -m "feat: add verification gate and end-to-end run orchestrator"
```

---

### Task 9: Add scheduler scripts, ops docs, and end-to-end smoke test

**Files:**
- Create: `scripts/run-morning.sh`, `scripts/install-cron.sh`, `docs/ops/local-scheduler.md`
- Modify: `cmd/daily-builder/main.go` (flags/help)
- Test: `tests/run/pipeline_test.go` (smoke mode)

- [ ] **Step 1: Write failing test for CLI morning mode flags**

```go
func TestCLI_AcceptsMorningMode(t *testing.T) {
    err := runCLI([]string{"run", "--date", "today", "--mode", "morning"})
    if err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Run CLI test and verify fail**

Run: `go test ./tests/run -run TestCLI_AcceptsMorningMode -v`
Expected: FAIL due to missing flag wiring.

- [ ] **Step 3: Implement scripts and docs for local schedule install**

Run:
- `bash scripts/run-morning.sh --dry-run`
- `go test ./tests/run -v`
Expected: dry-run prints planned steps; tests pass.

- [ ] **Step 4: Run full verification-before-completion checklist** (@superpowers:verification-before-completion)

Run:
- `go test ./...`
- `go run ./cmd/daily-builder run --date 2026-03-19 --mode morning --dry-run`
Expected: all tests pass, dry-run completes with no fatal errors.

- [ ] **Step 5: Commit release-ready MVP baseline**

```bash
git add scripts docs/ops cmd/daily-builder tests
git commit -m "chore: add scheduler ops flow and verify MVP pipeline"
```

---

## Plan-level Verification Checklist

- [ ] All TDD cycles executed per task (fail → implement minimal → pass)
- [ ] All tests pass (`go test ./...`)
- [ ] Dry-run pipeline generates expected dated output tree
- [ ] Publish gate blocks broken editions
- [ ] SFTP staging + latest switch validated in test path
- [ ] Required H5 fields present: category/summary/score/cover/detail/source/link/time
- [ ] Personalization outputs present: why_for_you, taste_growth_hint, knowledge_gap_hint

## Notes for execution agent

- Keep implementations minimal (DRY/YAGNI).
- Prefer small, isolated commits per task.
- Do not introduce API server or multi-user features in MVP.
- If repository is not initialized, run without commit steps and preserve the same task boundaries.
