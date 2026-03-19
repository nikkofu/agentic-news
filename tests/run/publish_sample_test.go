package run_test

import (
	"testing"

	"github.com/nikkofu/agentic-news/internal/run"
)

func TestParsePublishSampleArgs_DryRun(t *testing.T) {
	opts, err := run.ParsePublishSampleArgs([]string{"--date", "2026-03-19", "--dry-run"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !opts.DryRun {
		t.Fatal("expected dry-run true")
	}
	if opts.Date != "2026-03-19" {
		t.Fatalf("expected date 2026-03-19, got %s", opts.Date)
	}
}
