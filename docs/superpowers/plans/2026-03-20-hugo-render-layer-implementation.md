# Hugo Render Layer Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Export a theme-neutral edition package from the current Go pipeline and add a Hugo-based `editorial-ai` renderer that preserves the existing locally-openable edition shape, feedback hooks, and verification contract.

**Architecture:** Keep the current Go pipeline and Go HTML renderer as the source of truth and regression baseline. Add an edition-package writer under `output/_packages/YYYY/MM/DD`, then introduce a dedicated `render-hugo` path that consumes that package and writes the familiar themed output root at `output/<theme>/YYYY/MM/DD`. Preserve the current `data/`, `assets/`, and `articles/` layout in the final Hugo destination so `internal/verify` and `web/static/app.js` do not need a second runtime contract.

**Tech Stack:** Go 1.22+, stdlib, current `internal/run`, `internal/output`, `internal/render`, Hugo CLI, Go templates, `go test`.

**Execution Notes:** Follow @superpowers:test-driven-development task-by-task and @superpowers:verification-before-completion before claiming any task is complete. Real-Hugo integration tests must `t.Skip` when `hugo` is unavailable so `go test ./...` stays green on machines without the binary installed.

---

## File Structure Plan

### Core files to modify

- Modify: `cmd/daily-builder/main.go`
- Modify: `internal/output/writer.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/templates.go`
- Modify: `internal/run/pipeline.go`
- Modify: `internal/run/sample.go`
- Modify: `internal/verify/checks.go`
- Modify: `README.md`

### Core files to create

- Create: `internal/project/root.go`
- Create: `internal/output/package.go`
- Create: `internal/output/package_markdown.go`
- Create: `internal/run/render_hugo.go`
- Create: `internal/render/hugo.go`
- Create: `internal/render/hugo_workspace.go`
- Create: `tests/output/package_test.go`
- Create: `tests/run/render_hugo_test.go`
- Create: `tests/render/hugo_test.go`
- Create: `hugo/config.toml`
- Create: `hugo/themes/editorial-ai/theme.yaml`
- Create: `hugo/themes/editorial-ai/layouts/_default/baseof.html`
- Create: `hugo/themes/editorial-ai/layouts/issues/list.html`
- Create: `hugo/themes/editorial-ai/layouts/issues/single.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/hero.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/card-standard.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/card-brief.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/profile-panel.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/learning-panel.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/feedback-hooks.html`

### Tests to modify

- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/sample/sample_generation_test.go`
- Modify: `tests/verify/checks_test.go`

### Existing files to reference

- Reference: `docs/superpowers/specs/2026-03-20-hugo-render-layer-design.md`
- Reference: `internal/render/article.go`
- Reference: `web/templates/themes/editorial-ai/index.tmpl`
- Reference: `web/templates/themes/editorial-ai/article.tmpl`
- Reference: `web/static/app.js`
- Reference: `web/static/styles.css`

---

### Task 1: Add a theme-neutral edition package writer under `output/_packages`

**Files:**
- Create: `internal/project/root.go`
- Create: `internal/output/package.go`
- Create: `internal/output/package_markdown.go`
- Modify: `internal/output/writer.go`
- Modify: `internal/render/index.go`
- Modify: `internal/render/templates.go`
- Create: `tests/output/package_test.go`

- [ ] **Step 1: Write the failing package-path and package-layout tests**

```go
func TestPackagePath_IsDateScopedAndThemeNeutral(t *testing.T) {
	got := output.PackagePath("output", time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC))
	want := filepath.Join("output", "_packages", "2026", "03", "20")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWriteEditionPackage_WritesHugoContentDataMetaAndStaticAssets(t *testing.T) {
	editionRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(editionRoot, "assets", "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(editionRoot, "assets", "images", "pick-01-cover.jpg"), []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}

	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		Featured: []model.DailyPick{{
			ID:              "pick-01",
			CardType:        "standard",
			Category:        "tech",
			Title:           "Package headline",
			Summary:         "Package summary",
			CoverImageLocal: filepath.Join("assets", "images", "pick-01-cover.jpg"),
			SourceName:      "Example",
			SourceURL:       "https://example.com/pick-01",
			Insight:         model.Insight{Viewpoint: "Package viewpoint"},
		}},
		Learning: []string{"Track follow-up developments"},
	}

	root, err := output.WriteEditionPackage("output", editionRoot, daily)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	required := []string{
		filepath.Join(root, "content", "issues", "_index.md"),
		filepath.Join(root, "content", "issues", "posts", "pick-01.md"),
		filepath.Join(root, "data", "daily.json"),
		filepath.Join(root, "data", "learning.json"),
		filepath.Join(root, "meta", "edition.json"),
		filepath.Join(root, "static", "assets", "styles.css"),
		filepath.Join(root, "static", "assets", "app.js"),
		filepath.Join(root, "static", "assets", "images", "pick-01-cover.jpg"),
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected package file %s: %v", path, err)
		}
	}
}
```

- [ ] **Step 2: Run the output tests to verify red**

Run: `go test ./tests/output -run 'Test(PackagePath_IsDateScopedAndThemeNeutral|WriteEditionPackage_WritesHugoContentDataMetaAndStaticAssets)$' -v`
Expected: FAIL because there is no package helper or package writer yet.

- [ ] **Step 3: Implement the package root helper, Markdown writer, and shared project-root lookup**

```go
func PackagePath(baseDir string, date time.Time) string {
	return filepath.Join(baseDir, "_packages", date.Format("2006"), date.Format("01"), date.Format("02"))
}

