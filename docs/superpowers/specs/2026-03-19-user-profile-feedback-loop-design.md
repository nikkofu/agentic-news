# User Profile and Feedback Loop Design Spec

- **Date:** 2026-03-19
- **Project:** agentic-news
- **Scope:** Add a publishable, single-user profile and feedback learning loop to the existing static daily edition system using a lightweight co-located Go feedback service, append-only event storage, and derived profile snapshots.

## 1. Problem Statement

The current project already generates a publishable daily edition, but personalization is still mostly static. The existing `profile.UserProfile` type is minimal, there is no production feedback ingestion path, no persistent event stream, no profile snapshot lifecycle, and no way for user feedback to reliably affect the next edition.

That leaves the product short of the originally intended “news butler” behavior:

- the user can read a personalized product, but cannot teach it reliably
- the system cannot explain profile evolution using real evidence
- the next run cannot consume stable, persisted user preference changes
- the frontend has no lightweight interaction loop for “why for you” and learning feedback

The next milestone is to add a **complete but lightweight feedback loop** that remains compatible with the current publishable static-output architecture.

## 2. Confirmed Product Decisions

1. The first version should implement the **more complete loop** rather than only explicit feedback.
2. The deployed product remains primarily a **static site**, not a full dynamic application.
3. Feedback collection should use a **lightweight Go service** deployed on the same host as the static site.
4. The first version uses **no authentication** at the application layer, assuming the site itself is private.
5. Storage should keep:
   - raw append-only event history
   - derived latest profile snapshot
6. The profile should learn across all of these dimensions:
   - topics
   - content styles
   - cognitive preferences
   - with room to extend later
7. The feedback loop should be **split by time horizon**:
   - same-day feedback updates profile panels and learning hints immediately
   - next daily generation fully applies the profile to ranking and editorial explanation
8. The design must preserve the current standard:
   - edition remains publishable even if feedback service is unavailable
   - personalization is an enhancement, not a publish blocker

## 3. Goals and Non-Goals

### 3.1 Goals

- Collect explicit and implicit feedback from the published site
- Persist feedback in a durable, inspectable format
- Build a derived user profile snapshot from events
- Apply profile state to:
  - ranking
  - `why_for_you`
  - learning suggestions
- Show a same-day visible loop on article pages without converting the product into an SPA
- Keep the architecture simple enough for a single-user local/cloud MVP

### 3.2 Non-Goals

This iteration does **not** include:

- multi-user accounts
- login or session management
- per-device identity stitching
- online realtime homepage reranking
- heavy analytics pipelines
- external database infrastructure
- replacing the static publication model

## 4. Architecture Overview

The system remains a two-part product:

- **Daily edition generator (Go CLI):**
  RSS ingest, extraction, analysis, ranking, render, verify, publishable local artifact generation

- **Feedback loop service (Go HTTP):**
  Accepts user events, appends them to disk, refreshes a derived profile snapshot, and serves lightweight profile / learning endpoints back to the published pages

### 4.1 Deployment topology

The feedback service is deployed on the **same cloud host** as the static site and Nginx.

Recommended serving arrangement:

- `Nginx` serves static edition files
- `Nginx` reverse-proxies `/api/v1/feedback/*` and `/api/v1/profile/*` to the local Go service
- The Go service writes to local runtime storage under `state/`

This keeps the product operationally simple and avoids coupling the edition build process to an external backend.

### 4.2 Runtime data flow

There are two distinct flows.

#### Same-day interaction flow

1. User opens the generated static index or article detail page
2. Frontend script records interaction events and explicit feedback
3. Frontend sends events to the feedback API
4. Feedback API:
   - validates the request shape
   - appends the raw event to the event stream
   - recomputes the latest profile snapshot
   - returns a lightweight updated profile / learning payload
5. Frontend refreshes only local explanation surfaces:
   - profile summary panel
   - “why this matters to you” companion text
   - learning suggestions panel

#### Next-run generation flow

