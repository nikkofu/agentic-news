package sample_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/run"
)

func TestGenerateSampleEdition_WritesDailyOutput(t *testing.T) {
	outDir := t.TempDir()
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	result, err := run.GenerateSampleEdition(outDir, date, model.DefaultThemeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(result.OutputRoot, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("expected index at %s", indexPath)
	}

	dailyJSON := filepath.Join(result.OutputRoot, "data", "daily.json")
	if _, err := os.Stat(dailyJSON); err != nil {
		t.Fatalf("expected daily json at %s", dailyJSON)
	}
}

func TestGenerateSampleEdition_UsesRequestedThemeScopedRoot(t *testing.T) {
	outDir := t.TempDir()
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	result, err := run.GenerateSampleEdition(outDir, date, "ai-product-magazine")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedRoot := run.SampleEditionRoot(outDir, date, "ai-product-magazine")
	if result.OutputRoot != expectedRoot {
		t.Fatalf("expected root %q, got %q", expectedRoot, result.OutputRoot)
	}

	indexPath := filepath.Join(expectedRoot, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("expected themed index at %s", indexPath)
	}
}

func TestGenerateSampleEdition_WritesThemeOutputAndEditionPackage(t *testing.T) {
	outDir := t.TempDir()
	date := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	result, err := run.GenerateSampleEdition(outDir, date, "editorial-ai")
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

func TestGenerateSampleEdition_PopulatesEditorialSideNotesAndTags(t *testing.T) {
	outDir := t.TempDir()
	date := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	result, err := run.GenerateSampleEdition(outDir, date, "editorial-ai")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	articleHTML, err := os.ReadFile(filepath.Join(result.OutputRoot, "articles", "sample-1.html"))
	if err != nil {
		t.Fatalf("expected article html, got %v", err)
	}
	html := string(articleHTML)
	for _, required := range []string{
		`Matches your interest in deployment strategy and productization trade-offs.`,
		`Try a lower-level infrastructure teardown next to pressure-test the top-line narrative.`,
		`Review inference cost structure and enterprise rollout bottlenecks.`,
		`主题：AI Agents`,
		`风格：Explainer`,
		`认知：Systems Thinking`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("expected sample article html to contain %q, got %s", required, html)
		}
	}

	dailyJSON, err := os.ReadFile(filepath.Join(result.OutputRoot, "data", "daily.json"))
	if err != nil {
		t.Fatalf("expected daily json, got %v", err)
	}
	var payload struct {
		Featured []struct {
			ID            string        `json:"ID"`
			TopicTags     []string      `json:"TopicTags"`
			StyleTags     []string      `json:"StyleTags"`
			CognitiveTags []string      `json:"CognitiveTags"`
			Insight       model.Insight `json:"Insight"`
		} `json:"Featured"`
	}
	if err := json.Unmarshal(dailyJSON, &payload); err != nil {
		t.Fatalf("unmarshal daily json: %v", err)
	}
	if len(payload.Featured) == 0 {
		t.Fatal("expected featured items in sample daily json")
	}
	first := payload.Featured[0]
	if first.ID != "sample-1" {
		t.Fatalf("expected first sample item to be sample-1, got %q", first.ID)
	}
	if len(first.TopicTags) == 0 || len(first.StyleTags) == 0 || len(first.CognitiveTags) == 0 {
		t.Fatalf("expected sample item to include editorial tags, got %+v", first)
	}
	if strings.TrimSpace(first.Insight.WhyForYou) == "" || strings.TrimSpace(first.Insight.TasteGrowthHint) == "" || strings.TrimSpace(first.Insight.KnowledgeGapHint) == "" {
		t.Fatalf("expected sample item to include populated side-note copy, got %+v", first.Insight)
	}
}
