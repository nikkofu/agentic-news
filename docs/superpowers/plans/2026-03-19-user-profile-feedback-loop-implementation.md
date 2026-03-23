# User Profile and Feedback Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a publishable single-user profile and feedback learning loop that collects explicit/implicit feedback, persists append-only events plus derived snapshots, refreshes same-day profile/learning panels, and applies the saved profile to the next daily generation run.

**Architecture:** Keep the existing static-edition pipeline intact and add a small co-located Go feedback service. Store raw feedback as append-only JSONL under `state/feedback/`, materialize a latest profile snapshot and learning snapshot from those events, expose lightweight API endpoints for the frontend, and have `daily-builder run` load the latest snapshot to drive personalization scoring and explanation.

**Tech Stack:** Go 1.22+, stdlib `net/http` / `httptest` / `encoding/json`, existing `internal/run`, `internal/render`, `internal/profile`, `internal/rank`, static HTML templates, browser-side `web/static/app.js`, `go test`.

---

## File Structure Plan

### New files

- Create: `cmd/feedback-api/main.go`
  - standalone lightweight HTTP server for feedback ingestion and profile reads
- Create: `internal/feedback/types.go`
  - request/response DTOs and persisted event/snapshot shapes
- Create: `internal/feedback/store.go`
  - append-only JSONL event writes and snapshot read/write helpers under `state/feedback/`
- Create: `internal/feedback/service.go`
  - orchestrates append → recompute profile → recompute learning snapshot
- Create: `internal/feedback/http.go`
  - `POST /api/v1/feedback/events`, `GET /api/v1/profile`, `GET /api/v1/profile/learning`
- Create: `tests/profile/profile_test.go`
  - pure profile aggregation / learning-hint tests
- Create: `tests/feedback/store_test.go`
  - JSONL storage and snapshot persistence tests
- Create: `tests/feedback/http_test.go`
  - handler-level API tests with `httptest`

### Existing files to modify

- Modify: `internal/model/types.go`
  - add the article/card metadata needed for feedback tags and same-day explanation surfaces
- Modify: `internal/profile/profile.go`
  - expand `UserProfile` and add deterministic event-application helpers
- Modify: `internal/rank/scoring.go`
  - compute profile-aware personal relevance inputs
- Modify: `internal/analyze/pipeline.go`
  - generate richer `why_for_you`, taste, and knowledge hints from the real profile
- Modify: `internal/run/pipeline.go`
  - load profile snapshot before ranking/analysis and fall back safely when missing
- Modify: `internal/verify/checks.go`
  - validate personalized output fallback behavior without weakening publishability checks
- Modify: `internal/render/index.go`
  - include lightweight tracking metadata or profile entry hooks on cards/index
- Modify: `internal/render/article.go`
  - include article tags, feedback controls, and same-day profile/learning containers
- Modify: `web/templates/index.tmpl`
  - add click-tracking hooks and compact profile entry point
- Modify: `web/templates/article.tmpl`
  - add feedback controls, reason tags, “why recommended for you”, and live profile panels
- Modify: `web/static/app.js`
  - send events, record dwell/click signals, and refresh same-day profile/learning panels
- Modify: `cmd/daily-builder/main.go`
  - keep builder behavior stable while ensuring the run path points at `state/` profile data
- Modify: `README.md`
  - document the feedback API and the profile-learning loop
- Modify: `docs/index.md`
  - reflect that the product now includes same-day feedback learning
- Modify: `docs/ops/local-scheduler.md`
  - explain that daily generation uses `state/feedback/profile_snapshot.json` when present

### Existing files to reference

- Reference: `docs/superpowers/specs/2026-03-19-user-profile-feedback-loop-design.md`
- Reference: `internal/output/writer.go`
- Reference: `internal/run/options.go`
- Reference: `tests/run/pipeline_test.go`
- Reference: `tests/render/render_test.go`
- Reference: `tests/analyze/pipeline_test.go`

---