1. `daily-builder run` starts
2. The run pipeline loads the latest profile snapshot from disk
3. Ranking uses that profile to compute `PersonalRelevance`
4. Analysis/explanation generation uses the profile to build richer `why_for_you`
5. Daily learning suggestions derive from the latest snapshot
6. Rendering writes the personalized outputs into that edition’s artifacts

### 4.3 Publishability boundary

The daily edition must remain publishable **without** the feedback service being healthy at generation time.

Therefore:

- feedback collection failure must not break reading
- feedback service outage must not block daily edition generation
- missing profile snapshot must fall back to a deterministic default profile
- the existing publishability thresholds remain unchanged

## 5. Storage Design

The storage model deliberately uses simple files rather than a database.

### 5.1 State directory layout

Recommended runtime structure:

- `state/feedback/events/YYYY-MM.jsonl`
- `state/feedback/profile_snapshot.json`
- `state/feedback/learning_snapshot.json`

Optional later additions can include compaction artifacts or archived summaries, but they are not required for this version.

### 5.2 Raw event stream

The raw event log is the system of record.

Properties:

- append-only
- newline-delimited JSON (`JSONL`)
- no in-place mutation
- easy to inspect, replay, and re-derive

Each event should include at minimum:

- `event_id`
- `event_type`
- `timestamp`
- `edition_date`
- `article_id`
- `article_title`
- `article_url`
- `source_name`
- `topic_tags[]`
- `style_tags[]`
- `cognitive_tags[]`
- `metadata`

Supported event types in this iteration:

- `article_view`
- `dwell_report`
- `feedback_like`
- `feedback_dislike`
- `bookmark`
- `unbookmark`
- `detail_expand`
- `revisit`
- `profile_panel_view`

### 5.3 Derived profile snapshot

`profile_snapshot.json` stores the latest derived user state for generation and UI reads.

It should include:

- `focus_topics`
- `preferred_styles`
- `cognitive_preferences`
- `topic_affinity`
- `style_affinity`
- `cognitive_affinity`
- `source_affinity`
- `negative_signals`
- `recent_feedback_summary`
- `last_updated_at`

The snapshot is a **materialized view**, not the source of truth. It can always be rebuilt from the event stream.

### 5.4 Derived learning snapshot

`learning_snapshot.json` is a small read-optimized companion for same-day UI refresh.

It should include:

- `taste_growth_hint`
- `knowledge_gap_hint`
- `today_plan`
- `learning_tracks[]`
- `updated_at`

This can be regenerated at the same time as the profile snapshot.

## 6. Profile Model

The existing `profile.UserProfile` type must be expanded from a minimal topic-only structure into a fuller single-user preference model.

The model should support:

- stable high-level focus areas
- weighted per-tag affinity maps
- explicit negative preferences
- source preferences
- recent feedback summary for explanation

Suggested fields:

- `FocusTopics []string`
- `PreferredStyles []string`
- `CognitivePreferences []string`
- `ExplicitFeedback map[string]string`
- `BehaviorSignals map[string]float64`
- `TopicAffinity map[string]float64`
- `StyleAffinity map[string]float64`
- `CognitiveAffinity map[string]float64`
- `SourceAffinity map[string]float64`
- `NegativeSignals map[string]float64`
- `LastUpdatedAt time.Time`

The exact struct can be refined during implementation, but the model must support all three learning dimensions:

- topic
- style
- cognition

## 7. Event-to-Profile Update Rules

The first version should use deterministic weighted updates rather than a learned model.

### 7.1 Update philosophy

- explicit signals should dominate implicit ones
- negative feedback should meaningfully reduce repeat promotion
- passive signals should have caps to avoid accidental over-weighting
- the model should remain explainable from event history

### 7.2 Signal strength classes

Recommended weighting direction:

- `feedback_like`: strong positive
- `feedback_dislike`: strong negative
- `bookmark`: medium-strong positive
- `unbookmark`: medium negative
- `dwell_report`: weak-to-medium positive with cap
- `detail_expand`: weak positive
- `revisit`: weak positive
- `article_view`: minimal evidence only

Updates should apply to:

- article topic tags
- article style tags
- article cognitive tags
- optionally source name

### 7.3 Negative preference handling

Negative feedback should update both:

- the relevant affinity maps with negative weight
- `negative_signals` for easier ranking penalties and explainability

This is important so the next edition does not keep surfacing categories the user explicitly suppressed.

## 8. Frontend Interaction Design

The first version should focus interaction on the article detail page, with minimal but useful support on the index page.

### 8.1 Article page features

The article page should gain:

- explicit feedback controls:
  - `值得多给我`
  - `少一点这类`
  - `收藏`
- optional reason tags, for example:
  - `主题相关`
  - `太浅`
  - `太观点化`
  - `想看更多数据`
  - `更关注风险`
- a “why recommended for you” explanation area
- a lightweight profile-change panel
- a lightweight learning suggestion panel

Same-day interactions should update those explanation surfaces without turning the page into a full client-rendered app.

### 8.2 Index page behavior

The index page should:

- track click-through into article pages
- optionally expose a compact “today’s profile” entry point
- remain statically ordered until the next edition

There is **no requirement** for same-day homepage reranking in this iteration.

### 8.3 Frontend resilience

If the feedback API is unavailable:

- the page remains fully readable
- controls can show a lightweight sync-failed hint
- no hard dependency is introduced into the reading path

## 9. API Design

The feedback service should stay intentionally small.

Required endpoints:

- `POST /api/v1/feedback/events`
- `GET /api/v1/profile`
- `GET /api/v1/profile/learning`

### 9.1 `POST /api/v1/feedback/events`

Purpose:

- receive one feedback event
- append it to the event log
- recompute the profile snapshot
- recompute the learning snapshot
- optionally return the latest lightweight profile state

Required behavior:

- reject malformed payloads
- require `event_type`, `timestamp`, `article_id`
- tolerate optional metadata fields
- write event before recomputation

### 9.2 `GET /api/v1/profile`

Purpose:

- return the latest lightweight profile snapshot for UI use

This endpoint should expose only what the page needs for explanation and profile summary, not raw internal bookkeeping unless it is directly useful.

Example response shape:

```json
{
  "focus_topics": ["ai infrastructure", "policy", "semiconductors"],
  "preferred_styles": ["data-driven", "deep-analysis"],
  "cognitive_preferences": ["risk", "framework"],
  "recent_feedback_summary": [
    "你最近连续强化了 AI 基础设施与政策影响类内容",
    "你对纯观点型内容的偏好在下降"
  ],
  "last_updated_at": "2026-03-19T09:15:00Z"
}
```

### 9.3 `GET /api/v1/profile/learning`

Purpose:

- return the latest same-day learning hints derived from the newest snapshot

This allows the article page to refresh learning guidance after feedback without reloading the edition.

## 10. Integration with Ranking

The ranking model already includes `PersonalRelevance`.

This iteration changes its source of truth:

- instead of using a mostly static placeholder value
- compute it from the latest user profile snapshot

Inputs to personalization scoring should include:

- topic tag affinity
- style affinity
- cognitive affinity
- negative signal penalties
- optional source affinity

Constraints:

- keep the existing overall score formula
- keep deterministic behavior
- preserve `standard` cards ranking ahead of `brief` cards
- do not let personalization violate publishability rules

## 11. Integration with Analysis and Explanation

`analyze.RunPipeline(...)` already accepts a `profile.UserProfile`. This iteration makes that real.

### 11.1 `why_for_you`

`why_for_you` should be generated from actual profile matches, for example:

- recent topic affinity
- preference for a content style
- preference for a cognitive frame such as risk, opportunity, contrarian, or framework-building

The explanation should be concrete enough to feel personal, but honest about the underlying evidence.

### 11.2 Brief-card treatment

