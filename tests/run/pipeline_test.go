package run_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/analyze"
	"github.com/nikkofu/agentic-news/internal/config"
	"github.com/nikkofu/agentic-news/internal/content"
	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
	"github.com/nikkofu/agentic-news/internal/run"
)

func TestLoadConfig_ValidMinimalConfig(t *testing.T) {
	cfg, err := config.LoadConfig("testdata/config")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AI.QualityMode != "high" {
		t.Fatalf("expected high quality mode")
	}
	if len(cfg.RSS.Sources) == 0 {
		t.Fatalf("expected rss sources to be loaded")
	}
}

func TestShouldFallbackForDeadlineGuard(t *testing.T) {
	now := time.Date(2026, 3, 19, 6, 51, 0, 0, time.Local)
	if !run.ShouldFallback(now) {
		t.Fatal("expected fallback near deadline")
	}
}

func TestCLI_AcceptsMorningMode(t *testing.T) {
	opts, err := run.ParseRunArgs([]string{"--date", "today", "--mode", "morning"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Mode != "morning" {
		t.Fatalf("expected mode morning, got %s", opts.Mode)
	}
}

func TestRunPipeline_OutputRootIncludesThemeDirectory(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "ai-product-magazine",
	}

	var verifyRoot string
	result, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return singleSourceConfig(), nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
		Render: func(baseDir string, daily model.DailyEdition) error {
			if daily.ThemeID != req.Theme {
				t.Fatalf("expected daily edition theme %q, got %q", req.Theme, daily.ThemeID)
			}
			return nil
		},
		Verify: func(root string) error {
			verifyRoot = root
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedRoot := filepath.Join(req.OutputDir, req.Theme, "2026", "03", "19")
	if result.OutputRoot != expectedRoot {
		t.Fatalf("expected output root %q, got %q", expectedRoot, result.OutputRoot)
	}
	if verifyRoot != expectedRoot {
		t.Fatalf("expected verify root %q, got %q", expectedRoot, verifyRoot)
	}
}

func TestRunPipeline_WritesEditionPackageAlongsideThemeOutput(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}

	result, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return singleSourceConfig(), nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
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

func TestRunPipeline_SanitizesDirtyThemeIntoEditionAndOutputRoot(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "../ Editorial/AI.theme",
	}

	var rendered model.DailyEdition
	var verifyRoot string
	result, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return singleSourceConfig(), nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
		Render: func(baseDir string, daily model.DailyEdition) error {
			rendered = daily
			return nil
		},
		Verify: func(root string) error {
			verifyRoot = root
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedTheme := "editorial-ai-theme"
	if rendered.ThemeID != expectedTheme {
		t.Fatalf("expected rendered theme %q, got %q", expectedTheme, rendered.ThemeID)
	}

	expectedRoot := filepath.Join(req.OutputDir, expectedTheme, "2026", "03", "19")
	if result.OutputRoot != expectedRoot {
		t.Fatalf("expected output root %q, got %q", expectedRoot, result.OutputRoot)
	}
	if verifyRoot != expectedRoot {
		t.Fatalf("expected verify root %q, got %q", expectedRoot, verifyRoot)
	}
}

func TestRunPipeline_ArchivesSelectedCoverIntoEditionAssets(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}

	var archivedEditionRoot string
	var archivedPickID string
	var archivedCover string
	var archivedCandidates []model.ArticleImage
	var rendered model.DailyEdition

	_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return singleSourceConfig(), nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				CoverImage:   "https://img.example.com/original-cover.jpg",
				ImageCandidates: []model.ArticleImage{
					{URL: "https://img.example.com/logo.svg", Source: "body"},
					{URL: "https://img.example.com/hero.jpg", Source: "body"},
				},
				ContentText: "Enough content to qualify as a standard article for analysis.",
				PublishedAt: item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
		ArchiveImage: func(ctx context.Context, editionRoot, pickID string, cover string, candidates []model.ArticleImage) (string, error) {
			archivedEditionRoot = editionRoot
			archivedPickID = pickID
			archivedCover = cover
			archivedCandidates = append([]model.ArticleImage(nil), candidates...)
			localPath := filepath.Join("assets", "images", pickID+"-cover.jpg")
			absolutePath := filepath.Join(editionRoot, localPath)
			if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(absolutePath, []byte("archived-image"), 0o644); err != nil {
				return "", err
			}
			return localPath, nil
		},
		Render: func(baseDir string, daily model.DailyEdition) error {
			rendered = daily
			return nil
		},
		Verify: func(root string) error { return nil },
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if archivedEditionRoot != filepath.Join(req.OutputDir, "editorial-ai", "2026", "03", "19") {
		t.Fatalf("expected archive edition root to be edition directory, got %q", archivedEditionRoot)
	}
	if archivedPickID == "" {
		t.Fatal("expected archive helper to receive a pick id")
	}
	if archivedCover != "https://img.example.com/original-cover.jpg" {
		t.Fatalf("expected explicit cover to be passed through, got %q", archivedCover)
	}
	if len(archivedCandidates) != 2 {
		t.Fatalf("expected archive helper to receive image candidates, got %d", len(archivedCandidates))
	}

	var matched *model.DailyPick
	for index := range rendered.Featured {
		if rendered.Featured[index].ID == archivedPickID {
			matched = &rendered.Featured[index]
			break
		}
	}
	if matched == nil {
		t.Fatalf("expected archived pick %q to appear in rendered featured items", archivedPickID)
	}
	if got := matched.CoverImageLocal; got != filepath.Join("assets", "images", archivedPickID+"-cover.jpg") {
		t.Fatalf("expected local archived cover path %q, got %q", filepath.Join("assets", "images", archivedPickID+"-cover.jpg"), got)
	}
}

func TestRunPipeline_ThemesCanCoexistWithoutOutputClobbering(t *testing.T) {
	outputDir := t.TempDir()
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	runForTheme := func(t *testing.T, themeID string) (string, string) {
		t.Helper()
		req := run.DryRunRequest{
			ConfigDir: "testdata/config",
			OutputDir: outputDir,
			Date:      date,
			Mode:      "morning",
			Theme:     themeID,
		}

		result, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(10), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
			AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
				return analyze.RunPipeline(ctx, article, p)
			},
		})
		if err != nil {
			t.Fatalf("expected no error for theme %q, got %v", themeID, err)
		}

		indexPath := filepath.Join(result.OutputRoot, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			t.Fatalf("expected index page at %s, got %v", indexPath, err)
		}
		indexHTML, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("expected index html for theme %q, got %v", themeID, err)
		}
		return result.OutputRoot, string(indexHTML)
	}

	editorialRoot, editorialHTML := runForTheme(t, "editorial-ai")
	magazineRoot, magazineHTML := runForTheme(t, "ai-product-magazine")

	if editorialRoot == magazineRoot {
		t.Fatalf("expected distinct output roots, got same root %q", editorialRoot)
	}

	for _, required := range []string{
		`data-theme-id="editorial-ai"`,
		`data-layout="editorial-homepage"`,
	} {
		if !strings.Contains(editorialHTML, required) {
			t.Fatalf("expected editorial index html to contain %q, got %s", required, editorialHTML)
		}
	}
	for _, required := range []string{
		`data-theme-id="ai-product-magazine"`,
		`data-layout="ai-product-magazine-homepage"`,
	} {
		if !strings.Contains(magazineHTML, required) {
			t.Fatalf("expected magazine index html to contain %q, got %s", required, magazineHTML)
		}
	}
}