### Task 1: Expand the profile model and deterministic aggregation helpers

**Files:**
- Modify: `internal/profile/profile.go`
- Test: `tests/profile/profile_test.go`

- [ ] **Step 1: Write failing profile aggregation tests**

```go
func TestApplyEvent_UpdatesTopicStyleAndCognitiveAffinity(t *testing.T) {
    current := profile.UserProfile{}
    event := profile.EventInput{
        EventType:     "feedback_like",
        TopicTags:     []string{"ai infrastructure", "policy"},
        StyleTags:     []string{"deep-analysis", "data-driven"},
        CognitiveTags: []string{"risk", "framework"},
        SourceName:    "Example Source",
    }

    got := profile.ApplyEvent(current, event)

    if got.TopicAffinity["ai infrastructure"] <= 0 { t.Fatal("expected topic boost") }
    if got.StyleAffinity["deep-analysis"] <= 0 { t.Fatal("expected style boost") }
    if got.CognitiveAffinity["risk"] <= 0 { t.Fatal("expected cognitive boost") }
}

func TestApplyEvent_RecordsNegativeSignalsForDislike(t *testing.T) {
    current := profile.UserProfile{}
    event := profile.EventInput{
        EventType: "feedback_dislike",
        TopicTags: []string{"macro opinion"},
    }

    got := profile.ApplyEvent(current, event)

    if got.NegativeSignals["macro opinion"] >= 0 {
        t.Fatal("expected negative signal")
    }
}

func TestBuildLearningSnapshot_ReturnsTasteAndKnowledgeHints(t *testing.T) {
    current := profile.UserProfile{
        TopicAffinity:     map[string]float64{"semiconductors": 8},
        StyleAffinity:     map[string]float64{"data-driven": 6},
        CognitiveAffinity: map[string]float64{"risk": 7, "opportunity": 1},
    }

    learning := profile.BuildLearningSnapshot(current)

    if learning.TasteGrowthHint == "" || learning.KnowledgeGapHint == "" {
        t.Fatal("expected non-empty learning hints")
    }
    if learning.TodayPlan == "" || len(learning.LearningTracks) == 0 || learning.UpdatedAt.IsZero() {
        t.Fatal("expected full learning snapshot fields")
    }
    if !strings.Contains(strings.Join(learning.LearningTracks, " "), "balance") {
        t.Fatal("expected balancing guidance when one cognitive lens dominates")
    }
}

func TestApplyEvent_BuildsRecentFeedbackSummary(t *testing.T) {
    current := profile.UserProfile{}
    event := profile.EventInput{
        EventType:     "feedback_like",
        TopicTags:     []string{"ai infrastructure"},
        StyleTags:     []string{"data-driven"},
        CognitiveTags: []string{"risk"},
        ReasonTags:    []string{"想看更多数据"},
    }

    got := profile.ApplyEvent(current, event)

    if len(got.RecentFeedbackSummary) == 0 {
        t.Fatal("expected recent feedback summary entry")
    }
}
```

- [ ] **Step 2: Run the new profile tests to confirm red**

Run: `go test ./tests/profile -run 'Test(ApplyEvent_|BuildLearningSnapshot_)' -v`
Expected: FAIL because the current profile package only handles minimal topic affinity and has no event-driven or learning-snapshot helpers.

- [ ] **Step 3: Expand `UserProfile` and add pure aggregation helpers**

