# Digital Sanctuary Core Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first production-ready `sanctuary-api` slice for `apps/`, covering runtime skeleton, placeholder contracts, content reads, personal growth reads/writes, and RSS Source management on top of deterministic edition packages and file-backed sanctuary state.

**Architecture:** Keep `daily-builder` as the content engine and extend its package output with article dossiers under `output/_packages/YYYY/MM/DD/data/articles/`. Add a new `cmd/sanctuary-api` server backed by an `internal/sanctuary` package that reads package artifacts plus `state/sanctuary/...`, falls back to `state/feedback/...` where needed, and merges RSS runtime overrides back into the builder pipeline so future editions reflect user source management.

**Tech Stack:** Go 1.22+, stdlib `net/http` / `httptest` / `encoding/json`, existing `internal/output`, `internal/profile`, `internal/feedback`, file-backed JSON/JSONL state, `go test`.

**Execution Notes:** Follow @superpowers:test-driven-development for every task and @superpowers:verification-before-completion before claiming any task is done. Use Go 1.22 `http.ServeMux` path patterns for new parameterized routes, keep collection responses on the `items[]` / `next_cursor` / `total_estimate` contract, and run all `go test` commands with `GOCACHE="$(pwd)/.cache/go-build"` so verification stays workspace-local.

---

## File Structure Plan

### New files

- Create: `cmd/sanctuary-api/main.go`
  - standalone HTTP server for the new unified runtime API
- Create: `internal/output/reader.go`
  - package readers for `daily.json`, `learning.json`, `meta/edition.json`, and per-article dossier files
- Create: `internal/sanctuary/types.go`
  - shared response DTOs, target models, RSS override models, and placeholder payloads
- Create: `internal/sanctuary/store.go`
  - file-backed sanctuary state reads/writes plus feedback compatibility fallback
- Create: `internal/sanctuary/service.go`
  - top-level dependency container for package roots, state roots, and time helpers
- Create: `internal/sanctuary/http.go`
  - main router plus JSON helpers
- Create: `internal/sanctuary/middleware.go`
  - request-id, logging, panic recovery, and consistent error wrapping
- Create: `internal/sanctuary/content.go`
  - `briefing`, `articles`, `domains` read models and handlers
- Create: `internal/sanctuary/growth.go`
  - `knowledge-gaps`, `learning-plan`, `targets`, `tasks`, `profile`, and `profile/archives`
- Create: `internal/sanctuary/rss.go`
  - RSS Source list/add/update/delete and builder-facing effective-source merge helpers
- Create: `internal/sanctuary/placeholders.go`
  - preview-only `community` and `upgrade` responses
- Create: `tests/output/reader_test.go`
  - package reader regression tests
- Create: `tests/sanctuary/store_test.go`
  - sanctuary state IO, fallback, and effective RSS merge tests
- Create: `tests/sanctuary/http_test.go`
  - health, request-id, placeholder, and error-contract tests
- Create: `tests/sanctuary/content_test.go`
  - deterministic content API tests from package fixtures
- Create: `tests/sanctuary/growth_test.go`
  - growth read/write and archive tests
- Create: `tests/sanctuary/rss_test.go`
  - RSS Source contract tests including seed/runtime semantics

### Existing files to modify

- Modify: `internal/model/types.go`
  - add a dedicated `ArticleDossier` model and any minimal metadata needed for package readers
- Modify: `internal/output/package.go`
  - emit dossier JSON files under `data/articles/`
- Modify: `internal/run/pipeline.go`
  - build article dossiers during the daily run and apply RSS runtime overrides before feed fetching
- Modify: `internal/run/sample.go`
  - generate sample dossier artifacts and sample RSS-ready package data
- Modify: `internal/profile/profile.go`
  - add only the extra persisted profile fields actually required by first-release growth APIs
- Modify: `internal/feedback/service.go`
  - keep legacy snapshots compatible while persisting any new first-slice growth fields
- Modify: `cmd/daily-builder/main.go`
  - keep CLI output honest if the pipeline starts reporting applied RSS runtime overrides or richer package output
- Modify: `tests/output/package_test.go`
  - verify dossier emission alongside current package artifacts
- Modify: `tests/run/pipeline_test.go`
  - verify package dossier output and RSS override merge behavior
- Modify: `tests/sample/sample_generation_test.go`
  - verify sample edition writes dossier fixtures usable by `sanctuary-api`
