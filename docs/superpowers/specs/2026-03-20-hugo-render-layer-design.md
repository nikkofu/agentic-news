# Hugo Render Layer Integration Design

- **Date:** 2026-03-20
- **Project:** agentic-news
- **Scope:** Introduce Hugo as an optional render layer for `agentic-news`, while keeping the current Go pipeline as the content engine and feedback backend.

## 1. Problem Statement

The current `agentic-news` system already has a capable Go pipeline for RSS ingestion, extraction, ranking, AI analysis, image archival, personalization, and feedback learning. It also has a working multi-theme static renderer. However, as the product moves toward:

- more theme diversity
- faster template iteration
- future AI-generated templates
- stronger editorial/content-oriented page composition

the current custom renderer risks becoming the bottleneck for presentation-layer evolution.

Hugo is a strong candidate for the presentation layer because it is fast, static-first, content-oriented, and theme-friendly. But replacing the entire product with Hugo would be a mistake, because the most differentiated parts of `agentic-news` are not static-site concerns:

- RSS collection and normalization
- extraction quality control
- standard/brief downgrade rules
- ranking and personalization
- user profile + feedback loop
- image archival integrity

The correct architectural question is not “Should Hugo replace the system?” but:

> “Should Hugo become a new render backend fed by the existing Go content engine?”

This design recommends **yes**.

## 2. Recommendation

Adopt **Hugo as a new render layer**, not as the application’s primary business runtime.

Recommended architecture:

`RSS/Extraction/AI/Ranking/Profile/Feedback (Go) -> Edition Package -> Hugo Themes -> Static Site`

This means:

- keep the current Go pipeline as the source of truth for all business logic
- define a stable, renderer-agnostic edition package contract
- add Hugo as an alternative static-site renderer
- keep the existing Go HTML renderer during migration as a fallback and regression baseline

This is an incremental architecture decision, not a rewrite.

## 3. Goals and Non-Goals

### 3.1 Goals

- Preserve the current production-grade Go pipeline
- Improve theme/template development velocity
- Make multi-theme delivery easier to maintain
- Enable future AI-assisted theme generation
- Separate content generation from presentation rendering
- Keep output fully static and locally openable
- Keep feedback/profile logic compatible with static publication

### 3.2 Non-Goals

This iteration does **not** aim to:

- replace RSS ingestion with Hugo
- move ranking/personalization into Hugo
- move feedback persistence into Hugo
- turn the product into a Hugo-only architecture
- remove the existing Go renderer immediately
- support arbitrary runtime theme switching

## 4. System Boundary

### 4.1 Go Content Engine Responsibilities

The Go system remains responsible for:

- RSS ingestion and deduplication
- article fetching and content extraction
- analysis generation
- ranking and publishability checks
- standard vs brief card classification
- user profile loading and relevance scoring
- feedback API and snapshot persistence
- conservative local image archival
- producing a normalized edition package

### 4.2 Hugo Responsibilities

Hugo becomes responsible for:

- homepage rendering
- article rendering
- section/list layouts
- taxonomy/archive organization if needed
- theme packaging and visual identity
- static output generation from prepared content

### 4.3 Feedback Runtime Boundary

Feedback remains outside Hugo:

- Hugo pages remain static
- browser JS continues to call the Go feedback API
- same-day feedback panels remain a frontend + API responsibility

This preserves the existing “static delivery + lightweight companion service” model.

## 5. Core Architectural Decision

### 5.1 Chosen Option: Hugo as Render Backend

Use Hugo only after the Go pipeline has already:

- selected featured items
- decided card type
- finalized article metadata
- archived images
- derived explanation/profile fields

Then the Go pipeline exports a Hugo-compatible content bundle.

### 5.2 Rejected Option: Hugo-First Full Rewrite

A Hugo-first full rewrite is rejected because it would force business logic into an environment not designed for:

- extraction pipelines
- ranking logic
- personalization logic
- durable feedback storage
- graceful degraded-card semantics

That would increase coupling and reduce product clarity.

## 6. Edition Package Contract

The new render boundary should be a normalized edition package, not raw internal Go structs.

### 6.1 Package Structure

Recommended structure:

- `content/issues/YYYY-MM-DD/_index.md`
- `content/issues/YYYY-MM-DD/posts/<pick-id>.md`
- `data/edition/YYYY-MM-DD.json`
- `data/profile/latest.json`
- `static/assets/issues/YYYY-MM-DD/images/...`
- `meta/edition.json`

Equivalent internal output roots are acceptable as long as the structure remains stable and theme-friendly.

### 6.1.1 Theme-Neutral Package Requirement

The edition package itself should be **theme-neutral**.

That means:

- one content package represents one edition’s editorial/content truth
- multiple themes render from the same package
- theme differences must live in Hugo themes/layouts, not in duplicated content exports

This preserves side-by-side theme comparison and prevents presentation-layer drift from contaminating the content contract.

### 6.2 Why Markdown Alone Is Not Enough

Pure Markdown is insufficient because the edition includes:

- ordered featured lists
- homepage lead/secondary/supporting grouping
- learning modules
- profile modules
- brief vs standard card behavior

Therefore the package must be:

- Markdown for article content
- front matter for page-local metadata
- JSON data files for edition-level structure
- static assets for archived original images

### 6.3 Article Front Matter Fields

Each article page should expose fields such as:

