package sample_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/run"
)

func TestGenerateSampleEdition_WritesDailyOutput(t *testing.T) {
	outDir := t.TempDir()
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	root, err := run.GenerateSampleEdition(outDir, date)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	indexPath := filepath.Join(root, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("expected index at %s", indexPath)
	}

	dailyJSON := filepath.Join(root, "data", "daily.json")
	if _, err := os.Stat(dailyJSON); err != nil {
		t.Fatalf("expected daily json at %s", dailyJSON)
	}
}
