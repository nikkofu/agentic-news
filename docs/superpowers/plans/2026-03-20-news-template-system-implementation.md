# News Template System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a build-time-selectable multi-template news presentation system with four built-in themes, theme-scoped output roots, and honest local cover-image archival that reuses original article imagery only.

**Architecture:** Keep the existing RSS → extract → analyze → rank → render pipeline intact, but thread a validated `theme_id` through CLI, run, model, render, and output helpers so the same content can be rendered through multiple layouts. Introduce a small render theme registry, theme-specific Go templates, Tailwind-built local CSS, and a separate image archival helper that selects `og:image` or the first valid body image, downloads it into the edition asset tree, and renders relative local paths when available.

**Tech Stack:** Go 1.22+, stdlib, existing `internal/run`, `internal/render`, `internal/content`, `internal/output` packages, Go templates, Tailwind CSS CLI via `npm`, `go test`.

**Execution Notes:** Follow @superpowers:test-driven-development task-by-task and @superpowers:verification-before-completion before claiming any task is complete.

---

## File Structure Plan

### Core files to modify

- Modify: `internal/model/types.go`
- Modify: `internal/run/options.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/content/extractor.go`
- Modify: `internal/output/writer.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `internal/render/templates.go`
- Modify: `cmd/daily-builder/main.go`
- Modify: `web/static/styles.css`
- Modify: `README.md`
- Modify: `.gitignore`

### Core files to create

- Create: `internal/render/theme.go`
- Create: `internal/output/images.go`
- Create: `tests/run/options_test.go`
- Create: `tests/output/images_test.go`
- Create: `web/templates/themes/editorial-ai/index.tmpl`
- Create: `web/templates/themes/editorial-ai/article.tmpl`
- Create: `web/templates/themes/ai-product-magazine/index.tmpl`
- Create: `web/templates/themes/ai-product-magazine/article.tmpl`
- Create: `web/templates/themes/youth-signal/index.tmpl`
- Create: `web/templates/themes/youth-signal/article.tmpl`
- Create: `web/templates/themes/soft-focus/index.tmpl`
- Create: `web/templates/themes/soft-focus/article.tmpl`
- Create: `package.json`
- Create: `package-lock.json`
- Create: `tailwind.config.js`
- Create: `web/tailwind/input.css`

### Tests to modify

- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/render/render_test.go`
- Modify: `tests/content/extractor_test.go`

### Existing files to reference

- Reference: `docs/superpowers/specs/2026-03-19-news-template-system-design.md`
- Reference: `web/static/app.js`
- Reference: `internal/analyze/pipeline.go`
- Reference: `internal/rss/fetcher.go`
- Reference: `internal/rss/dedupe.go`

---

### Task 1: Add validated theme selection and theme-scoped output roots

**Files:**
- Modify: `internal/model/types.go`
- Modify: `internal/run/options.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/output/writer.go`
- Modify: `cmd/daily-builder/main.go`
- Create: `tests/run/options_test.go`
- Modify: `tests/run/pipeline_test.go`

- [ ] **Step 1: Write the failing tests for theme parsing and output root selection**

```go
func TestParseRunArgs_DefaultsThemeToEditorialAI(t *testing.T) {
    opts, err := run.ParseRunArgs([]string{"--mode", "morning"})
    if err != nil { t.Fatal(err) }
    if opts.Theme != "editorial-ai" { t.Fatalf("got %q", opts.Theme) }
}

func TestParseRunArgs_AcceptsThemeFlag(t *testing.T) {
    opts, err := run.ParseRunArgs([]string{"--mode", "morning", "--theme", "ai-product-magazine"})
    if err != nil { t.Fatal(err) }
    if opts.Theme != "ai-product-magazine" { t.Fatalf("got %q", opts.Theme) }
}

func TestRunPipeline_OutputRootIncludesThemeDirectory(t *testing.T) {
    // expect output/editorial-ai/YYYY/MM/DD or output/<theme>/YYYY/MM/DD
}
```