- `title`
- `date`
- `source_name`
- `source_url`
- `card_type`
- `fallback_reason`
- `cover_image`
- `score_final`
- `topic_tags`
- `style_tags`
- `cognitive_tags`
- `why_for_you`
- `taste_growth_hint`
- `knowledge_gap_hint`

### 6.4 Edition-Level Data

Edition-level data should include:

- featured ordering
- lead/secondary/supporting grouping
- edition keywords
- learning blocks
- metadata such as generated time and fallback state

## 7. Hugo Theme Contract

Hugo themes should not be treated as ad hoc skins. They should implement a stable theme contract.

### 7.1 Required Template Entry Points

Each theme should implement:

- `layouts/_default/baseof.html`
- `layouts/issues/list.html`
- `layouts/issues/single.html`
- `layouts/partials/hero.html`
- `layouts/partials/card-standard.html`
- `layouts/partials/card-brief.html`
- `layouts/partials/profile-panel.html`
- `layouts/partials/learning-panel.html`
- `layouts/partials/feedback-hooks.html`

### 7.2 Required Stable HTML Hooks

The following hooks should remain stable so the feedback/runtime layer does not drift:

- `data-theme-id`
- `data-article-id`
- `data-card-type`
- `data-reading-block`

### 7.3 Theme Metadata

Each theme should also carry metadata, for example in:

- `themes/<theme-id>/theme.yaml`

Suggested fields:

- theme name
- intended audience
- tone/style tags
- visual system tags
- layout density
- image emphasis
- version

This metadata is useful both for humans and for future AI theme generation workflows.

## 8. AI-Generated Theme Strategy

The long-term value of this architecture is not merely “using Hugo.” It is creating a safe target for AI-generated themes.

AI should generate only:

- layouts
- partials
- style tokens
- CSS/assets
- theme configuration

AI should **not** generate or modify:

- ingestion logic
- ranking logic
- feedback persistence
- profile aggregation
- extraction rules
- output contracts

This sharply reduces risk and keeps AI work bounded to the presentation layer.

## 9. Migration Strategy

## 9.1 Build and Output Model

The Hugo integration should preserve the current product requirement that multiple themes for the same edition can coexist locally without overwriting each other.

Recommended build behavior:

- produce one normalized edition package per date
- invoke Hugo once per theme against that same edition package
- emit output into theme-scoped roots such as:
  - `output/editorial-ai/YYYY/MM/DD`
  - `output/ai-product-magazine/YYYY/MM/DD`
  - `output/youth-signal/YYYY/MM/DD`
  - `output/soft-focus/YYYY/MM/DD`

Additional constraints:

- rendered pages must remain locally openable via relative asset paths
- archived images must remain edition-local static assets
- the Hugo build should not require a running web server to preview generated editions
- the final integration should preserve current side-by-side comparison convenience

### 9.2 Phase 1: Define the Edition Package

First, add a normalized content/package output without changing the current production renderer.

This phase decouples generation from rendering.

### 9.3 Phase 2: Add Hugo as a Second Renderer

Keep the existing Go renderer and add Hugo as a parallel render backend.

This enables:

- side-by-side comparisons
- regression checking
- local preview confidence
- low-risk experimentation

### 9.4 Phase 3: Migrate One Theme First

Migrate `editorial-ai` first.

Reason:

- it is the calmest and most “baseline” theme
- it is the easiest contract validator
- it reduces migration ambiguity

### 9.5 Phase 4: Migrate Remaining Themes

After `editorial-ai` proves the contract, migrate:

- `ai-product-magazine`
- `youth-signal`
- `soft-focus`

### 9.6 Phase 5: Decide Whether Hugo Becomes Default

Only promote Hugo to default after validating:

- content parity
- render fidelity
- local-open reliability
- theme development speed
- AI theme generation viability

Until then, the Go renderer remains the fallback and comparison baseline.

## 10. Benefits

### 10.1 Immediate Benefits

- better content-oriented templating ergonomics
- stronger multi-theme authoring workflow
- cleaner separation of content vs presentation

### 10.2 Medium-Term Benefits

- reusable theme contract
- easier onboarding for design/theme work
- improved AI-assisted theme generation workflow

### 10.3 Long-Term Benefits

- renderer-agnostic content assets
- safer experimentation with future frontends
- stronger editorial/product scalability

## 11. Risks

### 11.1 Dual-Renderer Complexity

During migration, there will be temporary cost:

- one content engine
- two renderers

This is acceptable and preferable to a risky rewrite.

### 11.2 Overfitting the Contract

If the edition package contract mirrors current templates too tightly, it will not age well. The contract should represent editorial/content semantics, not incidental HTML structure.

### 11.3 Pushing Dynamic Logic into Hugo

This must be avoided. Hugo should stay a render layer, not become the home of personalization or backend state transitions.

## 12. Decision Criteria for Defaulting to Hugo

Hugo should only become the default render path when all of the following are true:

1. `editorial-ai` achieves parity
2. at least one additional theme also works well
3. local asset pathing and image archival remain correct
4. brief/standard rendering remains honest
5. feedback hooks remain intact
6. new theme authoring is measurably faster
7. AI-generated theme prototypes are easier and safer than with the current renderer

## 13. Final Recommendation

Adopt Hugo as an **incremental, optional render layer** and design around a stable edition package contract.

Do **not** rewrite the product around Hugo.

The best next move is to create a dedicated implementation plan for:

- edition package export
- Hugo renderer scaffolding
- one-theme migration (`editorial-ai`)
- parity verification against the current renderer

This creates the foundation for a long-term theme factory without sacrificing the current product engine.