func WriteEditionPackage(baseDir, editionRoot string, daily model.DailyEdition) (string, error) {
	root := PackagePath(baseDir, daily.Date)
	// reset package root
	// write content/issues/_index.md
	// write content/issues/posts/<id>.md with front matter url="/articles/<id>.html"
	// write data/daily.json and data/learning.json
	// write meta/edition.json
	// copy web/static/styles.css and web/static/app.js to static/assets/
	// copy archived images referenced by CoverImageLocal into static/assets/images/
	return root, nil
}
```

Implementation notes:
- `internal/project/root.go` should own the reusable `Root()` helper so package writing, Go rendering, and Hugo rendering do not each reimplement repo-root discovery.
- Keep the package theme-neutral: no `theme` directory in the package root and no theme-specific HTML.
- Use `content/issues/_index.md` with front matter `url: "/"` so Hugo can render the homepage to `index.html`.
- Use `content/issues/posts/<pick-id>.md` with front matter `url: "/articles/<pick-id>.html"` so Hugo can preserve the current article output path.
- Derive the Markdown body from the existing `DailyPick` contract only: summary first, then optional viewpoint / why-for-you / learning copy. Do not introduce a dependency on raw extracted article HTML in this bootstrap phase.
- Copy only images already archived under `editionRoot/assets/images/...`; do not re-download images while writing the package.

- [ ] **Step 4: Re-run the output tests**

Run: `go test ./tests/output -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/project/root.go internal/output/package.go internal/output/package_markdown.go internal/output/writer.go internal/render/index.go internal/render/templates.go tests/output/package_test.go
git commit -m "feat: add theme-neutral edition package export"
```

---

### Task 2: Emit the edition package from the current pipeline and sample flow

**Files:**
- Modify: `internal/run/pipeline.go`
- Modify: `internal/run/sample.go`
- Modify: `cmd/daily-builder/main.go`
- Modify: `tests/run/pipeline_test.go`
- Modify: `tests/sample/sample_generation_test.go`

- [ ] **Step 1: Write the failing pipeline and sample tests for package emission**

```go
func TestRunPipeline_WritesEditionPackageAlongsideThemeOutput(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}

	result, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		// existing test hooks
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedPackageRoot := filepath.Join(req.OutputDir, "_packages", "2026", "03", "20")
	if result.PackageRoot != expectedPackageRoot {
		t.Fatalf("expected package root %q, got %q", expectedPackageRoot, result.PackageRoot)
	}
	if _, err := os.Stat(filepath.Join(expectedPackageRoot, "data", "daily.json")); err != nil {
		t.Fatalf("expected package daily.json: %v", err)
	}
}

