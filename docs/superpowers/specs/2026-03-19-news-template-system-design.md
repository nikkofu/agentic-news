# News Template System Design Spec

- **Date:** 2026-03-19
- **Project:** agentic-news
- **Scope:** Introduce a publishable, build-time-selectable template system for homepage and article pages, with first-party visual themes, layout presets, and local image fallback using original article assets only.

## 1. Problem Statement

The current `agentic-news` output is functionally correct and locally deliverable, but the presentation layer is still too generic for a real first-scene deployment. The user wants the generated news edition to feel significantly more premium, current, and differentiated.

The design direction has evolved from a restrained newspaper-like layout toward a more forward-looking AI media product. That means the system should no longer hard-code a single visual style. Instead, it should support multiple built-in templates that can be selected at build time, while still preserving:

1. static, locally openable output
2. honest use of article-originated imagery
3. a clean separation between content generation and presentation

At the same time, a single article should no longer display a missing visual block when `og:image` is absent. If the article itself contains usable images, the pipeline should preserve original source material by selecting the first valid body image, downloading it into the edition's local asset library, and rendering that local file instead of inventing new artwork.

## 2. Confirmed Product Decisions

1. The system should support **multiple built-in templates** rather than a single visual redesign.
2. The first implementation phase uses **build-time template selection**, not in-page runtime switching.
3. Templates should affect **both visual style and layout**, not just colors.
4. The first built-in templates are:
   - `editorial-ai`
   - `ai-product-magazine`
   - `youth-signal`
   - `soft-focus`
5. The default template should be `editorial-ai`.
6. Tailwind CSS is acceptable, but only as a **build-time tool** that emits local static CSS. The final edition must not depend on external CDN resources.
7. Generated output must remain **pure static artifacts** that can be opened locally.
8. Image usage must remain honest:
   - prefer `og:image`
   - otherwise scan body images in source order
   - skip obviously invalid images
   - download only original article assets
   - never invent or generate imagery
9. The initial template system should not alter content ranking, analysis, feedback semantics, or profile logic. It is a presentation-layer capability.

## 3. Goals and Non-Goals

### 3.1 Goals

- Make the news edition feel premium and differentiated enough for real user-facing delivery.
- Support multiple built-in presentation styles using one content pipeline.
- Preserve local preview reliability and local artifact completeness.
- Upgrade both homepage and article-page information architecture.
- Add local image archival for articles that lack a pre-declared cover image but include suitable inline imagery.

### 3.2 Non-Goals

- No runtime theme switcher in this phase.
- No user-authored custom template editor in this phase.
- No generated illustrations, stock substitution, or AI-created imagery.
- No changes to RSS ingestion, ranking semantics, or feedback persistence beyond what is necessary to expose template selection and local image paths.
- No remote publishing transport redesign in this phase.

## 4. Template Model

Templates become a first-class concept in the render layer. A template is not just a CSS skin; it is a **theme package** with two parts:

1. **Visual tokens**
   - color palette
   - typography family and scale
   - spacing rhythm
   - radius and border rules
   - shadow intensity
   - gradient usage
   - glass / blur usage

2. **Layout preset**
   - homepage masthead treatment
   - lead-story and secondary-story arrangement
   - section density and grid rules
   - article-page main-column / side-column ratio
   - placement of recommendation, profile, feedback, and learning modules

Templates should apply consistently across homepage and article pages, while still allowing page-specific layout choices under the same template identity.

## 5. First Built-In Templates

### 5.1 `editorial-ai`

- **Intent:** premium, calm, trustworthy, editorial
- **Visual:** light surface, restrained cool accents, subtle gradients, minimal glass
- **Homepage:** lead story + layered editorial grid
- **Article page:** reading-first, quiet side notes
- **Role:** default template for broad deployment

### 5.2 `ai-product-magazine`

- **Intent:** forward-looking, branded, future-facing
- **Visual:** stronger gradients, dark-to-rich surfaces, glass layering, luminous accents
- **Homepage:** more dramatic first fold, stronger visual framing
- **Article page:** still readable, but with richer atmosphere
- **Role:** demo / showcase template for AI-native product positioning

### 5.3 `youth-signal`

- **Intent:** energetic, lighter, faster-scanning
- **Visual:** higher contrast, brighter accents, more expressive labeling
- **Homepage:** faster scan rhythm and more dynamic blocks
- **Article page:** lighter auxiliary modules and quicker reading pace
- **Role:** younger audience / signal-driven consumption

### 5.4 `soft-focus`

- **Intent:** softer, friendlier, more lifestyle-adjacent
- **Visual:** softer palette, gentler borders, more relaxed spacing
- **Homepage:** more breathable blocks and lower visual aggression
- **Article page:** more companion-like and approachable
- **Role:** users who prefer a gentler emotional tone and softer aesthetics

## 6. Visual and Layout Rules

### 6.1 Homepage

All templates must support a premium, publication-like homepage with:

- a recognizable masthead / edition identity
- a clear lead-story hierarchy
- secondary and tertiary story layers
- lower “card” feeling than the current implementation
- more editorial density without becoming cluttered
- quieter treatment of personalization and learning modules

Template-specific variation is allowed in:

- masthead tone
- hero prominence
- grid density
- section separation
- summary length and rhythm
- image emphasis

### 6.2 Article page