- [ ] **Step 2: Run the run-option tests to verify red**

Run: `go test ./tests/run -run 'Test(ParseRunArgs_(DefaultsThemeToEditorialAI|AcceptsThemeFlag)|RunPipeline_OutputRootIncludesThemeDirectory)$' -v`
Expected: FAIL because `RunOptions` and `DryRunRequest` do not yet carry `Theme`, and output roots are date-only.

- [ ] **Step 3: Implement theme plumbing with a stable default**

```go
type RunOptions struct {
    Date  string
    Mode  string
    Dry   bool
    Theme string
}

type DryRunRequest struct {
    ConfigDir string
    OutputDir string
    StateDir  string
    Date      time.Time
    Mode      string
    Theme     string
}

type DailyEdition struct {
    Date        time.Time
    ThemeID     string
    Keywords    []string
    Featured    []DailyPick
    Learning    []string
    GeneratedAt time.Time
}
```

Implementation notes:
- add `--theme` to `run.ParseRunArgs`
- default to `editorial-ai`
- thread `Theme` from `cmd/daily-builder/main.go` into `run.RunDryPipeline`
- add an output helper that resolves `output/<theme>/YYYY/MM/DD`
- update theme parsing and edition-root resolution in the same change so CLI selection and asset paths cannot drift apart
- keep the rest of the pipeline content semantics unchanged

- [ ] **Step 4: Re-run the run-option tests and the run package**

Run: `go test ./tests/run -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/model/types.go internal/run/options.go internal/run/pipeline.go internal/output/writer.go cmd/daily-builder/main.go tests/run/options_test.go tests/run/pipeline_test.go
git commit -m "feat: add build-time theme selection and theme-scoped output roots"
```

---

### Task 2: Create a theme registry and theme-aware renderer selection

**Files:**
- Create: `internal/render/theme.go`
- Modify: `internal/render/templates.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Create: `web/templates/themes/editorial-ai/index.tmpl`
- Create: `web/templates/themes/editorial-ai/article.tmpl`
- Modify: `tests/render/render_test.go`

- [ ] **Step 1: Write the failing render tests for theme selection**

```go
func TestRenderDailyOutput_RejectsUnknownTheme(t *testing.T) {
    daily := model.DailyEdition{ThemeID: "not-a-theme", Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)}
    if err := render.DailyEdition(t.TempDir(), daily); err == nil {
        t.Fatal("expected unknown theme error")
    }
}

func TestRenderDailyOutput_InjectsThemeIDIntoIndexAndArticlePages(t *testing.T) {
    // expect data-theme-id="editorial-ai" on both generated pages
}
```

- [ ] **Step 2: Run the render tests to verify red**

Run: `go test ./tests/render -run 'TestRenderDailyOutput_(RejectsUnknownTheme|InjectsThemeIDIntoIndexAndArticlePages)$' -v`
Expected: FAIL because the renderer only knows one hard-coded template pair and emits no theme identifier.

- [ ] **Step 3: Implement a small render theme registry**

```go
const DefaultThemeID = "editorial-ai"

type Theme struct {
    ID              string
    IndexTemplate   string
    ArticleTemplate string
}