- Modify: `tests/profile/profile_test.go`
  - cover any minimal new persisted profile fields
- Modify: `tests/feedback/http_test.go`
  - keep legacy feedback endpoints backward compatible after profile snapshot changes
- Modify: `README.md`
  - document `sanctuary-api` startup, routes, and state layout
- Modify: `docs/index.md`
  - add the new app/runtime overview
- Modify: `docs/ops/local-scheduler.md`
  - document that builder runs now also respect `state/sanctuary/rss/overrides.json`

### Existing files to reference

- Reference: `docs/superpowers/specs/2026-03-21-digital-sanctuary-platform-design.md`
- Reference: `cmd/feedback-api/main.go`
- Reference: `internal/feedback/http.go`
- Reference: `internal/feedback/store.go`
- Reference: `internal/output/package.go`
- Reference: `tests/feedback/http_test.go`
- Reference: `tests/output/package_test.go`
- Reference: `tests/run/pipeline_test.go`

---

### Task 1: Extend the edition package with article dossiers and read helpers

**Files:**
- Create: `internal/output/reader.go`
- Modify: `internal/model/types.go`
- Modify: `internal/output/package.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/run/sample.go`
- Modify: `tests/output/package_test.go`
- Create: `tests/output/reader_test.go`
- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/sample/sample_generation_test.go`

- [ ] **Step 1: Write the failing dossier and package-reader tests**

```go
func TestWriteEditionPackage_WritesArticleDossierFiles(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "output")
	editionRoot := t.TempDir()
	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC),
		Featured: []model.DailyPick{{ID: "pick-01", Title: "Headline"}},
		Articles: []model.ArticleDossier{{
			ArticleID:      "pick-01",
			EditionDate:    "2026-03-21",
			Title:          "Headline",
			Domain:         "ai-infrastructure",
			SummaryBrief:   "Brief",
			SummaryDeep:    "Deep",
			SourceName:     "Example",
			SourceURL:      "https://example.com/pick-01",
			EnhancementStatus: "ready",
		}},
	}

	root, err := output.WriteEditionPackage(baseDir, editionRoot, daily)
	if err != nil { t.Fatal(err) }
	if _, err := os.Stat(filepath.Join(root, "data", "articles", "pick-01.json")); err != nil {
		t.Fatalf("expected dossier file: %v", err)
	}
}

func TestReadEditionPackage_LoadsDailyLearningMetaAndDossier(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "output")
	date := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	root := filepath.Join(baseDir, "_packages", "2026", "03", "21")
	writeJSON(t, filepath.Join(root, "data", "daily.json"), map[string]any{"Featured": []map[string]any{{"ID": "pick-01"}}})
	writeJSON(t, filepath.Join(root, "data", "learning.json"), map[string]any{"learning": []string{"Track semis"}})
	writeJSON(t, filepath.Join(root, "meta", "edition.json"), map[string]any{"generated_at": "2026-03-21T09:00:00Z"})
	writeJSON(t, filepath.Join(root, "data", "articles", "pick-01.json"), map[string]any{"article_id": "pick-01", "title": "Headline"})

	pkg, err := output.ReadEditionPackage(baseDir, date)
	if err != nil { t.Fatal(err) }
	if len(pkg.Daily.Featured) != 1 || len(pkg.Articles) != 1 {
		t.Fatalf("unexpected package contents: %+v", pkg)
	}
}
```

- [ ] **Step 2: Run the focused output/run/sample tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/output ./tests/run ./tests/sample -run 'Test(WriteEditionPackage_WritesArticleDossierFiles|ReadEditionPackage_LoadsDailyLearningMetaAndDossier|RunPipeline_WritesEditionPackageAlongsideThemeOutput|GenerateSampleEdition_WritesThemeOutputAndEditionPackage)$' -v`
Expected: FAIL because the package has no dossier model or reader helpers yet.

- [ ] **Step 3: Add the dossier model, emit dossier JSON, and implement package readers**

