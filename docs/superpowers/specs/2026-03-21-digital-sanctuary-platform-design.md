# Digital Sanctuary Platform Design Spec

- **Date:** 2026-03-21
- **Project:** agentic-news
- **Scope:** Introduce a unified `sanctuary-api` platform layer for the new `apps/` experience, covering content reading, personal growth, and RSS source management in the primary delivery slice, followed by reflections, while treating Community and Upgrade as placeholder contracts in the first release.

## 1. Problem Statement

`agentic-news` already has three meaningful foundations:

- a Go-based daily content pipeline
- static edition outputs and Hugo render support
- a lightweight feedback/profile learning loop

Those pieces are useful, but they are still shaped like a newsletter generator with companion interactions, not like the broader “Digital Sanctuary” product now being built in `apps/`.

The new product direction is materially different. The stitched page set describes a platform that combines:

- a daily editorial briefing
- article-level synthesis and learning affordances
- domain exploration
- knowledge gap mapping
- guided learning plans
- personal archives and cognitive profile surfaces
- daily reflections
- user-managed RSS inputs

That means the current architecture has a gap:

- the builder can generate editions, but there is no unified runtime API for the new application shell
- the feedback service can update a narrow profile snapshot, but it cannot serve the broader growth surfaces the new app requires
- there is no stable page-facing contract for the `apps/` team
- there is no clear boundary for where LLM enrichment belongs versus where deterministic platform behavior must remain available

The next milestone is therefore not “add a few more endpoints,” but:

> build a single backend platform surface that turns edition packages, user state, and curated LLM enhancement into a coherent API for the new Digital Sanctuary application.

## 2. Confirmed Product Decisions

The following decisions were explicitly confirmed during design:

1. The first release prioritizes **Content + Personal Growth + RSS Source management**.
2. **Community** and **Upgrade** should exist only as **placeholder contracts** in the first release.
3. A single backend owner should be responsible for:
   - API contracts
   - runtime backend service
   - LLM orchestration and fallback behavior
4. The new frontend under `apps/` should consume a **unified backend surface**, not raw builder files and not multiple disconnected services.
5. The current builder and static output pipeline should remain in place as the content production engine.
6. The first release should avoid unnecessary platform sprawl:
   - no full database migration
   - no payment backend
   - no real social graph
   - no auth-heavy multi-user architecture
7. LLM should be treated as an **enhancement layer**, not as the hard dependency for page availability.
8. Delivery should proceed in this order after the runtime skeleton is available:
   - Content
   - Personal Growth
   - RSS Source management
   - Reflection

## 3. Approaches Considered

### 3.1 Option A: Page-Oriented BFF

Build one backend-for-frontend endpoint per page, for example:

- `/api/apps/home`
- `/api/apps/article-detail`
- `/api/apps/profile`

#### Advantages

- fastest initial frontend hookup
- very direct mapping to stitched pages

#### Drawbacks

- duplicates logic across page handlers
- weak reuse between growth, profile, and reflection surfaces
- encourages backend contracts to ossify around temporary page composition
- becomes harder to evolve into a real platform

### 3.2 Option B: Unified Domain Service

Build one new runtime service with stable domain APIs for:

- content
- growth
- reflections
- RSS source management
- LLM enrichment

Pages then consume these stable read/write models.

#### Advantages

- aligns with the product direction of a real cognitive platform
- keeps business boundaries cleaner than page-shaped handlers
- allows frontends to evolve without rewriting core backend contracts
- makes LLM orchestration and fallback easier to centralize

#### Drawbacks

- requires more deliberate design up front
- first phase is slightly slower than a throwaway BFF

### 3.3 Option C: Immediate Full Platform Rebuild

Introduce a database, async job system, deeper runtime orchestration, and a broader platform rewrite immediately.

#### Advantages

- closer to a long-term production architecture

#### Drawbacks

- mismatched to current repo maturity
- slows down `apps/` integration
- creates too much surface area before the first product version is even stable

## 4. Recommendation

Choose **Option B: Unified Domain Service**.

Recommended architecture:

`daily-builder (content engine) -> edition package -> sanctuary-api -> apps frontend`

This means:

- keep the current Go build pipeline as the source of truth for content production
- add a unified runtime service for the app shell
- absorb the current narrow feedback/profile behavior into the broader platform service
- preserve deterministic content and state behavior
- allow LLM to enrich the product without controlling its availability

This is the most suitable option for a first release that still aims to become a real platform rather than a collection of stitched pages over mock data.

## 5. Goals and Non-Goals

### 5.1 Goals

- Provide a stable backend API for the new `apps/` frontend
- Preserve the current builder as the content engine
- Support first-release pages for:
  - Home / Briefing
  - Article Detail
  - Category / Domains
  - Knowledge Gap
  - Learning & Growth
  - Profile
  - RSS Source
