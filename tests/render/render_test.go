package render_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/render"
)

func TestRenderDailyOutput_WritesIndexAndArticlePages(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:     time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Keywords: []string{"AI", "Policy"},
		Featured: []model.DailyPick{
			{
				ID:          "a1",
				Category:    "tech",
				Title:       "Test Headline",
				Summary:     "Short summary",
				ScoreFinal:  88.2,
				CoverImage:  "https://img.example.com/a.jpg",
				SourceName:  "Example",
				SourceURL:   "https://example.com/a",
				PublishedAt: time.Date(2026, 3, 19, 6, 0, 0, 0, time.UTC),
				Insight:     model.Insight{Viewpoint: "Insight line"},
			},
		},
		Learning: []string{"Read trend signals"},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "index.html")
	articlePath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "articles", "a1.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("expected index page at %s", indexPath)
	}
	if _, err := os.Stat(articlePath); err != nil {
		t.Fatalf("expected article page at %s", articlePath)
	}
}

func TestRenderDailyOutput_ContainsRequiredFields(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Featured: []model.DailyPick{
			{
				ID:    "fallback-1",
				Title: "Fallback Headline",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "index.html")
	articlePath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "articles", "fallback-1.html")

	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	for _, required := range []string{"未分类", "暂无摘要", "来源待补充", "时间待定", "./articles/fallback-1.html"} {
		if !strings.Contains(string(indexHTML), required) {
			t.Fatalf("expected index html to contain %q, got %s", required, string(indexHTML))
		}
	}

	for _, required := range []string{"未分类", "暂无摘要", "来源待补充", "时间待定", "观点待补充", "href=\"#\""} {
		if !strings.Contains(string(articleHTML), required) {
			t.Fatalf("expected article html to contain %q, got %s", required, string(articleHTML))
		}
	}
}

func TestRenderDailyOutput_PrefersEditionLocalCoverPaths(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:              "pick-01",
				Title:           "Local Cover Headline",
				Summary:         "Short summary",
				CoverImage:      "https://img.example.com/original.jpg",
				CoverImageLocal: filepath.Join("assets", "images", "pick-01-cover.jpg"),
				SourceName:      "Example",
				SourceURL:       "https://example.com/pick-01",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "editorial-ai", "2026", "03", "19")
	indexHTML, err := os.ReadFile(filepath.Join(editionRoot, "index.html"))
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(filepath.Join(editionRoot, "articles", "pick-01.html"))
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	index := string(indexHTML)
	article := string(articleHTML)
	if !strings.Contains(index, `src="./assets/images/pick-01-cover.jpg"`) {
		t.Fatalf("expected homepage to use edition-local cover path, got %s", index)
	}
	if strings.Contains(index, "https://img.example.com/original.jpg") {
		t.Fatalf("expected homepage to prefer local cover path, got %s", index)
	}
	if !strings.Contains(article, `src="../assets/images/pick-01-cover.jpg"`) {
		t.Fatalf("expected article page to use edition-local cover path, got %s", article)
	}
	if strings.Contains(article, "https://img.example.com/original.jpg") {
		t.Fatalf("expected article page to prefer local cover path, got %s", article)
	}
}

func TestRenderDailyOutput_LabelsBriefCardsAndShowsFallbackReason(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Featured: []model.DailyPick{
			{
				ID:             "brief-1",
				CardType:       "brief",
				Category:       "policy",
				Title:          "Fallback Headline",
				Summary:        "RSS summary only",
				SourceName:     "Example",
				SourceURL:      "https://example.com/fallback",
				FallbackReason: "analysis_failed",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "index.html")
	articlePath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "articles", "brief-1.html")

	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	if !strings.Contains(string(indexHTML), "简版") {
		t.Fatalf("expected index html to label brief card, got %s", string(indexHTML))
	}
	if !strings.Contains(string(articleHTML), "analysis_failed") {
		t.Fatalf("expected article html to show fallback reason, got %s", string(articleHTML))
	}
	if strings.Contains(string(articleHTML), "AI点评") {
		t.Fatalf("expected brief article html to omit AI点评 section, got %s", string(articleHTML))
	}
}