```go
type UserProfile struct {
    FocusTopics           []string
    PreferredStyles       []string
    CognitivePreferences  []string
    ExplicitFeedback      map[string]string
    BehaviorSignals       map[string]float64
    TopicAffinity         map[string]float64
    StyleAffinity         map[string]float64
    CognitiveAffinity     map[string]float64
    SourceAffinity        map[string]float64
    NegativeSignals       map[string]float64
    RecentFeedbackSummary []string
    LastUpdatedAt         time.Time
}

type EventInput struct {
    EventType     string
    TopicTags     []string
    StyleTags     []string
    CognitiveTags []string
    SourceName    string
    DwellSeconds  float64
    Bookmarked    bool
    ReasonTags    []string
}

func ApplyEvent(p UserProfile, event EventInput) UserProfile
func BuildLearningSnapshot(p UserProfile) LearningSnapshot
func TopKeys(scores map[string]float64, n int) []string

type LearningSnapshot struct {
    TasteGrowthHint  string
    KnowledgeGapHint string
    TodayPlan        string
    LearningTracks   []string
    UpdatedAt        time.Time
}
```

Implementation notes:
- keep weighting deterministic and explainable
- make explicit feedback stronger than dwell/click signals
- cap dwell-time contribution
- refresh `FocusTopics`, `PreferredStyles`, and `CognitivePreferences` from the top affinity maps
- derive `RecentFeedbackSummary` from the newest event inputs using short human-readable statements such as “你最近连续强化了 AI 基础设施 + 数据型内容”
- keep only a small recent window (for example last 3 summary lines) so the snapshot stays UI-friendly
- if one cognitive lens becomes much stronger than the others, have `BuildLearningSnapshot` add balancing guidance into `LearningTracks` / hints (for example, suggest adding risk-validation reading when opportunity clicks dominate)

- [ ] **Step 4: Run the full profile package**

Run: `go test ./tests/profile -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/profile/profile.go tests/profile/profile_test.go
git commit -m "feat: expand user profile aggregation for feedback learning"
```

---

### Task 2: Add append-only feedback storage and snapshot persistence

**Files:**
- Create: `internal/feedback/types.go`
- Create: `internal/feedback/store.go`
- Test: `tests/feedback/store_test.go`

- [ ] **Step 1: Write failing storage tests for JSONL append and snapshot IO**

```go
func TestStore_AppendEventWritesJSONLRecord(t *testing.T) {
    dir := t.TempDir()
    store := feedback.NewStore(dir)

    err := store.AppendEvent(feedback.Event{
        EventID:   "evt-1",
        EventType: "feedback_like",
        ArticleID: "a1",
    })
    if err != nil { t.Fatal(err) }

    events, err := store.ReadEvents("2026-03")
    if err != nil { t.Fatal(err) }
    if len(events) != 1 || events[0].EventID != "evt-1" {
        t.Fatalf("unexpected events: %+v", events)
    }
}

func TestStore_WriteAndReadProfileSnapshot(t *testing.T) {
    dir := t.TempDir()
    store := feedback.NewStore(dir)
    want := profile.UserProfile{FocusTopics: []string{"policy"}}

    if err := store.WriteProfileSnapshot(want); err != nil { t.Fatal(err) }
    got, err := store.ReadProfileSnapshot()
    if err != nil { t.Fatal(err) }
    if !reflect.DeepEqual(want.FocusTopics, got.FocusTopics) {
        t.Fatalf("focus topics mismatch: want %v got %v", want.FocusTopics, got.FocusTopics)
    }
}

func TestStore_WriteAndReadLearningSnapshot(t *testing.T) {
    dir := t.TempDir()
    store := feedback.NewStore(dir)
    want := profile.LearningSnapshot{
        TasteGrowthHint:  "Prefer data-backed reporting.",
        KnowledgeGapHint: "Review export control basics.",
        TodayPlan:        "Read one standard card and compare source claim vs viewpoint.",
        LearningTracks:   []string{"taste", "cognition", "knowledge"},
        UpdatedAt:        time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC),
    }

    if err := store.WriteLearningSnapshot(want); err != nil { t.Fatal(err) }
    got, err := store.ReadLearningSnapshot()
    if err != nil { t.Fatal(err) }
    if got.TodayPlan == "" || len(got.LearningTracks) != 3 || got.UpdatedAt.IsZero() {
        t.Fatalf("unexpected learning snapshot: %+v", got)
    }
}
```

- [ ] **Step 2: Run store tests to confirm red**