func TestRunPipeline_DryRunExecutesCoreStages(t *testing.T) {
	var stages []string
	var seenExtract bool
	var seenAnalyze bool
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
	}

	result, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			stages = append(stages, "config")
			return config.Config{
				AI: config.AIConfig{QualityMode: "high"},
				RSS: config.RSSConfig{Sources: []config.RSSSource{
					{SourceID: "s1", RSSURL: "feed://one", Name: "Source One"},
				}},
			}, nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			stages = append(stages, "fetch")
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem {
			stages = append(stages, "dedupe")
			return items
		},
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			if !seenExtract {
				stages = append(stages, "extract")
				seenExtract = true
			}
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			if !seenAnalyze {
				stages = append(stages, "analyze")
				seenAnalyze = true
			}
			return analyze.RunPipeline(ctx, article, p)
		},
		Rank: func(items []model.DailyPick) []model.DailyPick {
			stages = append(stages, "rank")
			return items
		},
		Render: func(baseDir string, daily model.DailyEdition) error {
			stages = append(stages, "render")
			return nil
		},
		Verify: func(root string) error {
			stages = append(stages, "verify")
			return nil
		},
		Now: func() time.Time {
			return time.Date(2026, 3, 19, 6, 51, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{"config", "fetch", "dedupe", "extract", "analyze", "rank", "render", "verify"}
	if len(stages) != len(want) {
		t.Fatalf("expected stages %v, got %v", want, stages)
	}
	for i := range want {
		if stages[i] != want[i] {
			t.Fatalf("expected stage %d to be %s, got %s", i, want[i], stages[i])
		}
	}
	if !result.UsedFallback {
		t.Fatal("expected fallback near deadline")
	}
	if result.FeaturedCount != 10 {
		t.Fatalf("expected 10 featured items, got %d", result.FeaturedCount)
	}
}

func TestRunPipeline_ProducesPublishableMixedEditionFromRealRSS(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
	}

	var rendered model.DailyEdition
	_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return config.Config{
				AI: config.AIConfig{QualityMode: "high"},
				RSS: config.RSSConfig{Sources: []config.RSSSource{
					{SourceID: "s1", RSSURL: "feed://one", Name: "Source One"},
				}},
			}, nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(12), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem {
			return items
		},
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			if item.ItemID == "raw-01" || item.ItemID == "raw-02" || item.ItemID == "raw-03" || item.ItemID == "raw-04" {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough extracted article content to be analyzed as a standard card.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			}
			return model.Article{}, content.ExtractionStatus{
				StandardEligible: false,
				UsedFallbackText: true,
				FallbackReason:   "extract_failed",
			}, fmt.Errorf("extract failed")
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
		Render: func(baseDir string, daily model.DailyEdition) error {
			rendered = daily
			return nil
		},
		Verify: func(root string) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rendered.Featured) != 10 {
		t.Fatalf("expected 10 featured items, got %d", len(rendered.Featured))
	}

	standardCount := 0
	briefCount := 0
	for _, item := range rendered.Featured {
		switch item.CardType {
		case "standard":
			standardCount++
		case "brief":
			briefCount++
		}
	}
	if standardCount < 3 {
		t.Fatalf("expected at least 3 standard cards, got %d", standardCount)
	}
	if briefCount == 0 {
		t.Fatal("expected at least one brief card in mixed edition")
	}
}

