# Digital Sanctuary Reflection and LLM Follow-Up Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the post-core Sanctuary slice covering Reflection plus cached LLM enhancement for article detail, growth guidance, and reflection summaries without making any read path depend on live LLM success.

**Architecture:** Extend `internal/sanctuary` with a reflection store backed by JSONL + index files under `state/sanctuary/reflections/`, then add a file-backed enhancement cache under `state/sanctuary/cache/` plus a small provider abstraction that can reuse existing analysis/prompt infrastructure where helpful. Read handlers keep returning deterministic package/state data immediately and only attach cached enhancement blocks or `enhancement_status` metadata.

**Tech Stack:** Go 1.22+, stdlib `net/http` / `httptest` / `encoding/json`, existing `internal/analyze`, file-backed JSON/JSONL cache and state, `go test`.

**Execution Notes:** This plan starts only after the core Sanctuary plan is green. Follow @superpowers:test-driven-development and @superpowers:verification-before-completion. Every read endpoint must continue returning `200` without a live model call; any generation triggered from writes must persist first and degrade safely.

---

## File Structure Plan

### New files

- Create: `internal/sanctuary/reflection.go`
  - reflection list/create handlers and summary/archive read models
- Create: `internal/sanctuary/llm.go`
  - provider abstraction and background-trigger orchestration
- Create: `internal/sanctuary/cache.go`
  - file-backed enhancement cache helpers under `state/sanctuary/cache/`
- Create: `tests/sanctuary/reflection_test.go`
  - reflection read/write tests
- Create: `tests/sanctuary/llm_test.go`
  - cache, fallback, and non-blocking enhancement behavior tests

### Existing files to modify

- Modify: `internal/sanctuary/types.go`
  - reflection DTOs and enhancement status shapes
- Modify: `internal/sanctuary/store.go`
  - reflection JSONL/index persistence plus cache helpers
- Modify: `internal/sanctuary/http.go`
  - register reflection routes and any enhancement-aware JSON helpers
- Modify: `internal/sanctuary/content.go`
  - attach cached article enhancement blocks and statuses
- Modify: `internal/sanctuary/growth.go`
  - attach cached growth explanation blocks and statuses
- Modify: `tests/sanctuary/content_test.go`
  - assert cached enhancement behavior on article reads
- Modify: `tests/sanctuary/growth_test.go`
  - assert growth responses stay deterministic when cache is empty
- Modify: `README.md`
  - add reflection and enhancement runtime notes

### Existing files to reference

- Reference: `docs/superpowers/specs/2026-03-21-digital-sanctuary-platform-design.md`
- Reference: `internal/analyze/client.go`
- Reference: `internal/analyze/prompts.go`
- Reference: `internal/sanctuary/content.go`
- Reference: `internal/sanctuary/growth.go`

---

### Task 1: Add reflection persistence and `/api/v1/reflections` endpoints

**Files:**
- Modify: `internal/sanctuary/types.go`
- Modify: `internal/sanctuary/store.go`
- Modify: `internal/sanctuary/http.go`
- Create: `internal/sanctuary/reflection.go`
- Create: `tests/sanctuary/reflection_test.go`

- [ ] **Step 1: Write failing reflection tests**

```go
func TestSanctuaryAPI_PostAndListReflections(t *testing.T) {
	handler := sanctuary.NewHandler(sanctuary.NewService(seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)), t.TempDir()))

	postJSON(t, handler, http.MethodPost, "/api/v1/reflections", `{
		"content":"今天对 AI infra 的资本开支节奏理解更清晰了",
		"tags":["ai-infrastructure"],
		"related_article_ids":["pick-01"]
	}`, http.StatusCreated)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflections", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"items"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run the focused reflection tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_PostAndListReflections$' -v`
Expected: FAIL because reflection routes and storage are not implemented yet.

- [ ] **Step 3: Implement reflection JSONL + index persistence and handlers**

```go
type ReflectionEntry struct {
	ReflectionID      string   `json:"reflection_id"`
	CreatedAt         string   `json:"created_at"`
	Content           string   `json:"content"`
	Tags              []string `json:"tags"`
	RelatedArticleIDs []string `json:"related_article_ids"`
	RelatedDomains    []string `json:"related_domains"`
	Summary           string   `json:"summary"`
	EnhancementStatus string   `json:"enhancement_status"`
}
```

Implementation notes:
- Raw writes append to `state/sanctuary/reflections/YYYY-MM.jsonl`.
- `index.json` keeps a lightweight newest-first list of `reflection_id`, `created_at`, `summary`, and tag/domain pointers for listing.
- `POST /api/v1/reflections` must persist first, then optionally schedule summary/tag enhancement work afterward.
- `GET /api/v1/reflections` should paginate newest-first and always return deterministic fields even before summary enhancement exists.

- [ ] **Step 4: Re-run the focused reflection tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_PostAndListReflections$' -v`
Expected: PASS with reflection state working.

- [ ] **Step 5: Commit**

```bash
git add internal/sanctuary/types.go internal/sanctuary/store.go internal/sanctuary/http.go internal/sanctuary/reflection.go tests/sanctuary/reflection_test.go
git commit -m "feat: add sanctuary reflection api"
```

---

### Task 2: Add file-backed enhancement cache and provider abstraction

**Files:**
- Create: `internal/sanctuary/llm.go`
- Create: `internal/sanctuary/cache.go`
- Modify: `internal/sanctuary/store.go`
- Create: `tests/sanctuary/llm_test.go`

- [ ] **Step 1: Write failing cache and enhancement tests**