func ResolveTheme(raw string) (Theme, error) {
    // validate known IDs and return template file paths
}
```

Implementation notes:
- `internal/render/theme.go` should own theme IDs and template-path lookup
- `internal/render/templates.go` should load the theme-specific index/article template pair
- the render data passed to templates should include `ThemeID`
- seed the initial `editorial-ai` template files with the current layout so later tasks can evolve them safely

- [ ] **Step 4: Re-run the focused render tests**

Run: `go test ./tests/render -run 'TestRenderDailyOutput_(RejectsUnknownTheme|InjectsThemeIDIntoIndexAndArticlePages)$' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/theme.go internal/render/templates.go internal/render/index.go internal/render/article.go web/templates/themes/editorial-ai/index.tmpl web/templates/themes/editorial-ai/article.tmpl tests/render/render_test.go
git commit -m "feat: add theme registry and theme-aware template loading"
```

---

### Task 3: Introduce Tailwind build tooling and a shared theme shell

**Files:**
- Create: `package.json`
- Create: `package-lock.json`
- Create: `tailwind.config.js`
- Create: `web/tailwind/input.css`
- Modify: `web/templates/themes/editorial-ai/index.tmpl`
- Modify: `web/templates/themes/editorial-ai/article.tmpl`
- Modify: `web/static/styles.css`
- Modify: `.gitignore`
- Modify: `tests/render/render_test.go`

- [ ] **Step 1: Write the failing render tests for shared shell chrome**

```go
func TestRenderDailyOutput_UsesThemeShellClasses(t *testing.T) {
    // expect exact hooks "theme-shell", "theme-home", and "theme-article"
}

func TestRenderDailyOutput_LinksOnlyLocalStylesheetAssets(t *testing.T) {
    // expect ./assets/styles.css and no external stylesheet URL
}
```

- [ ] **Step 2: Run the render tests to verify red**

Run: `go test ./tests/render -run 'TestRenderDailyOutput_(UsesThemeShellClasses|LinksOnlyLocalStylesheetAssets)$' -v`
Expected: FAIL because the HTML does not yet expose stable shared theme shell hooks.

- [ ] **Step 3: Add Tailwind CLI build support and shared theme CSS entrypoints**

```json
{
  "scripts": {
    "build:styles": "tailwindcss -i ./web/tailwind/input.css -o ./web/static/styles.css --minify"
  }
}
```

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer components {
  .theme-shell { @apply min-h-screen bg-slate-50 text-slate-950 antialiased; }
  .theme-home { @apply mx-auto max-w-7xl px-4 py-6 md:px-6; }
  .theme-article { @apply mx-auto max-w-6xl px-4 py-6 md:px-6; }
}
```

Implementation notes:
- keep `web/static/styles.css` as the compiled local artifact copied into editions
- add `node_modules/` to `.gitignore`
- keep all stylesheet references local; do not use CDN links
- update the editorial theme templates so the page root emits stable shell hooks such as `theme-shell`, `theme-home`, and `theme-article`

- [ ] **Step 4: Install dependencies, build the stylesheet, and rerun render tests**

Run: `npm install -D tailwindcss`
Expected: creates `package-lock.json` and installs Tailwind locally.

Run: `npm run build:styles`
Expected: writes `web/static/styles.css` successfully.

Run: `go test ./tests/render -run 'TestRenderDailyOutput_(UsesThemeShellClasses|LinksOnlyLocalStylesheetAssets)$' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add package.json package-lock.json tailwind.config.js web/tailwind/input.css web/templates/themes/editorial-ai/index.tmpl web/templates/themes/editorial-ai/article.tmpl web/static/styles.css .gitignore tests/render/render_test.go
git commit -m "build: add Tailwind-backed local theme stylesheet pipeline"
```

---

### Task 4: Implement the `editorial-ai` homepage and article layouts

**Files:**
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Modify: `web/templates/themes/editorial-ai/index.tmpl`
- Modify: `web/templates/themes/editorial-ai/article.tmpl`
- Modify: `web/tailwind/input.css`
- Modify: `tests/render/render_test.go`

- [ ] **Step 1: Write the failing layout tests for the default theme**

```go
func TestRenderDailyOutput_EditorialAIHomepageIncludesMastheadLeadAndSecondaryRail(t *testing.T) {
    // expect masthead, lead story region, and secondary rail hooks
}

func TestRenderDailyOutput_EditorialAIArticleKeepsReadingColumnAndSideNotes(t *testing.T) {
    // expect reading-first main column and quieter side-note modules
}
```

- [ ] **Step 2: Run the editorial render tests to verify red**