func TestRunPipeline_FailsWhenPublishabilityThresholdCannotBeMet(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
	}

	_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return config.Config{
				AI: config.AIConfig{QualityMode: "high"},
				RSS: config.RSSConfig{Sources: []config.RSSSource{
					{SourceID: "s1", RSSURL: "feed://one", Name: "Source One"},
				}},
			}, nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem {
			return items
		},
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			if item.ItemID == "raw-01" || item.ItemID == "raw-02" {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough extracted article content to be analyzed as a standard card.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			}
			return model.Article{}, content.ExtractionStatus{
				StandardEligible: false,
				UsedFallbackText: true,
				FallbackReason:   "extract_failed",
			}, fmt.Errorf("extract failed")
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
	})
	if err == nil {
		t.Fatal("expected publishability failure")
	}
}

func TestRunPipeline_LoadsProfileSnapshotAndFallsBackWhenMissing(t *testing.T) {
	baseReq := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		StateDir:  "/tmp/custom-state",
	}

	t.Run("defaults state dir when unset", func(t *testing.T) {
		req := baseReq
		req.StateDir = ""
		var loadedStateDir string

		_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			LoadProfile: func(stateDir string) (profile.UserProfile, error) {
				loadedStateDir = stateDir
				return profile.UserProfile{
					FocusTopics: []string{"technology"},
				}, nil
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(10), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
			AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
				return analyze.RunPipeline(ctx, article, p)
			},
			Render: func(baseDir string, daily model.DailyEdition) error { return nil },
			Verify: func(root string) error { return nil },
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if loadedStateDir != "state" {
			t.Fatalf("expected default state dir %q, got %q", "state", loadedStateDir)
		}
	})

	t.Run("loads saved snapshot", func(t *testing.T) {
		var loadedStateDir string
		var seenProfile profile.UserProfile
		var rendered model.DailyEdition

		saved := profile.UserProfile{
			FocusTopics:          []string{"semiconductors"},
			PreferredStyles:      []string{"data-driven"},
			CognitivePreferences: []string{"systems thinking"},
			TopicAffinity: map[string]float64{
				"semiconductors": 10,
			},
			StyleAffinity: map[string]float64{
				"data-driven": 8,
			},
			CognitiveAffinity: map[string]float64{
				"systems thinking": 7,
			},
			SourceAffinity: map[string]float64{
				"Source One": 5,
			},
		}

		_, err := run.RunDryPipeline(context.Background(), baseReq, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			LoadProfile: func(stateDir string) (profile.UserProfile, error) {
				loadedStateDir = stateDir
				return saved, nil
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(10), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
			AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
				seenProfile = p
				return analyze.RunPipeline(ctx, article, p)
			},
			Render: func(baseDir string, daily model.DailyEdition) error {
				rendered = daily
				return nil
			},
			Verify: func(root string) error { return nil },
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if loadedStateDir != baseReq.StateDir {
			t.Fatalf("expected state dir %q, got %q", baseReq.StateDir, loadedStateDir)
		}
		if !reflect.DeepEqual(seenProfile.FocusTopics, saved.FocusTopics) {
			t.Fatalf("expected analyze profile to use saved snapshot, got %#v", seenProfile)
		}
		if len(rendered.Featured) == 0 {
			t.Fatal("expected rendered edition to contain featured items")
		}
		if got := rendered.Featured[0].TopicTags; len(got) == 0 || got[0] != "semiconductors" {
			t.Fatalf("expected deterministic topic tags from profile, got %#v", got)
		}
		if got := rendered.Featured[0].StyleTags; len(got) == 0 || got[0] != "data-driven" {
			t.Fatalf("expected deterministic style tags from profile, got %#v", got)
		}
		if got := rendered.Featured[0].CognitiveTags; len(got) == 0 || got[0] != "systems thinking" {
			t.Fatalf("expected deterministic cognitive tags from profile, got %#v", got)
		}
	})

	t.Run("falls back when snapshot missing", func(t *testing.T) {
		var seenProfile profile.UserProfile
		var rendered model.DailyEdition

		_, err := run.RunDryPipeline(context.Background(), baseReq, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			LoadProfile: func(stateDir string) (profile.UserProfile, error) {
				return profile.UserProfile{}, fs.ErrNotExist
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(10), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
			AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
				seenProfile = p
				return analyze.RunPipeline(ctx, article, p)
			},
			Render: func(baseDir string, daily model.DailyEdition) error {
				rendered = daily
				return nil
			},
			Verify: func(root string) error { return nil },
		})
		if err != nil {
			t.Fatalf("expected no error on missing snapshot, got %v", err)
		}
		if len(seenProfile.FocusTopics) != 0 || len(seenProfile.PreferredStyles) != 0 || len(seenProfile.CognitivePreferences) != 0 {
			t.Fatalf("expected safe empty profile for analysis, got %#v", seenProfile)
		}
		insight := firstStandardInsight(t, rendered)
		if !strings.Contains(insight.WhyForYou, "broad strategic relevance") {
			t.Fatalf("expected safe generic why_for_you, got %q", insight.WhyForYou)
		}
		if strings.TrimSpace(insight.TasteGrowthHint) == "" {
			t.Fatal("expected non-empty safe taste growth hint")
		}
		if strings.TrimSpace(insight.KnowledgeGapHint) == "" {
			t.Fatal("expected non-empty safe knowledge gap hint")
		}
	})

	t.Run("falls back when snapshot effectively empty", func(t *testing.T) {
		var seenProfile profile.UserProfile
		var rendered model.DailyEdition

		_, err := run.RunDryPipeline(context.Background(), baseReq, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			LoadProfile: func(stateDir string) (profile.UserProfile, error) {
				return profile.UserProfile{}, nil
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(10), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
			AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
				seenProfile = p
				return analyze.RunPipeline(ctx, article, p)
			},
			Render: func(baseDir string, daily model.DailyEdition) error {
				rendered = daily
				return nil
			},
			Verify: func(root string) error { return nil },
		})
		if err != nil {
			t.Fatalf("expected no error on empty snapshot, got %v", err)
		}
		if len(seenProfile.FocusTopics) != 0 || len(seenProfile.PreferredStyles) != 0 || len(seenProfile.CognitivePreferences) != 0 {
			t.Fatalf("expected safe empty profile for analysis, got %#v", seenProfile)
		}
		insight := firstStandardInsight(t, rendered)
		if !strings.Contains(insight.WhyForYou, "broad strategic relevance") {
			t.Fatalf("expected safe generic why_for_you, got %q", insight.WhyForYou)
		}
		if strings.TrimSpace(insight.TasteGrowthHint) == "" {
			t.Fatal("expected non-empty safe taste growth hint")
		}
		if strings.TrimSpace(insight.KnowledgeGapHint) == "" {
			t.Fatal("expected non-empty safe knowledge gap hint")
		}
	})

	t.Run("surfaces real load errors", func(t *testing.T) {
		_, err := run.RunDryPipeline(context.Background(), baseReq, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			LoadProfile: func(stateDir string) (profile.UserProfile, error) {
				return profile.UserProfile{}, errors.New("disk offline")
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(10), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
		})
		if err == nil {
			t.Fatal("expected load profile error")
		}
	})
}