```go
type ArticleDossier struct {
	ArticleID          string   `json:"article_id"`
	EditionDate        string   `json:"edition_date"`
	Title              string   `json:"title"`
	Domain             string   `json:"domain"`
	SourceName         string   `json:"source_name"`
	SourceURL          string   `json:"source_url"`
	PublishedAt        string   `json:"published_at"`
	HeroImage          string   `json:"hero_image"`
	SummaryBrief       string   `json:"summary_brief"`
	SummaryDeep        string   `json:"summary_deep"`
	KeyPoints          []string `json:"key_points"`
	Viewpoint          string   `json:"viewpoint"`
	OpportunityRisk    string   `json:"opportunity_risk"`
	ContrarianTake     string   `json:"contrarian_take"`
	EvidenceSnippets   []string `json:"evidence_snippets"`
	ConceptToMaster    string   `json:"concept_to_master"`
	Quote              string   `json:"quote"`
	ReadingTimeMin     int      `json:"reading_time_min"`
	TopicTags          []string `json:"topic_tags"`
	StyleTags          []string `json:"style_tags"`
	CognitiveTags      []string `json:"cognitive_tags"`
	WhyForYou          string   `json:"why_for_you"`
	TasteGrowthHint    string   `json:"taste_growth_hint"`
	KnowledgeGapHint   string   `json:"knowledge_gap_hint"`
	EnhancementStatus  string   `json:"enhancement_status"`
}

type EditionPackage struct {
	Daily    model.DailyEdition
	Learning []string
	Meta     map[string]any
	Articles []model.ArticleDossier
}

func ReadEditionPackage(baseDir string, date time.Time) (EditionPackage, error)
func ReadArticleDossier(baseDir string, date time.Time, articleID string) (model.ArticleDossier, error)
```

Implementation notes:
- Keep `DailyPick` lean and put article-detail-only fields in `model.ArticleDossier`.
- In `internal/run/pipeline.go`, carry a dossier alongside each candidate so the final featured set can write `daily.Articles` deterministically.
- In `internal/run/sample.go`, seed sample dossier records for `sample-1` and `sample-2` with realistic `summary_deep`, `concept_to_master`, and `enhancement_status`.
- `internal/output/reader.go` should read dossier files by listing `data/articles/*.json`; do not rely on Hugo output for API reads.
- If a dossier is missing for a featured item, readers should return a typed `os.ErrNotExist`-style error so the API layer can map it to `404`.

- [ ] **Step 4: Re-run the focused output/run/sample tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/output ./tests/run ./tests/sample -v`
Expected: PASS with new dossier artifacts and package readers in place.

- [ ] **Step 5: Commit**

```bash
git add internal/model/types.go internal/output/package.go internal/output/reader.go internal/run/pipeline.go internal/run/sample.go tests/output/package_test.go tests/output/reader_test.go tests/run/pipeline_test.go tests/sample/sample_generation_test.go
git commit -m "feat: add sanctuary article dossier package artifacts"
```

---

### Task 2: Add sanctuary state storage, feedback fallback, and effective RSS merge helpers

**Files:**
- Create: `internal/sanctuary/types.go`
- Create: `internal/sanctuary/store.go`
- Modify: `internal/run/pipeline.go`
- Create: `tests/sanctuary/store_test.go`
- Modify: `tests/run/pipeline_test.go`

- [ ] **Step 1: Write failing store and RSS-merge tests**

```go
func TestStore_ReadProfileSnapshotFallsBackToFeedbackState(t *testing.T) {
	stateDir := t.TempDir()
	feedbackStore := feedback.NewStore(stateDir)
	if err := feedbackStore.WriteProfileSnapshot(profile.UserProfile{
		FocusTopics: []string{"AI 基础设施"},
	}); err != nil {
		t.Fatal(err)
	}

	store := sanctuary.NewStore(stateDir)
	got, err := store.ReadProfileSnapshot()
	if err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(got.FocusTopics, []string{"AI 基础设施"}) {
		t.Fatalf("unexpected fallback snapshot: %+v", got)
	}
}

func TestEffectiveRSSSources_MergesSeedAndRuntimeSources(t *testing.T) {
	base := []config.RSSSource{{
		SourceID: "seed-1",
		Name:     "Seed Source",
		RSSURL:   "https://example.com/feed.xml",
		Domain:   "ai",
	}}
	overrides := sanctuary.RSSOverrideState{
		Items: []sanctuary.RSSSourceRecord{
			{SourceID: "seed-1", SourceKind: "seed", Enabled: false},
			{SourceID: "runtime-1", SourceKind: "runtime", Name: "Runtime Source", FeedURL: "https://runtime.example/feed.xml", Enabled: true},
		},
	}

	got := sanctuary.EffectiveRSSSources(base, overrides)
	if len(got) != 1 || got[0].SourceID != "runtime-1" {
		t.Fatalf("unexpected effective sources: %+v", got)
	}
}
```

- [ ] **Step 2: Run the focused sanctuary/run tests to verify red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary ./tests/run -run 'Test(Store_|EffectiveRSSSources_)' -v`
Expected: FAIL because `internal/sanctuary` and builder-side RSS merge support do not exist yet.