All templates must keep article pages **reading-first**. The primary reading column should remain calm even when a template uses stronger gradients or glass effects elsewhere.

Template-specific variation is allowed in:

- title area framing
- source/meta presentation
- image framing
- side-note vs inline module placement
- tone of recommendation / feedback / learning containers

But all templates must preserve:

- obvious article title
- obvious source / original-link access
- readable summary and article context
- non-intrusive recommendation / feedback modules

## 7. Image Preservation and Local Asset Library

### 7.1 Primary image selection

For each article:

1. If `og:image` exists and is usable, keep it as the first-choice cover.
2. Otherwise, inspect inline body images in original document order.
3. Skip clearly invalid images and continue scanning until the first valid candidate is found.
4. Download the selected source image into the current edition’s local asset library.
5. Render the local file path in generated HTML when a local asset was saved successfully.

### 7.2 Invalid image heuristics

The system should skip images that are obviously not article illustrations, including:

- site logos
- author avatars / profile headshots
- tracking or analytics pixels
- very small icons
- decorative SVG marks or tiny badges

The implementation should prefer explicit, conservative heuristics rather than guessing aggressively. If no valid image can be found confidently, the edition should omit the image rather than fabricate a better-looking result.

### 7.3 Local asset pathing

The edition output should gain a stable local image area under the date-specific output root, for example:

- `images/`
- or `assets/images/`

The exact directory can be chosen during implementation, but it must be:

- edition-local
- static-file friendly
- referenced with relative paths from rendered pages

## 8. Build-Time Template Selection

The CLI / build flow should accept a template identifier such as:

- `--theme editorial-ai`

If no template is specified, it should default to `editorial-ai`.

The selected template should influence rendering without changing the content generation contract. The same set of picks should be renderable through multiple templates.

## 9. Output Strategy

The first implementation should prefer **theme-separated output roots** so multiple versions of the same edition can coexist for comparison and review.

For example, the implementation may choose an output structure equivalent to:

- one directory per date + theme
- or one date directory containing theme-specific subdirectories

The final choice should prioritize:

- local preview convenience
- side-by-side comparison
- no accidental overwrite between themes

## 10. Tailwind CSS Strategy

Tailwind may be introduced, but only under these constraints:

1. Tailwind is a **build-time dependency**, not a runtime network dependency.
2. Generated static output must reference only local CSS.
3. Template tokens should map cleanly to Tailwind-friendly utility composition and/or compiled component classes.
4. The project should avoid coupling render logic to opaque class chaos; maintainability matters because multiple templates will exist.

This means the system may use Tailwind to accelerate layout, spacing, gradients, blur, and responsive structure, while still producing a portable static artifact.

## 11. Architecture Boundaries

### 11.1 Presentation-layer ownership

The template system owns:

- page-level theme selection
- CSS and layout variation
- component presentation
- local image rendering paths

### 11.2 Content-layer ownership

The existing pipeline still owns:

- RSS fetching
- item dedupe
- article extraction
- analysis
- ranking
- profile and feedback semantics

The only cross-boundary changes allowed are those needed to carry presentation metadata such as:

- selected theme ID
- local image path when downloaded

## 12. File-Level Design

### 12.1 Rendering

- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `internal/render/templates.go`
- Modify: `web/templates/index.tmpl`
- Modify: `web/templates/article.tmpl`

These changes should make rendering theme-aware and layout-aware.

### 12.2 Model updates

- Modify: `internal/model/types.go`

The model should carry enough presentation data to render local image paths and template context cleanly.

### 12.3 Content extraction / media support

- Modify: `internal/content/extractor.go`

The extractor / related content layer should expose enough structured image information to support:

- `og:image` preference
- ordered body-image fallback
- local archival decisions

### 12.4 Output helpers

- Modify or add: output / asset helper files as needed

The implementation will likely need dedicated helpers for:

- local image download
- safe filename derivation
- relative asset path resolution

### 12.5 Static style assets

- Modify or add: `web/static/*`
- Add build assets for Tailwind if adopted

Template styles should be organized so that adding a fifth or sixth template does not require rewriting the entire default stylesheet.

## 13. Testing and Verification Expectations

The implementation should prove:

1. theme selection changes output deterministically
2. homepage and article page both render the selected template
3. output remains locally openable without networked CSS dependencies
4. image fallback preserves original article imagery only
5. invalid body images are skipped in favor of the next valid image when available
6. if no valid image exists, rendering remains clean without fabricated fallback
7. multiple themes can coexist without clobbering each other’s output

Tests should cover both template selection and image fallback behavior.

## 14. Acceptance Criteria

This design is considered satisfied when:

- the build command can generate at least the first built-in templates by theme ID
- homepage and article page differ meaningfully across templates in both style and layout
- the default theme is premium enough for first real-scene usage
- static output opens locally with local CSS only
- articles without `og:image` can use the first valid inline image from the source article
- selected inline images are archived into the edition output instead of only hot-linking remote URLs
- no generated or invented imagery is introduced

## 15. Implementation Ordering Guidance

Recommended order:

1. establish theme selection skeleton
2. implement `editorial-ai`
3. implement `ai-product-magazine`
4. add local image archival + fallback selection
5. add `youth-signal`
6. add `soft-focus`

This ordering gets the system to a meaningful, high-quality default quickly, while still proving the multi-template architecture before expanding the full library.