func TestRenderDailyOutput_ArticlePageIncludesFeedbackControlsAndWhyForYou(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Featured: []model.DailyPick{
			{
				ID:            "a1",
				CardType:      "standard",
				Category:      "tech",
				Title:         "Personalized Headline",
				Summary:       "Short summary",
				ScoreFinal:    91.5,
				SourceName:    "Example",
				SourceURL:     "https://example.com/a1",
				PublishedAt:   time.Date(2026, 3, 19, 9, 30, 0, 0, time.UTC),
				TopicTags:     []string{"AI Agents", "Startups"},
				StyleTags:     []string{"Explainer"},
				CognitiveTags: []string{"Systems Thinking"},
				Insight: model.Insight{
					Viewpoint:        "Insight line",
					WhyForYou:        "Matches your interest in AI product strategy.",
					TasteGrowthHint:  "Try a more technical market analysis next.",
					KnowledgeGapHint: "Brush up on inference cost trade-offs.",
				},
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	articlePath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "articles", "a1.html")
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	html := string(articleHTML)
	for _, required := range []string{
		`data-feedback-surface="article"`,
		`data-page-kind="article"`,
		`data-article-id="a1"`,
		`data-feedback-value="like"`,
		`data-feedback-value="dislike"`,
		`data-feedback-value="bookmark"`,
		`class="reason-tags"`,
		`Matches your interest in AI product strategy.`,
		`class="profile-panel"`,
		`data-topic-tags=`,
		`data-style-tags=`,
		`data-cognitive-tags=`,
		`Try a more technical market analysis next.`,
		`Brush up on inference cost trade-offs.`,
		`<script defer src="../assets/app.js"></script>`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderDailyOutput_IndexPageIncludesTrackingHooks(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Featured: []model.DailyPick{
			{
				ID:            "track-1",
				Title:         "Tracked Headline",
				Category:      "business",
				SourceName:    "Newswire",
				SourceURL:     "https://example.com/track-1",
				TopicTags:     []string{"Economy"},
				StyleTags:     []string{"Analysis"},
				CognitiveTags: []string{"Macro"},
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "index.html")
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}

	html := string(indexHTML)
	for _, required := range []string{
		`data-page-kind="index"`,
		`data-track-event="article_click"`,
		`data-article-id="track-1"`,
		`data-card-type="standard"`,
		`data-source-url="https://example.com/track-1"`,
		`data-topic-tags=`,
		`data-style-tags=`,
		`data-cognitive-tags=`,
		`href="./articles/track-1.html"`,
		`<script defer src="./assets/app.js"></script>`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected index html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderDailyOutput_EditorialAIHomepageIncludesMastheadLeadAndSecondaryRail(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:          "lead-1",
				CardType:    "standard",
				Title:       "Lead Story Headline",
				ScoreFinal:  93.2,
				SourceURL:   "https://example.com/lead-1",
				SourceName:  "Lead Wire",
				PublishedAt: time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC),
			},
			{
				ID:          "rail-1",
				CardType:    "brief",
				Title:       "Secondary Rail Story",
				ScoreFinal:  81.4,
				SourceURL:   "https://example.com/rail-1",
				SourceName:  "Rail Desk",
				PublishedAt: time.Date(2026, 3, 20, 9, 15, 0, 0, time.UTC),
			},
			{
				ID:          "support-1",
				Title:       "Supporting Story",
				ScoreFinal:  77.6,
				SourceURL:   "https://example.com/support-1",
				SourceName:  "Support Daily",
				PublishedAt: time.Date(2026, 3, 20, 11, 45, 0, 0, time.UTC),
			},
			{
				ID:          "support-2",
				Title:       "Supporting Story 2",
				ScoreFinal:  74.1,
				SourceURL:   "https://example.com/support-2",
				SourceName:  "Support Daily",
				PublishedAt: time.Date(2026, 3, 20, 12, 30, 0, 0, time.UTC),
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "editorial-ai", "2026", "03", "20", "index.html")
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}

	html := string(indexHTML)
	for _, required := range []string{
		`data-layout="editorial-homepage"`,
		`class="container editorial-homepage editorial-ambient"`,
		`class="header editorial-masthead editorial-glass-hero"`,
		`class="card card-lead editorial-lead-story"`,
		`data-home-region="masthead"`,
		`data-home-region="lead-story"`,
		`data-home-region="secondary-rail"`,
		`data-home-region="supporting-grid"`,
		`data-feature-role="lead"`,
		`data-feature-role="secondary"`,
		`data-feature-role="supporting"`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected index html to contain %q, got %s", required, html)
		}
	}

	for _, required := range []string{
		`https://example.com/rail-1`,
		`评分：81.4`,
		`2026-03-20 09:15 UTC`,
		`https://example.com/support-1`,
		`评分：77.6`,
		`2026-03-20 11:45 UTC`,
		`https://example.com/support-2`,
		`评分：74.1`,
		`2026-03-20 12:30 UTC`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected index html to preserve secondary/supporting semantics %q, got %s", required, html)
		}
	}
}