- [ ] **Step 3: Implement the sanctuary store and effective-source merge logic**

```go
type Store struct {
	stateDir      string
	feedbackStore *feedback.Store
}

func NewStore(stateDir string) *Store
func (s *Store) ReadProfileSnapshot() (profile.UserProfile, error)
func (s *Store) ReadLearningSnapshot() (profile.LearningSnapshot, error)
func (s *Store) ReadTargets() (TargetsState, error)
func (s *Store) WriteTargets(TargetsState) error
func (s *Store) ReadRSSOverrides() (RSSOverrideState, error)
func (s *Store) WriteRSSOverrides(RSSOverrideState) error
func (s *Store) ReadRSSSourceStats() (RSSSourceStatsState, error)
func (s *Store) WriteRSSSourceStats(RSSSourceStatsState) error

func EffectiveRSSSources(base []config.RSSSource, overrides RSSOverrideState) []RSSSourceRecord
```

Implementation notes:
- New sanctuary state roots:
  - `state/sanctuary/profile/profile_snapshot.json`
- `state/sanctuary/growth/learning_snapshot.json`
- `state/sanctuary/growth/targets.json`
- `state/sanctuary/rss/overrides.json`
- `state/sanctuary/rss/source_stats.json`
- `ReadProfileSnapshot` and `ReadLearningSnapshot` should try sanctuary first, then fall back to `feedback.NewStore(stateDir)` reads.
- `RSSSourceRecord` should distinguish `SourceKind: "seed" | "runtime"` so `DELETE` semantics stay deterministic later.
- `RSSSourceStatsState` should track at least `source_id`, `last_seen_at`, `last_featured_at`, and `featured_count`.
- In `internal/run/pipeline.go`, apply `EffectiveRSSSources` immediately after `loadConfig` and derive fetch URLs from the merged set, not directly from `cfg.RSS.Sources`.
- Keep source ordering stable: enabled seed sources first in config order, then runtime sources in persisted order.

- [ ] **Step 4: Re-run the focused sanctuary/run tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary ./tests/run -v`
Expected: PASS with state fallback and builder RSS merge behavior covered.

- [ ] **Step 5: Commit**

```bash
git add internal/sanctuary/types.go internal/sanctuary/store.go internal/run/pipeline.go tests/sanctuary/store_test.go tests/run/pipeline_test.go
git commit -m "feat: add sanctuary state storage and rss override merge"
```

---

### Task 3: Scaffold `sanctuary-api` with middleware, health, and placeholder contracts

**Files:**
- Create: `cmd/sanctuary-api/main.go`
- Create: `internal/sanctuary/service.go`
- Create: `internal/sanctuary/http.go`
- Create: `internal/sanctuary/middleware.go`
- Create: `internal/sanctuary/placeholders.go`
- Create: `tests/sanctuary/http_test.go`

- [ ] **Step 1: Write failing HTTP skeleton tests**

```go
func TestSanctuaryAPI_HealthAndPreviewContracts(t *testing.T) {
	handler := sanctuary.NewHandler(sanctuary.NewService("output", "state"))

	for _, path := range []string{"/healthz", "/api/v1/community/preview", "/api/v1/upgrade/offer"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, rec.Code)
		}
		if got := rec.Header().Get("X-Request-ID"); got == "" {
			t.Fatalf("%s missing request id", path)
		}
	}
}