func TestRunPipeline_UsesPerItemMetadataTagsWhenAvailable(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		StateDir:  "/tmp/custom-state",
	}

	var rendered model.DailyEdition
	_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return singleSourceConfig(), nil
		},
		LoadProfile: func(stateDir string) (profile.UserProfile, error) {
			return profile.UserProfile{
				FocusTopics:          []string{"semiconductors"},
				PreferredStyles:      []string{"data-driven"},
				CognitivePreferences: []string{"systems thinking"},
			}, nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(10), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			article := model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}
			switch item.ItemID {
			case "raw-01":
				article.Keywords = []string{"semiconductor capex"}
				article.CategorySecondary = []string{"analysis", "systems"}
			case "raw-02":
				article.Keywords = []string{"industrial policy"}
				article.CategorySecondary = []string{"briefing", "policy reasoning"}
			}
			return article, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
		Rank: func(items []model.DailyPick) []model.DailyPick { return items },
		Render: func(baseDir string, daily model.DailyEdition) error {
			rendered = daily
			return nil
		},
		Verify: func(root string) error { return nil },
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	itemByID := map[string]model.DailyPick{}
	for _, item := range rendered.Featured {
		itemByID[item.ID] = item
	}

	first := itemByID["raw-01"]
	second := itemByID["raw-02"]
	third := itemByID["raw-03"]

	if !contains(first.TopicTags, "semiconductor capex") {
		t.Fatalf("expected raw-01 topic tags to use article metadata, got %#v", first.TopicTags)
	}
	if !contains(second.TopicTags, "industrial policy") {
		t.Fatalf("expected raw-02 topic tags to use article metadata, got %#v", second.TopicTags)
	}
	if !contains(first.StyleTags, "analysis") || !contains(second.StyleTags, "briefing") {
		t.Fatalf("expected per-item style tags from metadata, got raw-01=%#v raw-02=%#v", first.StyleTags, second.StyleTags)
	}
	if !contains(first.CognitiveTags, "systems") || !contains(second.CognitiveTags, "policy reasoning") {
		t.Fatalf("expected per-item cognitive tags from metadata, got raw-01=%#v raw-02=%#v", first.CognitiveTags, second.CognitiveTags)
	}
	if !contains(third.TopicTags, "semiconductors") {
		t.Fatalf("expected fallback profile topic tags when metadata missing, got %#v", third.TopicTags)
	}
}