func TestRenderDailyOutput_EditorialAIArticleKeepsReadingColumnAndSideNotes(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:         "article-layout-1",
				CardType:   "standard",
				Category:   "tech",
				Title:      "Reading-First Layout Headline",
				Summary:    "Main-column summary.",
				CoverImage: "https://img.example.com/story.jpg",
				SourceName: "Layout Times",
				SourceURL:  "https://example.com/story",
				Insight: model.Insight{
					Viewpoint:        "Insight line",
					WhyForYou:        "Recommended for your systems interest.",
					TasteGrowthHint:  "Try longform market reports next.",
					KnowledgeGapHint: "Revisit retrieval-augmented generation basics.",
				},
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	articlePath := filepath.Join(outDir, "editorial-ai", "2026", "03", "20", "articles", "article-layout-1.html")
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	html := string(articleHTML)
	for _, required := range []string{
		`data-layout="editorial-article"`,
		`class="container editorial-article editorial-ambient"`,
		`class="article-reading-column editorial-reading-glass"`,
		`class="side-note editorial-glass-note"`,
		`data-article-region="reading-column"`,
		`data-article-region="side-notes"`,
		`data-reading-block="title"`,
		`data-reading-block="source"`,
		`data-reading-block="cover-image"`,
		`data-reading-block="summary"`,
		`data-side-note="recommendation"`,
		`data-side-note="profile"`,
		`data-side-note="feedback"`,
		`data-side-note="learning"`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderDailyOutput_AIProductMagazineIncludesGradientHeroChrome(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "ai-product-magazine",
		Keywords: []string{
			"AI Product",
			"Showcase",
		},
		Featured: []model.DailyPick{
			{
				ID:            "mag-hero-1",
				CardType:      "standard",
				Category:      "tech",
				Title:         "Gradient Hero Story",
				Summary:       "Showcase-first but still readable.",
				ScoreFinal:    95.4,
				SourceName:    "Demo Desk",
				SourceURL:     "https://example.com/mag-hero-1",
				PublishedAt:   time.Date(2026, 3, 20, 8, 30, 0, 0, time.UTC),
				TopicTags:     []string{"AI Product"},
				StyleTags:     []string{"Showcase"},
				CognitiveTags: []string{"Strategy"},
			},
		},
		Learning: []string{"Keep reading fidelity while adding visual chrome."},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "ai-product-magazine", "2026", "03", "20", "index.html")
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}

	html := string(indexHTML)
	for _, required := range []string{
		`data-theme-id="ai-product-magazine"`,
		`data-layout="ai-product-magazine-homepage"`,
		`data-home-region="hero-chrome"`,
		`data-home-region="showcase-grid"`,
		`data-home-accent="gradient"`,
		`data-home-chrome="glass"`,
		`agentic-news · ai-product-magazine`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected index html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderDailyOutput_AIProductMagazineArticleKeepsReadingAndInteractionHooks(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "ai-product-magazine",
		Featured: []model.DailyPick{
			{
				ID:            "ai-mag-article-1",
				CardType:      "standard",
				Category:      "tech",
				Title:         "AI Product Story",
				Summary:       "Readable article with showcase chrome.",
				ScoreFinal:    92.8,
				SourceName:    "Showcase Desk",
				SourceURL:     "https://example.com/ai-mag-article-1",
				PublishedAt:   time.Date(2026, 3, 20, 10, 15, 0, 0, time.UTC),
				TopicTags:     []string{"AI Product"},
				StyleTags:     []string{"Showcase"},
				CognitiveTags: []string{"Systems Thinking"},
				Insight: model.Insight{
					Viewpoint:        "Focus on product packaging and durable UX value.",
					WhyForYou:        "Matches your AI product execution interests.",
					TasteGrowthHint:  "Compare this with a deeper infra-focused analysis.",
					KnowledgeGapHint: "Review model serving cost trade-offs.",
				},
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	articlePath := filepath.Join(outDir, "ai-product-magazine", "2026", "03", "20", "articles", "ai-mag-article-1.html")
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	html := string(articleHTML)
	for _, required := range []string{
		`data-layout="ai-product-magazine-article"`,
		`data-article-region="reading-column"`,
		`data-article-region="side-notes"`,
		`data-side-note="recommendation"`,
		`data-side-note="profile"`,
		`data-side-note="feedback"`,
		`data-side-note="learning"`,
		`data-feedback-surface="article"`,
		`data-feedback-value="like"`,
		`data-feedback-value="dislike"`,
		`data-feedback-value="bookmark"`,
		`data-profile-panel`,
		`data-profile-summary`,
		`<link rel="stylesheet" href="../assets/styles.css" />`,
		`<script defer src="../assets/app.js"></script>`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, html)
		}
	}
}

func TestRenderDailyOutput_YouthSignalUsesFasterBrighterHomepageChrome(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "youth-signal",
		Keywords: []string{
			"Signals",
			"Momentum",
		},
		Featured: []model.DailyPick{
			{
				ID:            "signal-lead-1",
				CardType:      "standard",
				Category:      "tech",
				Title:         "Signal Surge Story",
				Summary:       "Fast scan, high-contrast framing.",
				ScoreFinal:    94.1,
				SourceName:    "Signal Desk",
				SourceURL:     "https://example.com/signal-lead-1",
				PublishedAt:   time.Date(2026, 3, 20, 7, 45, 0, 0, time.UTC),
				TopicTags:     []string{"Signals"},
				StyleTags:     []string{"Fast Scan"},
				CognitiveTags: []string{"Pattern Spotting"},
			},
			{
				ID:          "signal-rail-1",
				CardType:    "brief",
				Category:    "business",
				Title:       "Momentum Rail Story",
				Summary:     "Secondary card still readable.",
				ScoreFinal:  82.6,
				SourceName:  "Momentum Wire",
				SourceURL:   "https://example.com/signal-rail-1",
				PublishedAt: time.Date(2026, 3, 20, 8, 10, 0, 0, time.UTC),
			},
		},
		Learning: []string{"Separate urgency cues from durable importance."},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "youth-signal", "2026", "03", "20")
	indexHTML, err := os.ReadFile(filepath.Join(editionRoot, "index.html"))
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(filepath.Join(editionRoot, "articles", "signal-lead-1.html"))
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	index := string(indexHTML)
	for _, required := range []string{
		`data-theme-id="youth-signal"`,
		`data-layout="youth-signal-homepage"`,
		`data-home-region="signal-banner"`,
		`data-home-region="speed-lane"`,
		`data-home-region="context-grid"`,
		`data-home-tone="high-contrast"`,
		`data-label-style="expressive"`,
		`data-feature-role="lead"`,
		`data-feature-role="secondary"`,
		`agentic-news · youth-signal`,
	} {
		if !strings.Contains(index, required) {
			t.Fatalf("expected index html to contain %q, got %s", required, index)
		}
	}

	article := string(articleHTML)
	for _, required := range []string{
		`data-layout="youth-signal-article"`,
		`data-article-region="reading-column"`,
		`data-article-region="side-notes"`,
		`data-reading-shell="signal-briefing"`,
	} {
		if !strings.Contains(article, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, article)
		}
	}
}