func TestSanctuaryAPI_NotFoundUsesStructuredError(t *testing.T) {
	handler := sanctuary.NewHandler(sanctuary.NewService("output", "state"))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("unexpected error body: %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run the focused HTTP tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_(HealthAndPreviewContracts|NotFoundUsesStructuredError)$' -v`
Expected: FAIL because the new command/router/middleware do not exist yet.

- [ ] **Step 3: Implement the runtime skeleton, request-id middleware, and placeholder handlers**

```go
func NewService(outputDir, stateDir string) *Service

func NewHandler(svc *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /api/v1/community/preview", getCommunityPreview(svc))
	mux.HandleFunc("GET /api/v1/upgrade/offer", getUpgradeOffer(svc))
	// content / growth / rss handlers attach in later tasks
	return withMiddleware(mux)
}
```

Implementation notes:
- `cmd/sanctuary-api/main.go` should mirror `cmd/feedback-api/main.go`: read `AGENTIC_NEWS_OUTPUT_DIR`, `AGENTIC_NEWS_STATE_DIR`, and `AGENTIC_NEWS_SANCTUARY_ADDR`, then start an `http.Server` with sane timeouts.
- `internal/sanctuary/middleware.go` should:
  - assign `req_<timestamp>` request IDs when the client did not send one
  - echo the value as `X-Request-ID`
  - recover panics into `{ "error": { ... } }` JSON
- Placeholder payloads should be honest, small, and explicitly preview-only:
  - `community` includes `status`, `headline`, `body`, `cta`
  - `upgrade` includes `status`, `headline`, `body`, `offer_items[]`, `price_display`
- Add a JSON helper that always returns RFC3339 UTC timestamps and the shared error envelope.

- [ ] **Step 4: Re-run the focused HTTP tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_(HealthAndPreviewContracts|NotFoundUsesStructuredError)$' -v`
Expected: PASS with the skeleton runtime in place.

- [ ] **Step 5: Commit**

```bash
git add cmd/sanctuary-api/main.go internal/sanctuary/service.go internal/sanctuary/http.go internal/sanctuary/middleware.go internal/sanctuary/placeholders.go tests/sanctuary/http_test.go
git commit -m "feat: scaffold sanctuary api runtime and preview contracts"
```

---

### Task 4: Implement deterministic content APIs from package artifacts

**Files:**
- Modify: `internal/sanctuary/types.go`
- Modify: `internal/sanctuary/service.go`
- Modify: `internal/sanctuary/http.go`
- Create: `internal/sanctuary/content.go`
- Create: `tests/sanctuary/content_test.go`

- [ ] **Step 1: Write failing content API tests**

```go
func TestSanctuaryAPI_GetBriefingReturnsEditionCards(t *testing.T) {
	baseDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	handler := sanctuary.NewHandler(sanctuary.NewService(baseDir, t.TempDir()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/briefing?date=2026-03-21", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d", rec.Code) }
	if !strings.Contains(rec.Body.String(), `"edition_date":"2026-03-21"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSanctuaryAPI_GetArticleReturnsDossier(t *testing.T) {
	baseDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	handler := sanctuary.NewHandler(sanctuary.NewService(baseDir, t.TempDir()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles/pick-01", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d", rec.Code) }
	if !strings.Contains(rec.Body.String(), `"article_id":"pick-01"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSanctuaryAPI_GetDomainsAggregatesFeaturedItems(t *testing.T) {
	baseDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	handler := sanctuary.NewHandler(sanctuary.NewService(baseDir, t.TempDir()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/domains", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d", rec.Code) }
	if !strings.Contains(rec.Body.String(), `"items"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSanctuaryAPI_GetDomainDetailFiltersMatchingItems(t *testing.T) {
	baseDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	handler := sanctuary.NewHandler(sanctuary.NewService(baseDir, t.TempDir()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/domains/ai-infrastructure", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d", rec.Code) }
	if !strings.Contains(rec.Body.String(), `"domain_slug":"ai-infrastructure"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run the focused content tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_Get(BriefingReturnsEditionCards|ArticleReturnsDossier|DomainsAggregatesFeaturedItems|DomainDetailFiltersMatchingItems)$' -v`
Expected: FAIL because the content handlers and package reads are not wired yet.

- [ ] **Step 3: Implement `briefing`, `articles`, and `domains` handlers**

```go
type CollectionResponse[T any] struct {
	Items         []T    `json:"items"`
	NextCursor    string `json:"next_cursor"`
	TotalEstimate int    `json:"total_estimate"`
}

func getBriefing(svc *Service) http.HandlerFunc
func getArticle(svc *Service) http.HandlerFunc
func getDomains(svc *Service) http.HandlerFunc
func getDomainDetail(svc *Service) http.HandlerFunc
```

Implementation notes:
- `GET /api/v1/briefing?date=YYYY-MM-DD` should read `daily.json`, `learning.json`, and `meta/edition.json` from the package root and return:
  - `edition_date`
  - `generated_at`
  - `items[]`
  - `learning[]`
  - `enhancement_status`
- `GET /api/v1/articles/{article_id}` should read the dossier file and return deterministic detail even if `summary_deep` or `concept_to_master` is empty; in those cases set `enhancement_status` to `pending` or `degraded`.
- `GET /api/v1/domains` should aggregate by dossier `domain`, falling back to `DailyPick.Category` when the dossier domain is blank.
- `GET /api/v1/domains/{domain_slug}` should return the matching domain summary plus filtered article cards from the same edition package.
- Parse `date` strictly with `time.Parse("2006-01-02", ...)`; invalid input must return `invalid_date`.

- [ ] **Step 4: Re-run the focused content tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary -run 'TestSanctuaryAPI_Get(BriefingReturnsEditionCards|ArticleReturnsDossier|DomainsAggregatesFeaturedItems|DomainDetailFiltersMatchingItems)$' -v`
Expected: PASS with deterministic package-backed content APIs.

- [ ] **Step 5: Commit**

```bash
git add internal/sanctuary/types.go internal/sanctuary/service.go internal/sanctuary/http.go internal/sanctuary/content.go tests/sanctuary/content_test.go
git commit -m "feat: add sanctuary content api from edition packages"
```

---

### Task 5: Implement growth APIs, targets workflow, and profile archives

**Files:**
- Modify: `internal/profile/profile.go`
- Modify: `internal/feedback/service.go`
- Modify: `internal/sanctuary/types.go`
- Modify: `internal/sanctuary/store.go`
- Modify: `internal/sanctuary/http.go`
- Create: `internal/sanctuary/growth.go`
- Create: `tests/sanctuary/growth_test.go`
- Modify: `tests/profile/profile_test.go`
- Modify: `tests/feedback/http_test.go`

- [ ] **Step 1: Write failing growth tests**

```go
func TestSanctuaryAPI_GrowthEndpointsReadProfileLearningAndTargets(t *testing.T) {
	outputDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	stateDir := seedGrowthStateFixture(t)
	handler := sanctuary.NewHandler(sanctuary.NewService(outputDir, stateDir))

	for _, path := range []string{
		"/api/v1/profile",
		"/api/v1/growth/knowledge-gaps",
		"/api/v1/growth/learning-plan",
		"/api/v1/profile/archives",
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, rec.Code)
		}
	}
}

func TestSanctuaryAPI_TargetLifecycleAndTaskCompletion(t *testing.T) {
	outputDir := seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
	stateDir := t.TempDir()
	handler := sanctuary.NewHandler(sanctuary.NewService(outputDir, stateDir))

	postJSON(t, handler, http.MethodPost, "/api/v1/growth/targets", `{"title":"补课 GPU 供应链","domain":"ai-infrastructure","priority":"high"}`, http.StatusCreated)
	postJSON(t, handler, http.MethodPost, "/api/v1/growth/tasks/target-1/complete", `{}`, http.StatusOK)
	patchJSON(t, handler, "/api/v1/growth/targets/target-1", `{"status":"paused"}`, http.StatusOK)
}
```

- [ ] **Step 2: Run the focused growth/profile tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary ./tests/profile ./tests/feedback -run 'Test(SanctuaryAPI_GrowthEndpointsReadProfileLearningAndTargets|SanctuaryAPI_TargetLifecycleAndTaskCompletion)' -v`
Expected: FAIL because the growth handlers, target persistence, and archive listing do not exist yet.

- [ ] **Step 3: Implement growth responses, minimal snapshot expansion, and target persistence**

```go
type GrowthTarget struct {
	TargetID   string `json:"target_id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Source     string `json:"source"`
	Domain     string `json:"domain"`
	Priority   string `json:"priority"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type TargetsState struct {
	WeekOf string         `json:"week_of"`
	Items  []GrowthTarget `json:"items"`
}
```

Implementation notes:
- Keep `profile.UserProfile` changes minimal and first-slice-driven; only add persisted fields if a growth response cannot be derived cleanly from existing affinity maps and summaries.
- `GET /api/v1/profile` should expose the persisted snapshot plus derived `current_focus` and `knowledge_points` if those are useful to the page contract.
- `GET /api/v1/growth/knowledge-gaps` should derive ranked gap items from:
  - `LearningSnapshot.KnowledgeGapHint`
  - negative or weak affinity areas in the profile snapshot
  - current target domains
- `GET /api/v1/growth/learning-plan` should return deterministic tasks built from persisted targets first, then add non-completable system recommendations from the learning snapshot.
- `POST /api/v1/growth/tasks/{task_id}/complete` should only mutate target-backed tasks; return `unsupported_action` for non-persisted recommendation rows.
- `GET /api/v1/profile/archives` should scan package roots under `output/_packages/*/*/*` and build a paginated archive list from `meta/edition.json` plus `daily.json`, newest first.
- Preserve legacy `GET /api/v1/profile` and `GET /api/v1/profile/learning` behavior under `cmd/feedback-api`; do not break existing tests or response fields there.

- [ ] **Step 4: Re-run the focused growth/profile tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary ./tests/profile ./tests/feedback -v`
Expected: PASS with growth reads/writes, archive listing, and feedback compatibility intact.

- [ ] **Step 5: Commit**

```bash
git add internal/profile/profile.go internal/feedback/service.go internal/sanctuary/types.go internal/sanctuary/store.go internal/sanctuary/http.go internal/sanctuary/growth.go tests/sanctuary/growth_test.go tests/profile/profile_test.go tests/feedback/http_test.go
git commit -m "feat: add sanctuary growth api and target workflow"
```

---

### Task 6: Implement RSS Source management APIs and hook them back into the builder

**Files:**
- Modify: `internal/sanctuary/types.go`
- Modify: `internal/sanctuary/store.go`
- Modify: `internal/sanctuary/http.go`
- Create: `internal/sanctuary/rss.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/run/sample.go`
- Create: `tests/sanctuary/rss_test.go`
- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/sample/sample_generation_test.go`

- [ ] **Step 1: Write failing RSS Source API tests**

```go
func TestSanctuaryAPI_RSSSourceCRUDHonorsSeedVsRuntimeRules(t *testing.T) {
	stateDir := t.TempDir()
	handler := sanctuary.NewHandler(sanctuary.NewService(seedEditionPackageFixture(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)), stateDir))

	postJSON(t, handler, http.MethodPost, "/api/v1/rss/sources", `{
		"name":"Runtime Feed",
		"feed_url":"https://runtime.example/feed.xml",
		"domain":"systems"
	}`, http.StatusCreated)

	patchJSON(t, handler, "/api/v1/rss/sources/seed-1", `{"enabled":false}`, http.StatusOK)
	deleteReq(t, handler, "/api/v1/rss/sources/runtime-1", http.StatusNoContent)
	deleteReq(t, handler, "/api/v1/rss/sources/seed-1", http.StatusConflict)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rss/sources", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"items"`) {
		t.Fatalf("unexpected list body: %s", rec.Body.String())
	}
}

func TestRunPipeline_AppliesSanctuaryRSSOverrides(t *testing.T) {
	stateDir := t.TempDir()
	seedRSSOverrides(t, stateDir, sanctuary.RSSOverrideState{
		Items: []sanctuary.RSSSourceRecord{
			{SourceID: "seed-1", SourceKind: "seed", Enabled: false},
			{SourceID: "runtime-1", SourceKind: "runtime", Name: "Runtime Feed", FeedURL: "https://runtime.example/feed.xml", Enabled: true},
		},
	})

	var fetched []string
	_, err := run.RunDryPipeline(context.Background(), run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		StateDir:  stateDir,
		Date:      time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}, run.DryRunHooks{
		LoadConfig: func(string) (config.Config, error) { return singleSourceConfig(), nil },
		FetchFeeds: func(_ context.Context, urls []string) ([]model.RawItem, error) {
			fetched = append([]string(nil), urls...)
			return testRawItems(10), nil
		},
	})
	if err != nil { t.Fatal(err) }
	if len(fetched) != 1 || fetched[0] != "https://runtime.example/feed.xml" {
		t.Fatalf("unexpected fetched urls: %#v", fetched)
	}
}
```

- [ ] **Step 2: Run the focused RSS tests to confirm red**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary ./tests/run -run 'Test(SanctuaryAPI_RSSSourceCRUDHonorsSeedVsRuntimeRules|RunPipeline_AppliesSanctuaryRSSOverrides)$' -v`
Expected: FAIL because the RSS handlers and persisted override semantics are not implemented yet.

- [ ] **Step 3: Implement RSS list/add/update/delete with persisted override semantics**

```go
type RSSSourceRecord struct {
	SourceID         string `json:"source_id"`
	SourceKind       string `json:"source_kind"`
	Name             string `json:"name"`
	FeedURL          string `json:"feed_url"`
	Domain           string `json:"domain"`
	Enabled          bool   `json:"enabled"`
	DensityMode      string `json:"density_mode"`
	QualityOverride  string `json:"quality_override"`
	DomainOverride   string `json:"domain_override"`
	UpdatedAt        string `json:"updated_at"`
}
```

Implementation notes:
- `GET /api/v1/rss/sources` should join config-backed seed sources with override state and return a stable collection payload.
- `GET /api/v1/rss/sources` should also join `source_stats.json` so each row can expose deterministic metrics such as `last_seen_at`, `last_featured_at`, and `featured_count`.
- `POST /api/v1/rss/sources` should create a runtime source with generated `source_id`, `source_kind: "runtime"`, and `enabled: true`.
- `PATCH /api/v1/rss/sources/{source_id}` should support partial updates to `enabled`, `density_mode`, `quality_override`, and `domain_override`.
- `DELETE /api/v1/rss/sources/{source_id}` should:
  - remove runtime sources physically from `overrides.json`
  - reject seed-source deletes with `conflict` or `unsupported_action`
- After each successful builder run and sample generation, refresh `state/sanctuary/rss/source_stats.json` from the produced edition so source metrics remain deterministic and file-backed.
- Update builder tests so the next `daily-builder run` respects the merged effective source list without mutating `config/rss_sources.yaml`.

- [ ] **Step 4: Re-run the focused RSS tests**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/sanctuary ./tests/run -v`
Expected: PASS with persisted RSS source management and builder integration covered.

- [ ] **Step 5: Commit**

```bash
git add internal/sanctuary/types.go internal/sanctuary/store.go internal/sanctuary/http.go internal/sanctuary/rss.go internal/run/pipeline.go internal/run/sample.go tests/sanctuary/rss_test.go tests/run/pipeline_test.go tests/sample/sample_generation_test.go
git commit -m "feat: add sanctuary rss source management"
```

---

### Task 7: Document the runtime and run the full backend verification sweep

**Files:**
- Modify: `cmd/daily-builder/main.go`
- Modify: `README.md`
- Modify: `docs/index.md`
- Modify: `docs/ops/local-scheduler.md`

- [ ] **Step 1: Write the failing documentation/CLI expectations**

```text
- README should show how to run `cmd/sanctuary-api`
- docs/index.md should mention the Digital Sanctuary runtime split
- docs/ops/local-scheduler.md should note RSS overrides influence future builder runs
- `daily-builder` CLI output should stay accurate if package generation now includes dossier artifacts
```

- [ ] **Step 2: Run the final verification commands before editing docs**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./tests/output ./tests/sanctuary ./tests/run ./tests/sample ./tests/profile ./tests/feedback -v`
Expected: PASS before documentation edits, giving a clean backend verification baseline.

- [ ] **Step 3: Update docs and any small CLI/help text needed for the new runtime**

```bash
# README additions
- sanctuary-api environment variables
- route summary for /api/v1/briefing, /api/v1/articles/{id}, /api/v1/growth/*, /api/v1/rss/sources
- state roots under state/sanctuary/

# docs/ops/local-scheduler.md additions
- daily-builder reads state/sanctuary/rss/overrides.json on each run
- seed config remains authoritative for seed sources; runtime adds overlays only
```

Implementation notes:
- Keep docs concise and operational.
- Do not document Reflection or live LLM endpoints here; that belongs to the later plan.
- If `cmd/daily-builder/main.go` already prints enough information, avoid churn; only change text if it would otherwise be misleading.

- [ ] **Step 4: Run the full backend verification sweep**

Run: `GOCACHE="$(pwd)/.cache/go-build" go test ./... -v`
Expected: PASS with `sanctuary-api`, package readers, growth APIs, and RSS management all green.

- [ ] **Step 5: Commit**

```bash
git add cmd/daily-builder/main.go README.md docs/index.md docs/ops/local-scheduler.md
git commit -m "docs: document sanctuary api core runtime"
```
