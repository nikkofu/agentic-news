package run

import (
	"path/filepath"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/render"
)

func GenerateSampleEdition(outDir string, date time.Time) (string, error) {
	daily := model.DailyEdition{
		Date:     date,
		Keywords: []string{"AI", "Policy", "Finance"},
		Featured: []model.DailyPick{
			{
				ID:          "sample-1",
				Category:    "tech",
				Title:       "Sample: Frontier model rollout accelerates",
				Summary:     "Major labs accelerate deployment with stronger enterprise adoption signals.",
				ScoreFinal:  87.4,
				CoverImage:  "https://images.unsplash.com/photo-1518773553398-650c184e0bb3",
				SourceName:  "Example Tech Source",
				SourceURL:   "https://example.com/tech/frontier-models",
				PublishedAt: date.Add(6 * time.Hour),
				Insight: model.Insight{
					Viewpoint: "Near-term narrative strength remains high, but durability depends on deployment economics.",
				},
			},
			{
				ID:          "sample-2",
				Category:    "finance",
				Title:       "Sample: Capital rotation toward AI infra",
				Summary:     "Infrastructure-linked names outperformed amid sustained compute demand expectations.",
				ScoreFinal:  84.1,
				CoverImage:  "https://images.unsplash.com/photo-1460925895917-afdab827c52f",
				SourceName:  "Example Finance Source",
				SourceURL:   "https://example.com/finance/ai-infra",
				PublishedAt: date.Add(5 * time.Hour),
				Insight: model.Insight{
					Viewpoint: "Monitor valuation stretch versus realized order book growth.",
				},
			},
		},
		Learning: []string{
			"Compare policy language changes across regions and map to industry impact.",
			"Track whether capital spending guidance is leading or lagging demand data.",
		},
		GeneratedAt: time.Now(),
	}

	if err := render.DailyEdition(outDir, daily); err != nil {
		return "", err
	}
	y, m, d := date.Date()
	return filepath.Join(outDir, formatY(y), formatM(m), formatD(d)), nil
}

func formatY(y int) string { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") }
func formatM(m time.Month) string { return time.Date(2000, m, 1, 0, 0, 0, 0, time.UTC).Format("01") }
func formatD(d int) string { return time.Date(2000, 1, d, 0, 0, 0, 0, time.UTC).Format("02") }