func TestGenerateSampleEdition_WritesThemeOutputAndEditionPackage(t *testing.T) {
	outDir := t.TempDir()
	result, err := run.GenerateSampleEdition(outDir, time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), "editorial-ai")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.OutputRoot != filepath.Join(outDir, "editorial-ai", "2026", "03", "20") {
		t.Fatalf("unexpected output root %q", result.OutputRoot)
	}
	if result.PackageRoot != filepath.Join(outDir, "_packages", "2026", "03", "20") {
		t.Fatalf("unexpected package root %q", result.PackageRoot)
	}
}
```

- [ ] **Step 2: Run the focused run and sample tests to verify red**

Run: `go test ./tests/run ./tests/sample -run 'Test(RunPipeline_WritesEditionPackageAlongsideThemeOutput|GenerateSampleEdition_WritesThemeOutputAndEditionPackage)$' -v`
Expected: FAIL because `DryRunResult` and sample generation do not expose or write a package root yet.

- [ ] **Step 3: Wire package export into the existing Go-render path**

```go
type DryRunResult struct {
	OutputRoot    string
	PackageRoot   string
	FeaturedCount int
	UsedFallback  bool
}

type SampleResult struct {
	OutputRoot  string
	PackageRoot string
}
```

Implementation notes:
- Keep the current Go renderer as-is; package export is additive in this task.
- In `RunDryPipeline`, write the edition package after `daily` is finalized and before returning the result.
- Keep `verify.DailyEdition` pointed at the themed output root, not the package root.
- Update `sample` command output text to print both the themed edition root and the package root so the next Hugo command has a discoverable input.
- Do not add renderer-selection flags in this task; the incremental bridge is “generate content + package first, render with Hugo second.”

- [ ] **Step 4: Re-run the focused tests and the full run/sample packages**

Run: `go test ./tests/run ./tests/sample -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/run/pipeline.go internal/run/sample.go cmd/daily-builder/main.go tests/run/pipeline_test.go tests/sample/sample_generation_test.go
git commit -m "feat: emit edition packages from run and sample flows"
```

---

### Task 3: Add a dedicated `render-hugo` command and Hugo runner scaffold

**Files:**
- Create: `internal/run/render_hugo.go`
- Create: `internal/render/hugo.go`
- Create: `internal/render/hugo_workspace.go`
- Modify: `cmd/daily-builder/main.go`
- Create: `tests/run/render_hugo_test.go`

- [ ] **Step 1: Write the failing tests for Hugo arg parsing and destination selection**

```go
func TestParseRenderHugoArgs_DefaultsThemeAndDate(t *testing.T) {
	opts, err := run.ParseRenderHugoArgs([]string{"--date", "2026-03-20"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if opts.Theme != "editorial-ai" {
		t.Fatalf("expected default theme editorial-ai, got %q", opts.Theme)
	}
	if opts.Date != "2026-03-20" {
		t.Fatalf("expected date 2026-03-20, got %q", opts.Date)
	}
}

func TestRenderHugo_UsesPackageRootAndThemeScopedDestination(t *testing.T) {
	var gotSource string
	var gotDestination string
	var gotTheme string

	err := render.RenderHugo(render.HugoRequest{
		PackageRoot: filepath.Join("output", "_packages", "2026", "03", "20"),
		OutputRoot:  filepath.Join("output", "editorial-ai", "2026", "03", "20"),
		ThemeID:     "editorial-ai",
		Exec: func(req render.HugoExecRequest) error {
			gotSource = req.Source
			gotDestination = req.Destination
			gotTheme = req.Theme
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotSource != filepath.Join("output", "_packages", "2026", "03", "20") {
		t.Fatalf("unexpected source %q", gotSource)
	}
	if gotDestination != filepath.Join("output", "editorial-ai", "2026", "03", "20") {
		t.Fatalf("unexpected destination %q", gotDestination)
	}
	if gotTheme != "editorial-ai" {
		t.Fatalf("unexpected theme %q", gotTheme)
	}
}
```

- [ ] **Step 2: Run the new Hugo-runner tests to verify red**

Run: `go test ./tests/run -run 'Test(ParseRenderHugoArgs_DefaultsThemeAndDate|RenderHugo_UsesPackageRootAndThemeScopedDestination)$' -v`
Expected: FAIL because there is no `render-hugo` command path or Hugo runner yet.

- [ ] **Step 3: Implement the `render-hugo` command, package lookup, and command runner abstraction**

```go
type RenderHugoOptions struct {
	Date  string
	Theme string
}

type HugoExecRequest struct {
	Source      string
	Destination string
	Theme       string
	ConfigPath  string
	ThemesDir   string
}
```

Implementation notes:
- Add `render-hugo` as a new `cmd/daily-builder` subcommand rather than overloading `run`; this keeps the migration path explicit and low-risk.
- Resolve the package root from the same date using `output.PackagePath("output", date)` and the final themed destination from `output.DatePath("output", model.DailyEdition{Date: date, ThemeID: theme})`.
- Keep the Hugo shell-out behind an injectable function so tests do not need a real binary for command-shape verification.
- Fail fast with a clear error when the package root is missing.
- After Hugo succeeds, copy `data/daily.json`, `data/learning.json`, and `meta/edition.json` from the package into the final destination as `data/daily.json`, `data/learning.json`, and `meta.json` so the existing verifier contract remains intact.
- Use a temporary workspace assembled from the package root plus the repo-owned Hugo config and theme directory; do not mutate the package root in place.

- [ ] **Step 4: Re-run the run-package tests**

Run: `go test ./tests/run -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/run/render_hugo.go internal/render/hugo.go internal/render/hugo_workspace.go cmd/daily-builder/main.go tests/run/render_hugo_test.go
git commit -m "feat: add dedicated Hugo render command scaffold"
```

---

### Task 4: Migrate the `editorial-ai` theme into Hugo while preserving the frontend hook contract

**Files:**
- Create: `hugo/config.toml`
- Create: `hugo/themes/editorial-ai/theme.yaml`
- Create: `hugo/themes/editorial-ai/layouts/_default/baseof.html`
- Create: `hugo/themes/editorial-ai/layouts/issues/list.html`
- Create: `hugo/themes/editorial-ai/layouts/issues/single.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/hero.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/card-standard.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/card-brief.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/profile-panel.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/learning-panel.html`
- Create: `hugo/themes/editorial-ai/layouts/partials/feedback-hooks.html`
- Create: `tests/render/hugo_test.go`

- [ ] **Step 1: Write the failing Hugo integration tests for the `editorial-ai` contract**

```go
func TestRenderHugo_EditorialAIHomepageKeepsStableHooks(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}

	// build a tiny package fixture, run render-hugo, then assert:
	// - ./assets/styles.css
	// - data-theme-id="editorial-ai"
	// - data-page-kind="index"
	// - data-layout="editorial-homepage"
	// - data-article-id + data-card-type on cards
}

func TestRenderHugo_EditorialAIArticleKeepsFeedbackAndReadingHooks(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}

	// assert:
	// - ../assets/styles.css
	// - data-feedback-surface="article"
	// - data-feedback-value="like|dislike|bookmark"
	// - data-reading-block="title|source|cover-image|summary"
}
```

- [ ] **Step 2: Run the Hugo render tests to verify red**

Run: `go test ./tests/render -run 'TestRenderHugo_EditorialAI(HomepageKeepsStableHooks|ArticleKeepsFeedbackAndReadingHooks)$' -v`
Expected: FAIL when `hugo` is installed because the theme files do not exist yet; SKIP on machines without Hugo.

- [ ] **Step 3: Implement the initial Hugo theme and map the package contract into templates**

```toml
# hugo/config.toml
disableKinds = ["taxonomy", "term", "RSS", "sitemap"]
```

Implementation notes:
- Follow the design spec’s theme contract: `baseof`, `issues/list`, `issues/single`, plus partials for hero, standard card, brief card, profile panel, learning panel, and feedback hooks.
- Reproduce the current `editorial-ai` DOM contract from `web/templates/themes/editorial-ai/*.tmpl`, not a visually “similar” approximation.
- Keep the existing local asset references and hooks:
  - homepage: `./assets/styles.css`, `./assets/app.js`
  - articles: `../assets/styles.css`, `../assets/app.js`
  - `data-theme-id`, `data-article-id`, `data-card-type`, `data-page-kind`, `data-layout`, `data-reading-block`, `data-feedback-*`
- Read homepage ordering from `site.Data.daily.Featured` so Hugo does not reorder content by filename or publish date.
- Use the package Markdown/front matter only for article-local content and metadata; do not move ranking or personalization logic into the theme.

- [ ] **Step 4: Re-run the Hugo render tests**

Run: `go test ./tests/render -v`
Expected: PASS on machines with `hugo`; otherwise the Hugo-specific tests should SKIP and the rest of the render suite should still PASS.

- [ ] **Step 5: Commit**

```bash
git add hugo/config.toml hugo/themes/editorial-ai/theme.yaml hugo/themes/editorial-ai/layouts/_default/baseof.html hugo/themes/editorial-ai/layouts/issues/list.html hugo/themes/editorial-ai/layouts/issues/single.html hugo/themes/editorial-ai/layouts/partials/hero.html hugo/themes/editorial-ai/layouts/partials/card-standard.html hugo/themes/editorial-ai/layouts/partials/card-brief.html hugo/themes/editorial-ai/layouts/partials/profile-panel.html hugo/themes/editorial-ai/layouts/partials/learning-panel.html hugo/themes/editorial-ai/layouts/partials/feedback-hooks.html tests/render/hugo_test.go
git commit -m "feat: add editorial-ai Hugo theme baseline"
```

---

### Task 5: Extend verification and document the Hugo bootstrap workflow

**Files:**
- Modify: `internal/verify/checks.go`
- Modify: `tests/verify/checks_test.go`
- Modify: `README.md`

- [ ] **Step 1: Write the failing verification tests for renderer-agnostic output hooks**

```go
func TestVerifyDailyEdition_RequiresStableHomepageHooks(t *testing.T) {
	// create output missing data-theme-id or data-article-id on the homepage
	// expect verify.DailyEdition to return an error
}

func TestVerifyDailyEdition_RequiresArticleFeedbackHooks(t *testing.T) {
	// create output missing data-feedback-surface or feedback buttons
	// expect verify.DailyEdition to return an error
}
```

- [ ] **Step 2: Run the verification tests to verify red**

Run: `go test ./tests/verify -run 'TestVerifyDailyEdition_(RequiresStableHomepageHooks|RequiresArticleFeedbackHooks)$' -v`
Expected: FAIL because the verifier currently checks file existence and JSON thresholds only.

- [ ] **Step 3: Extend verification and update operator docs**

Implementation notes:
- Keep `verify.DailyEdition` renderer-agnostic: it should validate the final output root whether it was produced by Go templates or Hugo.
- In addition to the current file and JSON checks, inspect the homepage and one article page for:
  - `data-page-kind="index"` on the homepage
  - at least one `data-article-id` + `data-card-type` on the homepage
  - `data-feedback-surface="article"` on article pages
  - the three feedback buttons (`like`, `dislike`, `bookmark`)
  - `data-theme-id` on both homepage and article pages
- Update `README.md` with:
  - the new `output/_packages/YYYY/MM/DD` package location
  - `go run ./cmd/daily-builder sample 2026-03-20 --theme editorial-ai`
  - `go run ./cmd/daily-builder render-hugo --date 2026-03-20 --theme editorial-ai`
  - the note that `hugo` is an optional local prerequisite for the second render step

- [ ] **Step 4: Run the full verification matrix**

Run: `go test ./tests/output ./tests/run ./tests/render ./tests/verify -v`
Expected: PASS.

Run: `go test ./...`
Expected: PASS.

Optional smoke run when `hugo` is installed:

Run: `go run ./cmd/daily-builder sample 2026-03-20 --theme editorial-ai`
Expected: Go-rendered edition under `output/editorial-ai/2026/03/20` and package under `output/_packages/2026/03/20`.

Run: `go run ./cmd/daily-builder render-hugo --date 2026-03-20 --theme editorial-ai`
Expected: Hugo-rendered edition under `output/editorial-ai/2026/03/20` with local assets, copied JSON/meta files, and stable feedback hooks.

- [ ] **Step 5: Commit**

```bash
git add internal/verify/checks.go tests/verify/checks_test.go README.md
git commit -m "docs: document Hugo bootstrap workflow and parity checks"
```
