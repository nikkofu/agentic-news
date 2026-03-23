package run

import (
	"errors"
	"flag"
	"strings"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
	"github.com/nikkofu/agentic-news/internal/render"
)

type SampleOptions struct {
	Date  string
	Theme string
}

type SampleResult struct {
	OutputRoot  string
	PackageRoot string
}

func ParseSampleArgs(args []string) (SampleOptions, error) {
	flagArgs, positionalDate, err := splitSampleArgs(args)
	if err != nil {
		return SampleOptions{}, err
	}

	fs := flag.NewFlagSet("sample", flag.ContinueOnError)
	var opts SampleOptions
	fs.StringVar(&opts.Date, "date", "", "sample date in YYYY-MM-DD")
	fs.StringVar(&opts.Theme, "theme", model.DefaultThemeID, "render theme ID")
	if err := fs.Parse(flagArgs); err != nil {
		return SampleOptions{}, err
	}

	if opts.Date == "" {
		opts.Date = positionalDate
	}
	if opts.Date == "" {
		opts.Date = "today"
	}
	opts.Theme = model.NormalizeThemeID(opts.Theme)
	return opts, nil
}

func splitSampleArgs(args []string) ([]string, string, error) {
	flagArgs := make([]string, 0, len(args))
	positionalDate := ""

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--date" || arg == "-date" || arg == "--theme" || arg == "-theme":
			flagArgs = append(flagArgs, arg)
			if i+1 >= len(args) {
				return nil, "", errors.New("flag requires value")
			}
			i++
			flagArgs = append(flagArgs, args[i])
		case strings.HasPrefix(arg, "--date=") || strings.HasPrefix(arg, "--theme="):
			flagArgs = append(flagArgs, arg)
		case strings.HasPrefix(arg, "-"):
			flagArgs = append(flagArgs, arg)
		default:
			if positionalDate != "" {
				return nil, "", errors.New("sample accepts at most one positional date")
			}
			positionalDate = arg
		}
	}

	return flagArgs, positionalDate, nil
}

func SampleEditionRoot(outDir string, date time.Time, theme string) string {
	return output.DatePath(outDir, model.DailyEdition{
		Date:    date,
		ThemeID: model.NormalizeThemeID(theme),
	})
}

func GenerateSampleEdition(outDir string, date time.Time, theme string) (SampleResult, error) {
	daily := model.DailyEdition{
		Date:     date,
		ThemeID:  model.NormalizeThemeID(theme),
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
				TopicTags:   []string{"AI Agents", "Enterprise AI"},
				StyleTags:   []string{"Explainer"},
				CognitiveTags: []string{
					"Systems Thinking",
				},
				Insight: model.Insight{
					Viewpoint:        "Near-term narrative strength remains high, but durability depends on deployment economics.",
					WhyForYou:        "Matches your interest in deployment strategy and productization trade-offs.",
					TasteGrowthHint:  "Try a lower-level infrastructure teardown next to pressure-test the top-line narrative.",
					KnowledgeGapHint: "Review inference cost structure and enterprise rollout bottlenecks.",
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
				TopicTags:   []string{"AI Infrastructure", "Capital Allocation"},
				StyleTags:   []string{"Market Analysis"},
				CognitiveTags: []string{
					"Portfolio Construction",
				},
				Insight: model.Insight{
					Viewpoint:        "Monitor valuation stretch versus realized order book growth.",
					WhyForYou:        "Useful if you want to connect product narratives with capital-market confirmation.",
					TasteGrowthHint:  "Pair this with a product-led case study to compare narrative demand and balance-sheet demand.",
					KnowledgeGapHint: "Refresh the difference between capex cycles and recurring software revenue quality.",
				},
			},
		},
		Learning: []string{
			"Compare policy language changes across regions and map to industry impact.",
			"Track whether capital spending guidance is leading or lagging demand data.",
		},
		GeneratedAt: time.Now(),
	}

	editionRoot, err := output.ResetDateDir(outDir, daily)
	if err != nil {
		return SampleResult{}, err
	}
	if err := render.DailyEdition(outDir, daily); err != nil {
		return SampleResult{}, err
	}
	packageRoot, err := output.WriteEditionPackage(outDir, editionRoot, daily)
	if err != nil {
		return SampleResult{}, err
	}
	return SampleResult{
		OutputRoot:  SampleEditionRoot(outDir, date, daily.ThemeID),
		PackageRoot: packageRoot,
	}, nil
}