Run: `go test ./tests/render -run 'TestRenderDailyOutput_EditorialAI(HomepageIncludesMastheadLeadAndSecondaryRail|ArticleKeepsReadingColumnAndSideNotes)$' -v`
Expected: FAIL because the current default templates do not expose the premium editorial layout structure.

- [ ] **Step 3: Implement the default premium theme**

```go
type indexData struct {
    ThemeID        string
    DateLabel      string
    Lead           featuredCardData
    Secondary      []featuredCardData
    Supporting     []featuredCardData
    Learning       []string
}

type articleData struct {
    ThemeID           string
    ReadingColumnHTML template.HTML
    SideNoteSections  []sideNoteSection
}
```

Implementation notes:
- homepage should promote one lead story, a secondary rail, and denser lower sections
- article page should keep title / source / image / summary in the main reading column
- recommendation / profile / feedback / learning should become quieter side notes, not primary chrome
- use restrained gradients and minimal glass effects only where they add polish

- [ ] **Step 4: Rebuild styles and rerun the render package**

Run: `npm run build:styles`
Expected: PASS.

Run: `go test ./tests/render -run 'TestRenderDailyOutput_EditorialAI(HomepageIncludesMastheadLeadAndSecondaryRail|ArticleKeepsReadingColumnAndSideNotes)$' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/index.go internal/render/article.go web/templates/themes/editorial-ai/index.tmpl web/templates/themes/editorial-ai/article.tmpl web/tailwind/input.css tests/render/render_test.go
git commit -m "feat: implement editorial-ai homepage and article layouts"
```

---

### Task 5: Implement the `ai-product-magazine` theme and multi-theme coexistence

**Files:**
- Modify: `internal/render/theme.go`
- Create: `web/templates/themes/ai-product-magazine/index.tmpl`
- Create: `web/templates/themes/ai-product-magazine/article.tmpl`
- Modify: `web/tailwind/input.css`
- Modify: `tests/render/render_test.go`
- Modify: `tests/run/pipeline_test.go`

- [ ] **Step 1: Write the failing tests for the showcase theme and side-by-side output**

```go
func TestRenderDailyOutput_AIProductMagazineIncludesGradientHeroChrome(t *testing.T) {
    // expect stronger gradient/glass hooks and AI-magazine framing
}

func TestRunPipeline_ThemesCanCoexistWithoutOutputClobbering(t *testing.T) {
    // render editorial-ai and ai-product-magazine into the same base output dir
    // expect two distinct output roots
}
```

- [ ] **Step 2: Run the targeted tests to verify red**

Run: `go test ./tests/render ./tests/run -run 'Test(RenderDailyOutput_AIProductMagazineIncludesGradientHeroChrome|RunPipeline_ThemesCanCoexistWithoutOutputClobbering)$' -v`
Expected: FAIL because only the default template exists and theme separation is not yet proven end-to-end.

- [ ] **Step 3: Add the richer AI-native magazine theme**

```go
var themes = map[string]Theme{
    "editorial-ai": {
        ID: "editorial-ai",
        IndexTemplate: "web/templates/themes/editorial-ai/index.tmpl",
        ArticleTemplate: "web/templates/themes/editorial-ai/article.tmpl",
    },
    "ai-product-magazine": {
        ID: "ai-product-magazine",
        IndexTemplate: "web/templates/themes/ai-product-magazine/index.tmpl",
        ArticleTemplate: "web/templates/themes/ai-product-magazine/article.tmpl",
    },
}
```

Implementation notes:
- stronger gradients and glass are allowed, but keep article pages readable
- preserve local CSS only
- keep content semantics identical to the default theme
- ensure two theme runs can share the same base output directory without overwriting each other

- [ ] **Step 4: Rebuild styles and rerun the targeted packages**

Run: `npm run build:styles`
Expected: PASS.