func TestRunPipeline_FreshnessUsesInjectedNowHook(t *testing.T) {
	fixedNow := time.Date(2020, 1, 2, 10, 0, 0, 0, time.UTC)
	publishedAt := fixedNow.Add(-2 * time.Hour)
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		StateDir:  "/tmp/custom-state",
	}

	var rendered model.DailyEdition
	_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			cfg := singleSourceConfig()
			cfg.Scoring.Weights = config.ScoreWeights{
				Importance:        0,
				PersonalRelevance: 0,
				Credibility:       0,
				Novelty:           0,
				Freshness:         1,
			}
			return cfg, nil
		},
		LoadProfile: func(stateDir string) (profile.UserProfile, error) {
			return profile.UserProfile{}, fs.ErrNotExist
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return []model.RawItem{
				{
					ItemID:      "raw-01",
					SourceID:    "s1",
					Title:       "Freshness candidate",
					URL:         "https://example.com/freshness",
					RawContent:  "RSS summary",
					PublishedAt: publishedAt,
				},
			}, nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return analyze.RunPipeline(ctx, article, p)
		},
		Rank: func(items []model.DailyPick) []model.DailyPick { return items },
		Render: func(baseDir string, daily model.DailyEdition) error {
			rendered = daily
			return nil
		},
		Verify: func(root string) error { return nil },
		Now: func() time.Time {
			return fixedNow
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rendered.Featured) != 1 {
		t.Fatalf("expected 1 featured item, got %d", len(rendered.Featured))
	}

	got := rendered.Featured[0].ScoreFinal
	want := 91.4
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("expected score %.1f from injected now freshness, got %.4f", want, got)
	}
}

func singleSourceConfig() config.Config {
	return config.Config{
		AI: config.AIConfig{QualityMode: "high"},
		RSS: config.RSSConfig{Sources: []config.RSSSource{
			{SourceID: "s1", RSSURL: "feed://one", Name: "Source One"},
		}},
	}
}

func firstStandardInsight(t *testing.T, daily model.DailyEdition) model.Insight {
	t.Helper()
	for _, item := range daily.Featured {
		if item.CardType == "standard" {
			return item.Insight
		}
	}
	t.Fatal("expected at least one standard insight")
	return model.Insight{}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func testRawItems(n int) []model.RawItem {
	items := make([]model.RawItem, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, model.RawItem{
			ItemID:      fmt.Sprintf("raw-%02d", i+1),
			SourceID:    "s1",
			Title:       fmt.Sprintf("Raw item %02d", i+1),
			URL:         fmt.Sprintf("https://example.com/%02d", i+1),
			RawContent:  fmt.Sprintf("RSS summary %02d", i+1),
			PublishedAt: time.Date(2026, 3, 19, 8, i, 0, 0, time.UTC),
		})
	}
	return items
}
