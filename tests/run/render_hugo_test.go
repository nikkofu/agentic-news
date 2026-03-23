package run_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
	"github.com/nikkofu/agentic-news/internal/render"
	"github.com/nikkofu/agentic-news/internal/run"
)

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
	baseDir := t.TempDir()
	packageRoot := filepath.Join(baseDir, "output", "_packages", "2026", "03", "20")
	outputRoot := filepath.Join(baseDir, "output", "editorial-ai", "2026", "03", "20")
	writeTestFile(t, filepath.Join(packageRoot, "data", "daily.json"), `{"date":"2026-03-20"}`)
	writeTestFile(t, filepath.Join(packageRoot, "data", "learning.json"), `{"today":[]}`)
	writeTestFile(t, filepath.Join(packageRoot, "meta", "edition.json"), `{"theme":"editorial-ai"}`)

	var gotSource string
	var gotDestination string
	var gotTheme string

	err := render.RenderHugo(render.HugoRequest{
		PackageRoot: packageRoot,
		OutputRoot:  outputRoot,
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
	if gotSource != packageRoot {
		t.Fatalf("unexpected source %q", gotSource)
	}
	if gotDestination != outputRoot {
		t.Fatalf("unexpected destination %q", gotDestination)
	}
	if gotTheme != "editorial-ai" {
		t.Fatalf("unexpected theme %q", gotTheme)
	}
}

func TestRenderHugoEdition_RelativeBaseDirWritesHTMLToFinalOutputRoot(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}

	workspaceDir, err := os.MkdirTemp(".", "render-hugo-rel-")
	if err != nil {
		t.Fatalf("mkdir temp workspace: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(workspaceDir)
	})

	baseDir, err := filepath.Rel(".", workspaceDir)
	if err != nil {
		t.Fatalf("relative workspace path: %v", err)
	}

	editionRoot := t.TempDir()
	writeTestFile(t, filepath.Join(editionRoot, "assets", "images", "sample-1-cover.jpg"), "img")
	date := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	if _, err := output.WriteEditionPackage(baseDir, editionRoot, model.DailyEdition{
		Date:    date,
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{{
			ID:              "sample-1",
			CardType:        "standard",
			Category:        "tech",
			Title:           "Relative Hugo Render",
			Summary:         "Checks relative destination handling.",
			CoverImageLocal: filepath.Join("assets", "images", "sample-1-cover.jpg"),
			SourceName:      "Example",
			SourceURL:       "https://example.com/sample-1",
			PublishedAt:     time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC),
			Insight: model.Insight{
				Viewpoint:        "Viewpoint",
				WhyForYou:        "Why for you",
				TasteGrowthHint:  "Taste growth hint",
				KnowledgeGapHint: "Knowledge gap hint",
			},
		}},
		Learning: []string{"Learning hint"},
	}); err != nil {
		t.Fatalf("write package: %v", err)
	}

	result, err := run.RenderHugoEdition(baseDir, date, "editorial-ai")
	if err != nil {
		t.Fatalf("render hugo edition: %v", err)
	}

	if _, err := os.Stat(filepath.Join(result.OutputRoot, "index.html")); err != nil {
		t.Fatalf("expected index.html under final output root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputRoot, "articles", "sample-1.html")); err != nil {
		t.Fatalf("expected article html under final output root: %v", err)
	}

	wrongPath := filepath.Join(output.PackagePath(baseDir, date), result.OutputRoot, "index.html")
	if _, err := os.Stat(wrongPath); err == nil {
		t.Fatalf("expected no nested output under package root, found %s", wrongPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat nested output: %v", err)
	}
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
