package render_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
	"github.com/nikkofu/agentic-news/internal/render"
)

func TestRenderHugo_EditorialAIHomepageKeepsStableHooks(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}

	packageRoot := writeHugoPackageFixture(t, model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Keywords: []string{
			"AI Agents",
			"Editorial",
		},
		Featured: []model.DailyPick{
			{
				ID:              "lead-1",
				CardType:        "standard",
				Category:        "tech",
				Title:           "Lead Story Headline",
				Summary:         "Lead summary for Hugo homepage hooks.",
				CoverImageLocal: filepath.Join("assets", "images", "lead-1-cover.jpg"),
				SourceName:      "Lead Wire",
				SourceURL:       "https://example.com/lead-1",
				PublishedAt:     time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC),
				TopicTags:       []string{"AI Agents"},
				StyleTags:       []string{"Explainer"},
				CognitiveTags:   []string{"Systems Thinking"},
				Insight: model.Insight{
					Viewpoint:        "Lead viewpoint",
					WhyForYou:        "Lead why for you",
					TasteGrowthHint:  "Lead taste growth hint",
					KnowledgeGapHint: "Lead knowledge gap hint",
				},
			},
			{
				ID:             "brief-1",
				CardType:       "brief",
				Category:       "policy",
				Title:          "Brief Card Headline",
				Summary:        "Brief summary for homepage hook coverage.",
				SourceName:     "Policy Desk",
				SourceURL:      "https://example.com/brief-1",
				PublishedAt:    time.Date(2026, 3, 20, 10, 30, 0, 0, time.UTC),
				FallbackReason: "analysis_failed",
			},
		},
		Learning: []string{"Keep tracking editorial-agent workflows."},
	})
	outputRoot := filepath.Join(t.TempDir(), "editorial-ai", "2026", "03", "20")

	if err := render.RenderHugo(render.HugoRequest{
		PackageRoot: packageRoot,
		OutputRoot:  outputRoot,
		ThemeID:     "editorial-ai",
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexHTML, err := os.ReadFile(filepath.Join(outputRoot, "index.html"))
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}

	html := string(indexHTML)
	for _, required := range []string{
		`./assets/styles.css`,
		`data-theme-id="editorial-ai"`,
		`data-page-kind="index"`,
		`data-layout="editorial-homepage"`,
		`class="container editorial-homepage editorial-ambient"`,
		`class="header editorial-masthead editorial-glass-hero"`,
		`class="card card-lead editorial-lead-story"`,
		`data-article-id="lead-1"`,
		`data-card-type="standard"`,
		`data-article-id="brief-1"`,
		`data-card-type="brief"`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected homepage html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderHugo_EditorialAIArticleKeepsFeedbackAndReadingHooks(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}

	packageRoot := writeHugoPackageFixture(t, model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:              "article-1",
				CardType:        "standard",
				Category:        "tech",
				Title:           "Reading Hook Story",
				Summary:         "Article summary for hook coverage.",
				CoverImageLocal: filepath.Join("assets", "images", "article-1-cover.jpg"),
				SourceName:      "Example",
				SourceURL:       "https://example.com/article-1",
				PublishedAt:     time.Date(2026, 3, 19, 12, 15, 0, 0, time.UTC),
				TopicTags:       []string{"AI Agents"},
				StyleTags:       []string{"Explainer"},
				CognitiveTags:   []string{"Systems Thinking"},
				Insight: model.Insight{
					Viewpoint:        "Article viewpoint",
					WhyForYou:        "Why for you copy",
					TasteGrowthHint:  "Taste growth hint",
					KnowledgeGapHint: "Knowledge gap hint",
				},
			},
		},
		Learning: []string{"Refresh your understanding of feedback loops."},
	})
	outputRoot := filepath.Join(t.TempDir(), "editorial-ai", "2026", "03", "20")

	if err := render.RenderHugo(render.HugoRequest{
		PackageRoot: packageRoot,
		OutputRoot:  outputRoot,
		ThemeID:     "editorial-ai",
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	articleHTML, err := os.ReadFile(filepath.Join(outputRoot, "articles", "article-1.html"))
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	html := string(articleHTML)
	for _, required := range []string{
		`../assets/styles.css`,
		`data-feedback-surface="article"`,
		`class="container editorial-article editorial-ambient"`,
		`class="article-reading-column editorial-reading-glass"`,
		`class="side-note editorial-glass-note"`,
		`data-feedback-value="like"`,
		`data-feedback-value="dislike"`,
		`data-feedback-value="bookmark"`,
		`data-reading-block="title"`,
		`data-reading-block="source"`,
		`data-reading-block="cover-image"`,
		`data-reading-block="summary"`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderHugo_EditorialAIArticleFallsBackToDefaultPersonalizationCopy(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}

	packageRoot := writeHugoPackageFixture(t, model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:              "article-fallback-1",
				CardType:        "standard",
				Category:        "tech",
				Title:           "Fallback Copy Story",
				Summary:         "Article summary for fallback copy coverage.",
				CoverImageLocal: filepath.Join("assets", "images", "article-fallback-1-cover.jpg"),
				SourceName:      "Example",
				SourceURL:       "https://example.com/article-fallback-1",
				PublishedAt:     time.Date(2026, 3, 19, 12, 15, 0, 0, time.UTC),
				Insight: model.Insight{
					Viewpoint: "Article viewpoint",
				},
			},
		},
	})
	outputRoot := filepath.Join(t.TempDir(), "editorial-ai", "2026", "03", "20")

	if err := render.RenderHugo(render.HugoRequest{
		PackageRoot: packageRoot,
		OutputRoot:  outputRoot,
		ThemeID:     "editorial-ai",
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	articleHTML, err := os.ReadFile(filepath.Join(outputRoot, "articles", "article-fallback-1.html"))
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	html := string(articleHTML)
	for _, required := range []string{
		`这篇内容与你近期关注的主题和理解方式相关。`,
		`反馈后会在这里刷新口味拓展建议。`,
		`反馈后会在这里刷新知识补位建议。`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, html)
		}
	}
}

func writeHugoPackageFixture(t *testing.T, daily model.DailyEdition) string {
	t.Helper()

	baseDir := filepath.Join(t.TempDir(), "output")
	editionRoot := t.TempDir()
	for _, item := range daily.Featured {
		if item.CoverImageLocal == "" {
			continue
		}
		path := filepath.Join(editionRoot, item.CoverImageLocal)
		writeTestFile(t, path, "img")
	}

	root, err := output.WriteEditionPackage(baseDir, editionRoot, daily)
	if err != nil {
		t.Fatalf("expected package fixture, got %v", err)
	}
	return root
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
