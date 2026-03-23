package run_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/config"
	"github.com/nikkofu/agentic-news/internal/content"
	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
	"github.com/nikkofu/agentic-news/internal/run"
)

func TestRunPipeline_StartsMultipleExtractionsConcurrently(t *testing.T) {
	release := make(chan struct{})
	started := make(chan string, 8)

	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}

	resultCh := make(chan error, 1)
	go func() {
		_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
			LoadConfig: func(dir string) (config.Config, error) {
				return singleSourceConfig(), nil
			},
			FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
				return testRawItems(3), nil
			},
			Dedupe: func(items []model.RawItem) []model.RawItem { return items },
			Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
				started <- item.Title
				<-release
				return model.Article{
					Title:        item.Title,
					CanonicalURL: item.URL,
					ContentText:  "Enough content to qualify as a standard article for analysis.",
					PublishedAt:  item.PublishedAt,
				}, content.ExtractionStatus{StandardEligible: true}, nil
			},
			AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
				return model.Insight{
					SummaryBrief:       "brief",
					SummaryDeep:        "deep",
					Viewpoint:          "viewpoint",
					Confidence:         80,
					WhyForYou:          "why",
					TasteGrowthHint:    "taste",
					KnowledgeGapHint:   "gap",
					ModelName:          "test",
					ModelVersion:       "v1",
					PromptVersion:      "prompt",
					SourceRefs:         []string{article.CanonicalURL},
					EvidenceSnippets:   []string{"evidence"},
					GeneratedAt:        time.Now(),
				}, nil
			},
			Render: func(baseDir string, daily model.DailyEdition) error { return nil },
			Verify: func(root string) error { return nil },
		})
		resultCh <- err
	}()

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected first extraction to start")
	}

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		close(release)
		if err := <-resultCh; err != nil {
			t.Fatalf("expected pipeline to complete after release, got %v", err)
		}
		t.Fatal("expected a second extraction to start before the first one finished; pipeline is still processing candidates sequentially")
	}

	close(release)
	if err := <-resultCh; err != nil {
		t.Fatalf("expected pipeline to finish successfully, got %v", err)
	}
}

func TestRunPipeline_OnlyProcessesLatestCandidateWindow(t *testing.T) {
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: t.TempDir(),
		Date:      time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}

	var (
		mu        sync.Mutex
		extracted []string
	)

	_, err := run.RunDryPipeline(context.Background(), req, run.DryRunHooks{
		LoadConfig: func(dir string) (config.Config, error) {
			return singleSourceConfig(), nil
		},
		FetchFeeds: func(ctx context.Context, urls []string) ([]model.RawItem, error) {
			return testRawItems(40), nil
		},
		Dedupe: func(items []model.RawItem) []model.RawItem { return items },
		Extract: func(ctx context.Context, item model.RawItem) (model.Article, content.ExtractionStatus, error) {
			mu.Lock()
			extracted = append(extracted, item.ItemID)
			mu.Unlock()
			return model.Article{
				Title:        item.Title,
				CanonicalURL: item.URL,
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return model.Insight{
				SummaryBrief:       "brief",
				SummaryDeep:        "deep",
				Viewpoint:          "viewpoint",
				Confidence:         80,
				WhyForYou:          "why",
				TasteGrowthHint:    "taste",
				KnowledgeGapHint:   "gap",
				ModelName:          "test",
				ModelVersion:       "v1",
				PromptVersion:      "prompt",
				SourceRefs:         []string{article.CanonicalURL},
				EvidenceSnippets:   []string{"evidence"},
				GeneratedAt:        time.Now(),
			}, nil
		},
		Render: func(baseDir string, daily model.DailyEdition) error { return nil },
		Verify: func(root string) error { return nil },
	})
	if err != nil {
		t.Fatalf("expected pipeline to finish successfully, got %v", err)
	}

	sort.Strings(extracted)
	if len(extracted) != 24 {
		t.Fatalf("expected only the latest 24 candidates to be processed, got %d", len(extracted))
	}
	if contains(extracted, "raw-01") || contains(extracted, "raw-16") {
		t.Fatalf("expected oldest candidates to be skipped, got processed ids %v", extracted)
	}
	if !contains(extracted, "raw-17") || !contains(extracted, "raw-40") {
		t.Fatalf("expected newest candidate window to be processed, got ids %v", extracted)
	}
}

func TestRunPipeline_ClearsExistingEditionRootBeforeRebuild(t *testing.T) {
	outputDir := t.TempDir()
	req := run.DryRunRequest{
		ConfigDir: "testdata/config",
		OutputDir: outputDir,
		Date:      time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		Mode:      "morning",
		Theme:     "editorial-ai",
	}

	stalePath := filepath.Join(outputDir, "editorial-ai", "2026", "03", "20", "articles", "stale.html")
	if err := os.MkdirAll(filepath.Dir(stalePath), 0o755); err != nil {
		t.Fatalf("create stale dir: %v", err)
	}
	if err := os.WriteFile(stalePath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

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
				ContentText:  "Enough content to qualify as a standard article for analysis.",
				PublishedAt:  item.PublishedAt,
			}, content.ExtractionStatus{StandardEligible: true}, nil
		},
		AnalyzeInsight: func(ctx context.Context, article model.Article, p profile.UserProfile) (model.Insight, error) {
			return model.Insight{
				SummaryBrief:       "brief",
				SummaryDeep:        "deep",
				Viewpoint:          "viewpoint",
				Confidence:         80,
				WhyForYou:          "why",
				TasteGrowthHint:    "taste",
				KnowledgeGapHint:   "gap",
				ModelName:          "test",
				ModelVersion:       "v1",
				PromptVersion:      "prompt",
				SourceRefs:         []string{article.CanonicalURL},
				EvidenceSnippets:   []string{"evidence"},
				GeneratedAt:        time.Now(),
			}, nil
		},
		Render: func(baseDir string, daily model.DailyEdition) error { return nil },
		Verify: func(root string) error { return nil },
	})
	if err != nil {
		t.Fatalf("expected pipeline to finish successfully, got %v", err)
	}

	_, statErr := os.Stat(stalePath)
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected stale file to be removed before rebuild, stat err=%v", statErr)
	}
}