func TestRenderDailyOutput_SoftFocusUsesSofterReadingShell(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "soft-focus",
		Keywords: []string{
			"Reflection",
			"Calm Reading",
		},
		Featured: []model.DailyPick{
			{
				ID:            "soft-lead-1",
				CardType:      "standard",
				Category:      "culture",
				Title:         "Gentle Briefing Story",
				Summary:       "Reading-first layout with relaxed framing.",
				ScoreFinal:    89.3,
				SourceName:    "Soft Desk",
				SourceURL:     "https://example.com/soft-lead-1",
				PublishedAt:   time.Date(2026, 3, 20, 9, 5, 0, 0, time.UTC),
				TopicTags:     []string{"Reflection"},
				StyleTags:     []string{"Essay"},
				CognitiveTags: []string{"Sensemaking"},
				Insight: model.Insight{
					Viewpoint:        "Let the framing stay calm while preserving structure.",
					WhyForYou:        "Matches your preference for reflective analysis.",
					TasteGrowthHint:  "Try pairing this with a sharper opposing viewpoint.",
					KnowledgeGapHint: "Review the longer historical arc behind this topic.",
				},
			},
		},
		Learning: []string{"Use softer pacing without losing information hierarchy."},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "soft-focus", "2026", "03", "20")
	indexHTML, err := os.ReadFile(filepath.Join(editionRoot, "index.html"))
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(filepath.Join(editionRoot, "articles", "soft-lead-1.html"))
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	index := string(indexHTML)
	for _, required := range []string{
		`data-theme-id="soft-focus"`,
		`data-layout="soft-focus-homepage"`,
		`data-home-region="soft-hero"`,
		`data-home-region="calm-stack"`,
		`data-home-tone="gentle"`,
		`data-surface-style="soft"`,
	} {
		if !strings.Contains(index, required) {
			t.Fatalf("expected index html to contain %q, got %s", required, index)
		}
	}

	article := string(articleHTML)
	for _, required := range []string{
		`data-layout="soft-focus-article"`,
		`data-article-region="reading-column"`,
		`data-article-region="side-notes"`,
		`data-reading-shell="soft-focus"`,
		`data-article-tone="gentle"`,
		`data-side-note="recommendation"`,
		`data-side-note="profile"`,
		`data-side-note="feedback"`,
		`data-side-note="learning"`,
		`Matches your preference for reflective analysis.`,
	} {
		if !strings.Contains(article, required) {
			t.Fatalf("expected article html to contain %q, got %s", required, article)
		}
	}
}