```go
func TestEnhancementCache_ReadWriteArticleEntry(t *testing.T) {
	store := sanctuary.NewStore(t.TempDir())
	entry := sanctuary.ArticleEnhancement{
		ArticleID:          "pick-01",
		GeneratedAt:        "2026-03-21T09:00:00Z",
		ConceptToMaster:    "Inference economics",
		EnhancementStatus:  "ready",
	}

	if err := store.WriteArticleEnhancement(entry); err != nil { t.Fatal(err) }
	got, err := store.ReadArticleEnhancement("pick-01")
	if err != nil { t.Fatal(err) }
	if got.ConceptToMaster != "Inference economics" {
		t.Fatalf("unexpected cache entry: %+v", got)
	}
}
```

- [ ] **Step 2: Run the focused enhancement tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestEnhancementCache_ReadWriteArticleEntry$' -v`
Expected: FAIL because there is no cache or provider abstraction yet.

- [ ] **Step 3: Implement the cache and non-blocking enhancer interface**

```go
type Enhancer interface {
	EnrichArticle(context.Context, model.ArticleDossier, profile.UserProfile) (ArticleEnhancement, error)
	EnrichGrowth(context.Context, profile.UserProfile, profile.LearningSnapshot) (GrowthEnhancement, error)
	EnrichReflection(context.Context, ReflectionEntry) (ReflectionEnhancement, error)
}
```

Implementation notes:
- Cache roots:
  - `state/sanctuary/cache/articles/{article_id}.json`
  - `state/sanctuary/cache/growth/latest.json`
  - `state/sanctuary/cache/reflections/{reflection_id}.json`
- Provider wiring can start with a nil/no-op implementation; the key contract is persistence, status transitions, and non-blocking usage.
- Reuse `internal/analyze` prompt/client code only where it genuinely fits; do not contort builder analysis APIs into synchronous page-read dependencies.
- Cache entries must record `generated_at`, `model_name`, `prompt_version`, and `enhancement_status`.

- [ ] **Step 4: Re-run the focused enhancement tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestEnhancementCache_ReadWriteArticleEntry$' -v`
Expected: PASS with file-backed enhancement cache helpers in place.

- [ ] **Step 5: Commit**

```bash
git add internal/sanctuary/llm.go internal/sanctuary/cache.go internal/sanctuary/store.go tests/sanctuary/llm_test.go
git commit -m "feat: add sanctuary enhancement cache"
```

---

### Task 3: Wire cached enhancement blocks into article, growth, and reflection reads

**Files:**
- Modify: `internal/sanctuary/content.go`
- Modify: `internal/sanctuary/growth.go`
- Modify: `internal/sanctuary/reflection.go`
- Modify: `internal/sanctuary/http.go`
- Modify: `tests/sanctuary/content_test.go`
- Modify: `tests/sanctuary/growth_test.go`
- Modify: `tests/sanctuary/reflection_test.go`

- [ ] **Step 1: Write failing enhancement-status tests**

```go
func TestSanctuaryAPI_ArticleReadUsesCacheButDoesNotRequireLiveLLM(t *testing.T) {
	outputDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	stateDir := t.TempDir()
	store := sanctuary.NewStore(stateDir)
	_ = store.WriteArticleEnhancement(sanctuary.ArticleEnhancement{
		ArticleID:         "pick-01",
		ConceptToMaster:   "Inference economics",
		EnhancementStatus: "ready",
	})

	handler := sanctuary.NewHandler(sanctuary.NewService(outputDir, stateDir))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles/pick-01", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d", rec.Code) }
	if !strings.Contains(rec.Body.String(), `"enhancement_status":"ready"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run the focused enhancement-status tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_ArticleReadUsesCacheButDoesNotRequireLiveLLM$' -v`
Expected: FAIL because read handlers do not attach cached enhancement state yet.

- [ ] **Step 3: Merge cached enhancements into read models while keeping deterministic fallback**

```go
// Article read path
if enhancement, err := store.ReadArticleEnhancement(articleID); err == nil {
	resp.ConceptToMaster = firstNonEmpty(resp.ConceptToMaster, enhancement.ConceptToMaster)
	resp.EnhancementStatus = enhancement.EnhancementStatus
} else {
	resp.EnhancementStatus = deriveEnhancementStatusFromDossier(resp)
}
```

Implementation notes:
- Never call the provider from a GET handler.
- If cache is absent:
  - article responses derive `pending` / `degraded` / `unavailable` from dossier completeness
  - growth responses derive status from whether cached explanation blocks exist
  - reflection responses derive status from whether summary/tag enhancement exists
- If a write path triggers enhancement, return success immediately after persistence and record `pending` in the stored entry.

- [ ] **Step 4: Re-run the focused enhancement-status tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_ArticleReadUsesCacheButDoesNotRequireLiveLLM$' -v`
Expected: PASS with cached enrichment wired in and read paths still deterministic.

- [ ] **Step 5: Commit**

```bash
git add internal/sanctuary/content.go internal/sanctuary/growth.go internal/sanctuary/reflection.go internal/sanctuary/http.go tests/sanctuary/content_test.go tests/sanctuary/growth_test.go tests/sanctuary/reflection_test.go
git commit -m "feat: wire sanctuary cached enhancement responses"
```

---

### Task 4: Document the follow-up slice and run the verification sweep

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Write the documentation checklist**

```text
- Reflection endpoint contract
- Enhancement cache roots and status model
- Explicit note that GET handlers do not block on live LLM calls
```

- [ ] **Step 2: Run the focused sanctuary verification suite**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -v`
Expected: PASS before the final docs update.

- [ ] **Step 3: Update docs**

```bash
# README additions
- reflection endpoints
- cache roots under state/sanctuary/cache/
- enhancement_status meanings: ready, pending, degraded, unavailable
```

- [ ] **Step 4: Run the full repository verification sweep**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./... -v`
Expected: PASS with reflection and enhancement cache support integrated.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: describe sanctuary reflection and enhancement runtime"
```
