package render_test

import (
	"os"
	"path/filepath"
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
				ID:         "a1",
				Category:   "tech",
				Title:      "Test Headline",
				Summary:    "Short summary",
				ScoreFinal: 88.2,
				CoverImage: "https://img.example.com/a.jpg",
				SourceName: "Example",
				SourceURL:  "https://example.com/a",
				PublishedAt: time.Date(2026, 3, 19, 6, 0, 0, 0, time.UTC),
				Insight: model.Insight{Viewpoint: "Insight line"},
			},
		},
		Learning: []string{"Read trend signals"},
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(outDir, "2026", "03", "19", "index.html")
	articlePath := filepath.Join(outDir, "2026", "03", "19", "articles", "a1.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("expected index page at %s", indexPath)
	}
	if _, err := os.Stat(articlePath); err != nil {
		t.Fatalf("expected article page at %s", articlePath)
	}
}