func TestRenderDailyOutput_UsesConsistentEditionDateAcrossIndexAndArticle(t *testing.T) {
	outDir := t.TempDir()
	location := time.FixedZone("UTC-7", -7*60*60)
	daily := model.DailyEdition{
		Date: time.Date(2026, 3, 19, 23, 30, 0, 0, location),
		Featured: []model.DailyPick{
			{
				ID:    "date-1",
				Title: "Edition Date Check",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "index.html")
	articlePath := filepath.Join(outDir, "editorial-ai", "2026", "03", "19", "articles", "date-1.html")

	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	expected := `data-edition-date="2026-03-19"`
	if !strings.Contains(string(indexHTML), expected) {
		t.Fatalf("expected index html to contain %q, got %s", expected, string(indexHTML))
	}
	if !strings.Contains(string(articleHTML), expected) {
		t.Fatalf("expected article html to contain %q, got %s", expected, string(articleHTML))
	}
}

func TestRenderDailyOutput_RejectsUnknownTheme(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "not-a-theme",
	}

	err := render.DailyEdition(outDir, daily)
	if err == nil {
		t.Fatal("expected unknown theme error")
	}
	if !strings.Contains(err.Error(), "unknown theme") {
		t.Fatalf("expected unknown theme error, got %v", err)
	}
}

func TestRenderDailyOutput_InjectsThemeIDIntoIndexAndArticlePages(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:    "theme-1",
				Title: "Theme Injection Check",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "editorial-ai", "2026", "03", "20")
	indexPath := filepath.Join(editionRoot, "index.html")
	articlePath := filepath.Join(editionRoot, "articles", "theme-1.html")

	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	expected := `data-theme-id="editorial-ai"`
	if !strings.Contains(string(indexHTML), expected) {
		t.Fatalf("expected index html to contain %q, got %s", expected, string(indexHTML))
	}
	if !strings.Contains(string(articleHTML), expected) {
		t.Fatalf("expected article html to contain %q, got %s", expected, string(articleHTML))
	}
}

func TestRenderDailyOutput_UsesThemeShellClasses(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:    "shell-1",
				Title: "Theme Shell Check",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "editorial-ai", "2026", "03", "20")
	indexPath := filepath.Join(editionRoot, "index.html")
	articlePath := filepath.Join(editionRoot, "articles", "shell-1.html")

	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	for _, required := range []string{"theme-shell", "theme-home"} {
		if !strings.Contains(string(indexHTML), required) {
			t.Fatalf("expected index html to contain %q, got %s", required, string(indexHTML))
		}
	}
	for _, required := range []string{"theme-shell", "theme-article"} {
		if !strings.Contains(string(articleHTML), required) {
			t.Fatalf("expected article html to contain %q, got %s", required, string(articleHTML))
		}
	}
}