Run: `go test ./tests/feedback -run 'TestStore_' -v`
Expected: FAIL because no feedback storage package exists yet.

- [ ] **Step 3: Implement focused feedback persistence helpers**

```go
type Event struct {
    EventID        string
    EventType      string
    Timestamp      time.Time
    EditionDate    string
    ArticleID      string
    ArticleTitle   string
    ArticleURL     string
    SourceName     string
    TopicTags      []string
    StyleTags      []string
    CognitiveTags  []string
    Metadata       map[string]any
}

type Store struct { root string }

type LearningSnapshot = profile.LearningSnapshot

func NewStore(stateDir string) *Store
func (s *Store) AppendEvent(event Event) error
func (s *Store) ReadEvents(month string) ([]Event, error)
func (s *Store) WriteProfileSnapshot(p profile.UserProfile) error
func (s *Store) ReadProfileSnapshot() (profile.UserProfile, error)
func (s *Store) WriteLearningSnapshot(v profile.LearningSnapshot) error
func (s *Store) ReadLearningSnapshot() (profile.LearningSnapshot, error)
```

Implementation notes:
- store under `filepath.Join(stateDir, "feedback", ...)`
- use monthly event files named `YYYY-MM.jsonl`
- create parent directories lazily
- keep the snapshot JSON human-readable with `json.MarshalIndent`
- learning snapshot must round-trip `today_plan`, `learning_tracks`, and `updated_at`, not only taste/knowledge hint strings

- [ ] **Step 4: Run the full feedback store tests**

Run: `go test ./tests/feedback -run 'TestStore_' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/feedback/types.go internal/feedback/store.go tests/feedback/store_test.go
git commit -m "feat: add append-only feedback storage and snapshots"
```

---

### Task 3: Add feedback orchestration and HTTP endpoints

**Files:**
- Create: `internal/feedback/service.go`
- Create: `internal/feedback/http.go`
- Create: `cmd/feedback-api/main.go`
- Test: `tests/feedback/http_test.go`

- [ ] **Step 1: Write failing API tests for event ingestion and profile reads**

```go
func TestFeedbackAPI_PostEventAppendsAndReturnsUpdatedProfile(t *testing.T) {
    server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(t.TempDir()))))
    defer server.Close()

    payload := `{
      "event_type":"feedback_like",
      "timestamp":"2026-03-19T09:00:00Z",
      "edition_date":"2026-03-19",
      "article_id":"a1",
      "topic_tags":["ai infrastructure"],
      "style_tags":["deep-analysis"],
      "cognitive_tags":["risk"]
    }`

    resp, err := http.Post(server.URL+"/api/v1/feedback/events", "application/json", strings.NewReader(payload))
    if err != nil { t.Fatal(err) }
    if resp.StatusCode != http.StatusOK { t.Fatalf("unexpected status: %d", resp.StatusCode) }
}

func TestFeedbackAPI_GetProfileReturnsSnapshot(t *testing.T) {
    // seed snapshot, then GET /api/v1/profile and verify:
    // - focus_topics
    // - preferred_styles
    // - cognitive_preferences
    // - recent_feedback_summary
}

func TestFeedbackAPI_GetLearningReturnsTracksAndTodayPlan(t *testing.T) {
    // seed learning snapshot, then GET /api/v1/profile/learning and verify:
    // - taste_growth_hint
    // - knowledge_gap_hint
    // - today_plan
    // - learning_tracks
    // - updated_at
}

func TestFeedbackAPI_RejectsMalformedEventPayload(t *testing.T) {
    // missing event_type / article_id should return 400
}
```

- [ ] **Step 2: Run handler tests to confirm red**

Run: `go test ./tests/feedback -run 'TestFeedbackAPI_' -v`
Expected: FAIL because the feedback service and handlers do not exist yet.

- [ ] **Step 3: Implement service + handlers + binary**