Brief cards should still get an honest personalized explanation when possible, but they must not impersonate a completed deep-analysis card.

In other words:

- it is acceptable to say why the topic likely matters to the user
- it is not acceptable to fabricate deeper personalized analysis when the card is degraded

## 12. Integration with Learning Suggestions

There are two learning horizons:

- **same-day learning refresh** from the feedback service
- **next-edition learning generation** from the run pipeline

Learning output should stay aligned to the original product direction:

- `taste` guidance
- `cognition` guidance
- `knowledge` guidance

The system should also be able to surface balancing advice, for example if recent feedback over-concentrates around one lens such as only “opportunity” content.

## 13. File-Level Design

### Core model and profile logic

- Modify: `internal/profile/profile.go`
  - expand profile structure
  - add deterministic event-application and aggregation helpers

- Modify: `internal/model/types.go`
  - add fields required for article-level tags and client-side feedback surfaces if needed

### Ranking and analysis

- Modify: `internal/rank/scoring.go`
  - compute profile-aware personalization score

- Modify: `internal/analyze/pipeline.go`
  - produce profile-aware `why_for_you`, taste, and knowledge hints

- Modify: `internal/run/pipeline.go`
  - load latest profile snapshot before scoring/analyzing
  - fall back to default profile if snapshot is absent

### Feedback service

- Create: `internal/feedback/`
  - event validation
  - event append storage
  - snapshot derivation
  - HTTP handlers

- Create: `cmd/feedback-api/main.go`
  - lightweight service entrypoint

### Frontend and rendering

- Modify: `web/templates/article.tmpl`
  - feedback controls
  - profile panel
  - learning refresh surface

- Modify: `web/templates/index.tmpl`
  - click tracking hooks
  - optional compact profile summary entry

- Modify: `web/static/app.js`
  - event capture
  - feedback submission
  - same-day panel refresh

### Verification and docs

- Modify: `internal/verify/checks.go`
  - validate personalized output fallback behavior where relevant

- Modify: `docs/index.md`
  - reflect that the product includes a lightweight feedback learning loop

## 14. Error Handling

### 14.1 Feedback service unavailable

- reading still works
- frontend degrades gracefully
- daily generation still works from latest available snapshot or default profile

### 14.2 Event write failure

- request returns failure
- frontend may show a lightweight retry hint
- no partial snapshot update without event durability

### 14.3 Snapshot recomputation failure

- keep the raw event append if already durable
- leave previous snapshot in place
- surface an actionable server error

### 14.4 Missing or corrupt snapshot at run time

- rebuild if feasible in a later iteration
- for this version, fall back to default profile and continue generation

## 15. Verification and Testing Strategy

### 15.1 Unit tests

- event validation
- event append/read behavior
- profile aggregation math
- ranking personalization adjustments
- `why_for_you` generation
- learning suggestion generation

### 15.2 API tests

- `POST /api/v1/feedback/events`
- `GET /api/v1/profile`
- `GET /api/v1/profile/learning`

### 15.3 Integration tests

- seed event stream
- derive profile snapshot
- run `daily-builder run --dry-run`
- verify personalized fields appear in generated artifacts

### 15.4 Smoke tests

- run static edition locally
- run feedback API locally
- simulate click / dwell / explicit feedback
- verify same-day profile panel refresh
- verify next-run personalization changes output

## 16. Success Criteria

This design is successful when all of the following are true:

1. Published pages can collect explicit and implicit feedback through a lightweight API.
2. Feedback is durably stored as append-only events under `state/feedback/`.
3. The system produces a readable latest profile snapshot and learning snapshot.
4. Same-day article pages can refresh profile and learning surfaces after feedback without breaking static-readability.
5. The next edition run consumes the saved profile and changes:
   - personalization scoring
   - `why_for_you`
   - learning suggestions
6. The daily build remains publishable even when no feedback data exists.
7. The daily build remains publishable even if the feedback service is temporarily unavailable.