Run: `go test ./tests/render ./tests/run -run 'Test(RenderDailyOutput_AIProductMagazineIncludesGradientHeroChrome|RunPipeline_ThemesCanCoexistWithoutOutputClobbering)$' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/theme.go web/templates/themes/ai-product-magazine/index.tmpl web/templates/themes/ai-product-magazine/article.tmpl web/tailwind/input.css tests/render/render_test.go tests/run/pipeline_test.go
git commit -m "feat: add ai-product-magazine theme and theme coexistence coverage"
```

---

### Task 6: Expose ordered article image candidates and resolve relative URLs

**Files:**
- Modify: `internal/model/types.go`
- Modify: `internal/content/extractor.go`
- Modify: `tests/content/extractor_test.go`

- [ ] **Step 1: Write the failing extraction tests for ordered image candidates**

```go
func TestExtractArticle_PrefersOGImageButRetainsOrderedBodyCandidates(t *testing.T) {
    // expect CoverImage from og:image and body candidates preserved in source order
}

func TestExtractArticle_ResolvesRelativeBodyImageURLsAgainstArticleURL(t *testing.T) {
    // article URL https://example.com/posts/a
    // body image src /images/hero.jpg should become https://example.com/images/hero.jpg
}
```

- [ ] **Step 2: Run the content tests to verify red**

Run: `go test ./tests/content -run 'TestExtractArticle_(PrefersOGImageButRetainsOrderedBodyCandidates|ResolvesRelativeBodyImageURLsAgainstArticleURL)$' -v`
Expected: FAIL because `Article` does not yet expose structured image candidates and body image URLs are not normalized against the article URL.

- [ ] **Step 3: Extend extraction output with structured image metadata**

```go
type ArticleImage struct {
    URL    string
    Source string // "og" or "body"
}

type Article struct {
    CoverImage      string
    ImageCandidates []ArticleImage
    CanonicalURL    string
    Title           string
    ContentText     string
}
```

Implementation notes:
- keep `CoverImage` for the preferred `og:image` if present
- collect body images in source order without applying aggressive validity heuristics here
- resolve relative image URLs against the article canonical URL
- do not invent fallback images in the extractor

- [ ] **Step 4: Re-run the content package**

Run: `go test ./tests/content -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/model/types.go internal/content/extractor.go tests/content/extractor_test.go
git commit -m "feat: expose ordered article image candidates for local archival"
```

---

### Task 7: Archive original images locally and skip invalid body-image candidates

**Files:**
- Modify: `internal/model/types.go`
- Create: `internal/output/images.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/article.go`
- Create: `tests/output/images_test.go`
- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/render/render_test.go`

- [ ] **Step 1: Write the failing tests for invalid-image skipping and local archival**

```go
func TestArchivePreferredImage_SkipsLogoAndPixelCandidates(t *testing.T) {
    candidates := []model.ArticleImage{
        {URL: server.URL + "/logo.svg", Source: "body"},
        {URL: server.URL + "/pixel.gif", Source: "body"},
        {URL: server.URL + "/hero.jpg", Source: "body"},
    }
    // expect assets/images/pick-01-cover.jpg
}

func TestRunPipeline_ArchivesSelectedCoverIntoEditionAssets(t *testing.T) {
    // expect DailyPick.CoverImageLocal to point at assets/images/pick-01-cover.jpg
}

func TestRenderDailyOutput_PrefersEditionLocalCoverPaths(t *testing.T) {
    // expect article page to use ../assets/images/pick-01-cover.jpg and homepage to use ./assets/images/pick-01-cover.jpg
}
```

- [ ] **Step 2: Run the focused output / run / render tests to verify red**

Run: `go test ./tests/output ./tests/run ./tests/render -run 'Test(ArchivePreferredImage_SkipsLogoAndPixelCandidates|RunPipeline_ArchivesSelectedCoverIntoEditionAssets|RenderDailyOutput_PrefersEditionLocalCoverPaths)$' -v`
Expected: FAIL because there is no image archival helper, no invalid-image filtering, and no local cover-path rendering support.

- [ ] **Step 3: Implement conservative image archival**

```go
type DailyPick struct {
    CoverImage      string
    CoverImageLocal string
    SourceURL       string
    Title           string
    Summary         string
}