```go
type Service struct {
    store *Store
}

func (s *Service) RecordEvent(event Event) (ProfileResponse, LearningResponse, error) {
    // append event
    // load current profile snapshot (or zero-value profile)
    // apply event
    // write updated profile snapshot
    // write updated learning snapshot
    // return the lightweight response payloads
}

type LearningResponse struct {
    TasteGrowthHint  string   `json:"taste_growth_hint"`
    KnowledgeGapHint string   `json:"knowledge_gap_hint"`
    TodayPlan        string   `json:"today_plan"`
    LearningTracks   []string `json:"learning_tracks"`
    UpdatedAt        string   `json:"updated_at"`
}

func NewHandler(svc *Service) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/v1/feedback/events", postEventHandler(svc))
    mux.HandleFunc("/api/v1/profile", getProfileHandler(svc))
    mux.HandleFunc("/api/v1/profile/learning", getLearningHandler(svc))
    return mux
}

func main() {
    stateDir := firstNonEmpty(os.Getenv("AGENTIC_NEWS_STATE_DIR"), "state")
    addr := firstNonEmpty(os.Getenv("AGENTIC_NEWS_FEEDBACK_ADDR"), ":8081")
    log.Fatal(http.ListenAndServe(addr, feedback.NewHandler(feedback.NewService(feedback.NewStore(stateDir)))))
}
```

Implementation notes:
- return `400` for malformed payloads
- return `200` with latest profile summary after successful writes
- ensure `/api/v1/profile/learning` returns `today_plan`, `learning_tracks`, and `updated_at` in addition to the hint fields
- keep handler code thin; all append/recompute logic belongs in the service

- [ ] **Step 4: Run all feedback package tests**

Run: `go test ./tests/feedback -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/feedback/service.go internal/feedback/http.go cmd/feedback-api/main.go tests/feedback/http_test.go
git commit -m "feat: add feedback api for profile learning loop"
```

---

### Task 4: Add personalization metadata to cards and render same-day feedback surfaces

**Files:**
- Modify: `internal/model/types.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/article.tmpl`
- Modify: `web/static/app.js`
- Test: `tests/render/render_test.go`

- [ ] **Step 1: Write failing render tests for feedback controls and personalized surfaces**

```go
func TestRenderDailyOutput_ArticlePageIncludesFeedbackControlsAndWhyForYou(t *testing.T) {
    daily := model.DailyEdition{
        Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
        Featured: []model.DailyPick{
            {
                ID:            "a1",
                Title:         "Chip policy moves",
                Summary:       "Short summary",
                SourceName:    "Example",
                SourceURL:     "https://example.com/a1",
                TopicTags:     []string{"semiconductors", "policy"},
                StyleTags:     []string{"deep-analysis"},
                CognitiveTags: []string{"risk"},
                Insight: model.Insight{
                    WhyForYou:        "Because you have recently focused on semiconductor policy risk.",
                    TasteGrowthHint:  "Prefer data-backed reporting.",
                    KnowledgeGapHint: "Review export control basics.",
                },
            },
        },
    }

    // render, then verify HTML contains:
    // - feedback buttons
    // - reason tags container
    // - why_for_you text
    // - profile panel container
    // - data-topic-tags / data-style-tags / data-cognitive-tags
}

func TestRenderDailyOutput_IndexPageIncludesTrackingHooks(t *testing.T) {
    // verify rendered index contains article link metadata the frontend can use for click events
}
```

- [ ] **Step 2: Run the targeted render tests to confirm red**

Run: `go test ./tests/render -run 'TestRenderDailyOutput_(ArticlePageIncludesFeedbackControlsAndWhyForYou|IndexPageIncludesTrackingHooks)' -v`
Expected: FAIL because rendered pages currently have no feedback controls, no profile containers, and no event metadata for the frontend.

- [ ] **Step 3: Extend card/template data and add the lightweight UI shell**