- Support Reflection as the next slice after the primary delivery path above
- Provide stable placeholder contracts for:
  - Community
  - Upgrade
- Expand the current profile/learning loop into a broader growth model
- Introduce a structured article dossier model suitable for detailed article pages
- Centralize LLM invocation, caching, and fallback behavior
- Keep page reads deterministic and resilient even when LLM is slow or unavailable

### 5.2 Non-Goals

This release does **not** attempt to deliver:

- a real-time social/community backend
- subscriptions, billing, or checkout
- a general-purpose database-backed platform rewrite
- a multi-user identity and permissions system
- synchronous LLM execution on every read path
- replacement of the builder with a runtime-first content system

## 6. System Overview

The system becomes a two-runtime product:

- **Daily content engine (`daily-builder`)**
  - RSS ingest
  - extraction
  - ranking
  - analysis
  - edition generation
  - edition package generation

- **Application runtime (`sanctuary-api`)**
  - edition package reads
  - growth/profile state reads and writes
  - reflection management
  - RSS source management
  - LLM enrichment caching
  - page-facing API contracts

### 6.1 Existing Builder Responsibilities

The builder remains responsible for:

- collecting and normalizing source content
- selecting featured items
- computing scores and card types
- generating core editorial insights
- archiving selected images
- producing edition outputs and package artifacts

### 6.2 New Runtime Responsibilities

The new runtime service becomes responsible for:

- serving page-ready content APIs
- maintaining broader user state beyond current feedback snapshots
- managing reflections and RSS input overrides
- brokering enhancement-level LLM generation and cache reads
- exposing stable placeholder contracts for later platform modules

### 6.3 Feedback Service Integration

The current feedback runtime is not the correct long-term top-level boundary. Its behavior should be absorbed into `sanctuary-api` as a `growth` subdomain.

In first implementation terms:

- preserve compatibility with existing `feedback` store files
- migrate writes toward `state/sanctuary/...`
- stop designing new frontend contracts directly around `/api/v1/feedback/events`

## 7. Domain Modules

The unified runtime should be organized by domain, not by page filename.

Recommended internal modules:

- `content`
  - briefing
  - articles
  - domains
- `growth`
  - knowledge gaps
  - learning plan
  - profile map
  - targets
- `reflection`
  - entries
  - archive
  - summaries
- `rsssources`
  - source list
  - toggles
  - quality and density overrides
- `llm`
  - prompt rendering
  - provider execution
  - caching
  - traceability
- `placeholders`
  - community preview
  - upgrade offer

This structure keeps the backend stable even if the frontend changes route composition later.

## 8. Page-to-API Contract Mapping

The first release should expose domain-shaped APIs under `/api/v1/`.

### 8.1 Content Surfaces

- `GET /api/v1/briefing?date=YYYY-MM-DD`
  - powers Home / Briefing
- `GET /api/v1/articles/{article_id}`
  - powers Article Detail
- `POST /api/v1/articles/{article_id}/feedback`
  - powers article learning actions
- `GET /api/v1/domains`
  - powers Category / Domains browse view
- `GET /api/v1/domains/{domain_slug}`
  - powers future per-domain detail view

### 8.2 Personal Growth Surfaces

- `GET /api/v1/growth/knowledge-gaps`
  - powers Knowledge Gap
- `GET /api/v1/growth/learning-plan`
  - powers Learning & Growth
- `POST /api/v1/growth/targets`
  - create learning targets
- `PATCH /api/v1/growth/targets/{target_id}`
  - update target state or priority
- `POST /api/v1/growth/tasks/{task_id}/complete`
  - complete curated tasks
- `GET /api/v1/profile`
  - powers Profile
- `GET /api/v1/profile/archives`
  - powers paginated archive history

### 8.3 Reflection and Input Management

- `GET /api/v1/reflections`
  - powers Reflection
- `POST /api/v1/reflections`
  - archive a new reflection
- `GET /api/v1/rss/sources`
  - powers RSS Source
- `POST /api/v1/rss/sources`
  - add a new source
- `PATCH /api/v1/rss/sources/{source_id}`
  - update overrides
- `DELETE /api/v1/rss/sources/{source_id}`
  - remove a runtime-managed source or clear its override
  - config-backed seed sources are disabled via `PATCH`, not deleted

### 8.4 Placeholder Contracts

- `GET /api/v1/community/preview`
- `GET /api/v1/upgrade/offer`

These are intentionally read-only in the first release.

## 9. Data Architecture

The platform uses three data layers:

1. **Edition package layer**
2. **User/platform state layer**
3. **LLM enhancement cache layer**

### 9.1 Edition Package Layer