func ArchivePreferredImage(ctx context.Context, editionRoot, pickID string, cover string, candidates []model.ArticleImage) (string, error) {
    // 1) prefer cover if present
    // 2) otherwise scan candidates in order
    // 3) skip logo/avatar/icon/pixel/tiny-badge heuristics
    // 4) download first valid source asset
    // 5) save assets/images/<pickID>-cover.<ext>
    // 6) return edition-root-relative path
}
```

Implementation notes:
- treat `og:image` and body-image fallback as the same archival pipeline once selected
- prefer explicit heuristics over fuzzy guessing
- if archival fails, preserve graceful rendering; do not fabricate replacement art
- compute page-relative image URLs in render data, not by mutating template strings ad hoc

- [ ] **Step 4: Re-run the focused packages**

Run: `go test ./tests/output ./tests/run ./tests/render -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/model/types.go internal/output/images.go internal/run/pipeline.go internal/render/index.go internal/render/article.go tests/output/images_test.go tests/run/pipeline_test.go tests/render/render_test.go
git commit -m "feat: archive original article images locally with conservative fallback rules"
```

---

### Task 8: Add `youth-signal` and `soft-focus` themes, then document and verify the full matrix

**Files:**
- Modify: `internal/render/theme.go`
- Create: `web/templates/themes/youth-signal/index.tmpl`
- Create: `web/templates/themes/youth-signal/article.tmpl`
- Create: `web/templates/themes/soft-focus/index.tmpl`
- Create: `web/templates/themes/soft-focus/article.tmpl`
- Modify: `web/tailwind/input.css`
- Modify: `tests/render/render_test.go`
- Modify: `README.md`

- [ ] **Step 1: Write the failing render tests for the remaining built-in themes**

```go
func TestRenderDailyOutput_YouthSignalUsesFasterBrighterHomepageChrome(t *testing.T) {
    // expect youth-signal selectors and distinct layout treatment
}

func TestRenderDailyOutput_SoftFocusUsesSofterReadingShell(t *testing.T) {
    // expect soft-focus selectors and gentler article framing
}
```

- [ ] **Step 2: Run the remaining-theme tests to verify red**

Run: `go test ./tests/render -run 'TestRenderDailyOutput_(YouthSignalUsesFasterBrighterHomepageChrome|SoftFocusUsesSofterReadingShell)$' -v`
Expected: FAIL because the final two template pairs do not exist yet.

- [ ] **Step 3: Implement the remaining themes and update usage docs**

```md
# README updates
- npm install
- npm run build:styles
- go run ./cmd/daily-builder run --date 2026-03-20 --mode morning --dry-run --theme editorial-ai
- go run ./cmd/daily-builder run --date 2026-03-20 --mode morning --dry-run --theme ai-product-magazine
```

Implementation notes:
- complete the four-theme matrix in the registry
- keep the same data contract across all templates
- document the supported theme IDs and the local stylesheet build step
- describe the image archival behavior so operators know why original assets appear under `assets/images/`

- [ ] **Step 4: Rebuild styles and run the full verification matrix**

Run: `npm run build:styles`
Expected: PASS.

Run: `go test ./...`
Expected: PASS.

Optional smoke run when RSS connectivity is available:

Run: `go run ./cmd/daily-builder run --date 2026-03-20 --mode morning --dry-run --theme editorial-ai`
Expected: output root under `output/editorial-ai/2026/03/20` with local CSS and article pages.

- [ ] **Step 5: Commit**

```bash
git add internal/render/theme.go web/templates/themes/youth-signal/index.tmpl web/templates/themes/youth-signal/article.tmpl web/templates/themes/soft-focus/index.tmpl web/templates/themes/soft-focus/article.tmpl web/tailwind/input.css tests/render/render_test.go README.md
git commit -m "feat: complete built-in theme set and document template workflow"
```