```go
type DailyPick struct {
    // existing fields...
    TopicTags     []string
    StyleTags     []string
    CognitiveTags []string
}

type articleData struct {
    // existing fields...
    WhyForYou        string
    TasteGrowthHint  string
    KnowledgeGapHint string
    TopicTagsJSON    string
    StyleTagsJSON    string
    CognitiveTagsJSON string
}
```

Implementation notes:
- article template should add:
  - like / dislike / bookmark controls
  - reason-tag buttons
  - profile summary panel
  - learning refresh panel
- keep the page readable even if JavaScript never runs
- index/article markup should expose enough `data-*` or embedded JSON for `app.js` to send events
- `app.js` should:
  - send `article_view` on load
  - send `dwell_report` on page hide / unload
  - submit explicit feedback button clicks
  - fetch `/api/v1/profile` and `/api/v1/profile/learning` after successful feedback

- [ ] **Step 4: Run the full render test package**

Run: `go test ./tests/render -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/model/types.go internal/render/index.go internal/render/article.go web/templates/index.tmpl web/templates/article.tmpl web/static/app.js tests/render/render_test.go
git commit -m "feat: render feedback controls and live profile panels"
```

---

### Task 5: Apply saved profile data to scoring, explanation, and the daily run

**Files:**
- Modify: `internal/rank/scoring.go`
- Modify: `internal/analyze/pipeline.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/verify/checks.go`
- Modify: `cmd/daily-builder/main.go`
- Test: `tests/rank/scoring_test.go`
- Test: `tests/analyze/pipeline_test.go`
- Test: `tests/run/pipeline_test.go`
- Test: `tests/verify/checks_test.go`

- [ ] **Step 1: Write failing tests for profile-aware scoring and run-time snapshot loading**

```go
func TestScorePersonalRelevance_UsesTopicStyleCognitiveAndNegativeSignals(t *testing.T) {
    p := profile.UserProfile{
        TopicAffinity:     map[string]float64{"policy": 6},
        StyleAffinity:     map[string]float64{"deep-analysis": 4},
        CognitiveAffinity: map[string]float64{"risk": 5},
        NegativeSignals:   map[string]float64{"macro opinion": -8},
    }

    got := rank.ScorePersonalRelevance(p,
        []string{"policy"},
        []string{"deep-analysis"},
        []string{"risk"},
        "Example Source",
    )

    if got <= 0 { t.Fatal("expected positive relevance score") }
}

func TestAnalyzeRunPipeline_BuildsWhyForYouFromExpandedProfile(t *testing.T) {
    article := model.Article{
        Title:        "Chip export controls",
        ContentText:  "Policy changes may reshape semiconductor capacity planning.",
        CanonicalURL: "https://example.com/chips",
    }
    p := profile.UserProfile{
        FocusTopics:          []string{"semiconductors"},
        PreferredStyles:      []string{"deep-analysis"},
        CognitivePreferences: []string{"risk"},
    }

    insight, err := analyze.RunPipeline(context.Background(), article, p)
    if err != nil { t.Fatal(err) }
    if !strings.Contains(insight.WhyForYou, "semiconductors") {
        t.Fatalf("expected topic mention, got %q", insight.WhyForYou)
    }
}

func TestRunPipeline_LoadsProfileSnapshotAndFallsBackWhenMissing(t *testing.T) {
    // one subtest seeds a snapshot and expects personalized why_for_you
    // one subtest leaves state empty and expects no failure plus default profile behavior
}

func TestDailyEdition_AllowsMissingProfileSnapshotByVerifyingFallbackCopy(t *testing.T) {
    // render a daily.json generated from the default profile path
    // verify should still pass as long as personalized fields fall back to non-empty safe copy
}
```

- [ ] **Step 2: Run the targeted tests to confirm red**

Run: `go test ./tests/rank -run 'TestScorePersonalRelevance_' -v`
Expected: FAIL because rank package does not yet compute profile-driven relevance.