The builder already emits package-level content, but the first release needs a richer article contract for the application runtime.

Existing stable package artifacts include:

- `data/daily.json`
- `data/learning.json`
- `meta/edition.json`
- Hugo content markdown

That is sufficient for the homepage and basic listings, but not sufficient for the stitched article detail experience.

### 9.2 New Article Dossier Contract

The builder should add:

- `output/_packages/YYYY/MM/DD/data/articles/{article_id}.json`

This dossier should act as the canonical detailed article read model for the runtime API.

Suggested fields:

- `article_id`
- `edition_date`
- `title`
- `domain`
- `source_name`
- `source_url`
- `published_at`
- `hero_image`
- `summary_brief`
- `summary_deep`
- `key_points[]`
- `viewpoint`
- `opportunity_risk`
- `contrarian_take`
- `evidence_snippets[]`
- `concept_to_master`
- `quote`
- `reading_time_min`
- `topic_tags[]`
- `style_tags[]`
- `cognitive_tags[]`
- `why_for_you`
- `taste_growth_hint`
- `knowledge_gap_hint`

This keeps article detail pages decoupled from raw source extraction text while still being richer than the current front matter contract.

### 9.3 User and Platform State Layer

Recommended state roots:

- `state/sanctuary/profile/profile_snapshot.json`
- `state/sanctuary/growth/learning_snapshot.json`
- `state/sanctuary/growth/targets.json`
- `state/sanctuary/reflections/YYYY-MM.jsonl`
- `state/sanctuary/reflections/index.json`
- `state/sanctuary/rss/overrides.json`
- `state/sanctuary/rss/source_stats.json`

### 9.4 Compatibility Strategy

The first release should remain compatible with the current feedback snapshots:

- if `state/sanctuary/profile/profile_snapshot.json` is missing, fall back to `state/feedback/profile_snapshot.json`
- if `state/sanctuary/growth/learning_snapshot.json` is missing, fall back to `state/feedback/learning_snapshot.json`

New writes should go to the sanctuary state roots.

## 10. State Model Contracts

### 10.1 `profile_snapshot.json`

The profile snapshot should grow from the current lightweight structure into a fuller state model that can support Profile, Knowledge Gap, and Learning surfaces.

Suggested fields:

- `focus_topics`
- `preferred_styles`
- `cognitive_preferences`
- `recent_feedback_summary`
- `topic_affinity`
- `style_affinity`
- `cognitive_affinity`
- `source_affinity`
- `negative_signals`
- `domain_strengths`
- `reading_streak_days`
- `knowledge_points`
- `current_focus`
- `last_updated_at`

### 10.2 `learning_snapshot.json`

Suggested fields:

- `taste_growth_hint`
- `knowledge_gap_hint`
- `today_plan`
- `learning_tracks[]`
- `recommended_frontiers[]`
- `retention_metrics`
- `butler_suggestion`
- `updated_at`

### 10.3 `targets.json`

Suggested fields:

- `week_of`
- `items[]`
  - `target_id`
  - `title`
  - `status`
  - `source`
  - `domain`
  - `priority`

### 10.4 Reflection Store

Reflections should be append-only at the raw layer and indexed for fast listing.

Suggested write model:

- `state/sanctuary/reflections/YYYY-MM.jsonl`

Suggested index model:

- `state/sanctuary/reflections/index.json`

Each reflection entry should include:

- `reflection_id`
- `created_at`
- `content`
- `tags[]`
- `related_article_ids[]`
- `related_domains[]`
- `summary`
- `enhancement_status`

### 10.5 RSS Source Overrides

The first release should not mutate `config/rss_sources.yaml` directly.

Instead, runtime edits should write to:

- `state/sanctuary/rss/overrides.json`

This file should support:

- `enabled`
- `density_mode`
- `quality_override`
- `domain_override`
- `updated_at`

The builder then merges config sources and overrides when generating future editions.

To keep runtime behavior explicit, RSS source records should distinguish:

- `source_kind: seed | runtime`

Deletion rules:

- `seed` sources are never physically removed by the runtime
- `runtime` sources may be removed with `DELETE /api/v1/rss/sources/{source_id}`

## 11. LLM Responsibility Boundary

LLM must remain an enhancement layer, not the backbone of availability.

### 11.1 Deterministic First

The following pages must remain fully functional without LLM:

- Home / Briefing
- Category / Domains
- RSS Source
- base Profile and Learning metrics

### 11.2 LLM Enhancement Responsibilities

LLM is allowed to enrich:

- article deep synthesis
- `concept_to_master`
- knowledge gap explanation language
- butler suggestion wording
- cross-domain learning rationale
- reflection summaries and tag extraction

### 11.3 Read-Path Rule