func TestRenderDailyOutput_UsesLocalOnlyStylesheetLinks(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:    "css-1",
				Title: "Stylesheet Locality Check",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "editorial-ai", "2026", "03", "20")
	indexPath := filepath.Join(editionRoot, "index.html")
	articlePath := filepath.Join(editionRoot, "articles", "css-1.html")

	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}
	articleHTML, err := os.ReadFile(articlePath)
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}

	index := string(indexHTML)
	article := string(articleHTML)
	if !strings.Contains(index, `<link rel="stylesheet" href="./assets/styles.css" />`) {
		t.Fatalf("expected index html to use local stylesheet, got %s", index)
	}
	if !strings.Contains(index, `<script defer src="./assets/app.js"></script>`) {
		t.Fatalf("expected index html to use local script, got %s", index)
	}
	if regexp.MustCompile(`<link[^>]+rel="stylesheet"[^>]+href="https?://`).MatchString(index) {
		t.Fatalf("expected index html stylesheet link to be local, got %s", index)
	}
	if regexp.MustCompile(`<script[^>]+src="https?://`).MatchString(index) {
		t.Fatalf("expected index html script src to be local, got %s", index)
	}
	if !strings.Contains(article, `<link rel="stylesheet" href="../assets/styles.css" />`) {
		t.Fatalf("expected article html to use local stylesheet, got %s", article)
	}
	if !strings.Contains(article, `<script defer src="../assets/app.js"></script>`) {
		t.Fatalf("expected article html to use local script, got %s", article)
	}
	if regexp.MustCompile(`<link[^>]+rel="stylesheet"[^>]+href="https?://`).MatchString(article) {
		t.Fatalf("expected article html stylesheet link to be local, got %s", article)
	}
	if regexp.MustCompile(`<script[^>]+src="https?://`).MatchString(article) {
		t.Fatalf("expected article html script src to be local, got %s", article)
	}
}

func TestRenderDailyOutput_WorksWhenCWDOutsideRepo(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected current working directory, got %v", err)
	}
	tmpWD := t.TempDir()
	if err := os.Chdir(tmpWD); err != nil {
		t.Fatalf("expected to change cwd, got %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:    "cwd-1",
				Title: "CWD Independence Check",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "editorial-ai", "2026", "03", "20")
	for _, requiredPath := range []string{
		filepath.Join(editionRoot, "index.html"),
		filepath.Join(editionRoot, "articles", "cwd-1.html"),
		filepath.Join(editionRoot, "assets", "styles.css"),
		filepath.Join(editionRoot, "assets", "app.js"),
	} {
		if _, err := os.Stat(requiredPath); err != nil {
			t.Fatalf("expected generated file at %s, got %v", requiredPath, err)
		}
	}
}

func TestRenderDailyOutput_EmptyIDHandlingIsConsistent(t *testing.T) {
	outDir := t.TempDir()
	daily := model.DailyEdition{
		Date:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		ThemeID: "editorial-ai",
		Featured: []model.DailyPick{
			{
				ID:    "",
				Title: "Missing ID Item",
			},
			{
				ID:    "kept-1",
				Title: "Renderable Item",
			},
		},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	editionRoot := filepath.Join(outDir, "editorial-ai", "2026", "03", "20")
	indexPath := filepath.Join(editionRoot, "index.html")
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index html, got %v", err)
	}

	html := string(indexHTML)
	if !strings.Contains(html, `href="./articles/kept-1.html"`) {
		t.Fatalf("expected index html to contain kept article link, got %s", html)
	}
	if strings.Contains(html, `href="./articles/untitled.html"`) {
		t.Fatalf("expected index html to omit untitled article link, got %s", html)
	}
	if strings.Contains(html, "Missing ID Item") {
		t.Fatalf("expected index html to omit missing-id item, got %s", html)
	}

	articleDir := filepath.Join(editionRoot, "articles")
	entries, err := os.ReadDir(articleDir)
	if err != nil {
		t.Fatalf("expected article directory, got %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "kept-1.html" {
		t.Fatalf("expected only kept-1 article page, got %v", entries)
	}
}