Run: `go test ./tests/analyze -run 'TestAnalyzeRunPipeline_BuildsWhyForYouFromExpandedProfile' -v`
Expected: FAIL because `buildWhyForYou` currently only uses `FocusTopics`.

Run: `go test ./tests/run -run 'TestRunPipeline_LoadsProfileSnapshotAndFallsBackWhenMissing' -v`
Expected: FAIL because the run pipeline does not yet load `state/feedback/profile_snapshot.json`.

Run: `go test ./tests/verify -run 'TestDailyEdition_AllowsMissingProfileSnapshotByVerifyingFallbackCopy' -v`
Expected: FAIL because the verifier does not yet assert personalized fallback fields.

- [ ] **Step 3: Implement snapshot loading and personalization wiring**

```go
func ScorePersonalRelevance(
    p profile.UserProfile,
    topicTags []string,
    styleTags []string,
    cognitiveTags []string,
    sourceName string,
) float64

type DryRunRequest struct {
    ConfigDir string
    OutputDir string
    StateDir  string
    Date      time.Time
    Mode      string
}

type DryRunHooks struct {
    // existing hooks...
    LoadProfile func(stateDir string) (profile.UserProfile, error)
}
```

Implementation notes:
- default `StateDir` to `state`
- if snapshot load fails with “file not found”, use a deterministic default profile and continue
- use snapshot data to set `signals.PersonalRelevance`
- enrich `why_for_you`, `TasteGrowthHint`, and `KnowledgeGapHint`
- update verification so generated cards still require non-empty safe personalized copy even when default-profile fallback is used
- keep `standard` before `brief` ranking behavior unchanged

- [ ] **Step 4: Run the affected packages**

Run: `go test ./tests/rank ./tests/analyze ./tests/run ./tests/verify -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/rank/scoring.go internal/analyze/pipeline.go internal/run/pipeline.go internal/verify/checks.go cmd/daily-builder/main.go tests/rank/scoring_test.go tests/analyze/pipeline_test.go tests/run/pipeline_test.go tests/verify/checks_test.go
git commit -m "feat: apply saved profile data to scoring and daily generation"
```

---

### Task 6: Document the feedback-loop runtime and validate the publishable path

**Files:**
- Modify: `README.md`
- Modify: `docs/index.md`
- Modify: `docs/ops/local-scheduler.md`

- [ ] **Step 1: Update the product and ops documentation**

Add concise documentation covering:

- what the feedback API does
- where state is stored (`state/feedback/`)
- how same-day profile/learning refresh differs from next-run full personalization
- how to start the feedback API locally
- that the daily builder still works when no snapshot exists

- [ ] **Step 2: Run the full automated test suite**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 3: Run a builder dry-run with the new profile-aware path**

Run: `go run ./cmd/daily-builder run --date 2026-03-19 --mode morning --dry-run`
Expected: PASS and generate a publishable edition under `output/2026/03/19`.

- [ ] **Step 4: Run the feedback API locally and sanity-check the endpoints**

Run: `AGENTIC_NEWS_FEEDBACK_ADDR=127.0.0.1:18081 go run ./cmd/feedback-api`
Expected: service starts and listens without crashing.

Then in another shell:

Run: `curl -sS -X POST http://127.0.0.1:18081/api/v1/feedback/events -H 'Content-Type: application/json' -d '{"event_type":"feedback_like","timestamp":"2026-03-19T09:00:00Z","edition_date":"2026-03-19","article_id":"a1","topic_tags":["policy"],"style_tags":["deep-analysis"],"cognitive_tags":["risk"]}'`
Expected: JSON response containing non-empty `focus_topics` or updated profile summary fields.

Run: `curl -sS http://127.0.0.1:18081/api/v1/profile`
Expected: JSON response containing `focus_topics`, `preferred_styles`, `cognitive_preferences`, and `last_updated_at`.

- [ ] **Step 5: Commit**

```bash
git add README.md docs/index.md docs/ops/local-scheduler.md
git commit -m "docs: describe feedback loop runtime and validation flow"
```