No read API should require a fresh LLM call to return `200`.

Read APIs should:

- return deterministic content immediately
- include enhancement blocks when cache is available
- surface `enhancement_status` when cache is absent or degraded

### 11.4 Write-Path Rule

Writes may trigger synchronous light generation only when:

- the user explicitly performed a creation action
- the generation can complete within a tight timeout budget

Otherwise they should:

- persist first
- enqueue or trigger background enhancement generation
- return success without blocking on a full LLM response

## 12. API Protocol Rules

### 12.1 General Rules

- All APIs return JSON
- All timestamps use RFC3339 UTC
- All date filters use `YYYY-MM-DD`
- All resource IDs are stable strings
- Collection endpoints return:
  - `items[]`
  - `next_cursor`
  - `total_estimate`

### 12.2 Idempotency

All write APIs should support idempotency through:

- `Idempotency-Key` request header

The runtime may also accept a body field fallback during transition, but the header should be treated as canonical.

### 12.3 Error Contract

All errors should use a consistent structure:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "feed_url is required",
    "details": {
      "field": "feed_url"
    },
    "request_id": "req_..."
  }
}
```

Suggested first-release error codes:

- `bad_request`
- `validation_failed`
- `not_found`
- `conflict`
- `unsupported_action`
- `invalid_cursor`
- `invalid_date`
- `source_unreachable`
- `feed_parse_failed`
- `internal_error`

### 12.4 Enhancement Status Contract

Any endpoint with LLM-backed optional fields should expose:

- `ready`
- `pending`
- `degraded`
- `unavailable`

This allows the frontend to render deterministic fallbacks without API-level failure.

## 13. Placeholder Contract Policy

Community and Upgrade should exist as stable preview contracts in release one, but they must remain explicitly non-production:

- `GET /api/v1/community/preview`
  - read-only
  - returns preview data only
- `GET /api/v1/upgrade/offer`
  - read-only
  - returns offer copy and pricing display only

The first release should not expose:

- community posting
- circle membership writes
- DMs or messaging
- payment intent creation
- subscription lifecycle mutation

This preserves honest scope and keeps platform boundaries clean.

## 14. Rollout Plan

Implementation should proceed in phases.

### Phase 0: Runtime Skeleton

- create `sanctuary-api`
- add health/error/request-id middleware
- expose preview placeholder endpoints
- add edition package readers
- add sanctuary state readers with feedback fallback

### Phase 1: Content Read APIs

- add article dossier package output
- implement briefing, article, and domains reads
- keep all responses deterministic-first

### Phase 2: Growth and RSS Runtime

- expand profile snapshot
- add learning targets and growth endpoints
- add RSS source override management
- add source stats aggregation
- absorb semantic feedback actions

### Phase 3: Reflection

- add reflection persistence and archive reads

### Phase 4: Enhancement and Hardening

- add LLM cache layer
- enrich article and growth surfaces
- harden fallback behavior
- add contract verification and compatibility tests

## 15. Risks and Mitigations

### 15.1 Risk: Builder and runtime contracts drift

Mitigation:

- version package schema explicitly
- add contract tests over package readers
- treat dossier generation as part of build verification

### 15.2 Risk: LLM slows page reliability

Mitigation:

- never require synchronous LLM for read availability
- cache enhancement outputs
- expose `enhancement_status`

### 15.3 Risk: Runtime state becomes fragmented

Mitigation:

- centralize new writes under `state/sanctuary`
- keep temporary compatibility reads explicit and narrow
- avoid long-lived dual-write logic

### 15.4 Risk: Frontend binds to temporary shape

Mitigation:

- define stable domain APIs before implementation
- keep placeholder preview endpoints explicit rather than vague

## 16. Success Criteria

This design is successful when:

1. The `apps/` frontend can consume a single backend surface for all first-release pages.
2. Content reading pages are powered by stable edition package reads rather than ad hoc file access.
3. Growth pages are powered by persisted state rather than one-off UI-only computations.
4. Reflection and RSS source management are writable through explicit runtime contracts.
5. Community and Upgrade are truthfully represented as preview-only placeholder APIs.
6. Page availability does not depend on live LLM success.
7. The backend remains consistent with the current Go-first architecture instead of bypassing it.

## 17. Next Step

After this spec is reviewed and approved, the next step is to produce phased implementation plans rather than one monolithic plan. The plans should break the work into concrete delivery slices:

- service skeleton
- package schema extension
- content APIs
- growth APIs
- RSS management APIs
- reflection APIs
- LLM cache integration

The first plan should cover the runtime skeleton plus the primary delivery path (Content + Personal Growth + RSS Source). Reflection and LLM hardening should follow as separate plans or clearly separated plan phases with their own verification checkpoints.
