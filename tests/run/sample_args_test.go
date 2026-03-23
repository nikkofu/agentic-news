package run_test

import (
	"testing"

	"github.com/nikkofu/agentic-news/internal/run"
)

func TestParseSampleArgs_DefaultsThemeToEditorialAI(t *testing.T) {
	opts, err := run.ParseSampleArgs([]string{"2026-03-19"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Date != "2026-03-19" {
		t.Fatalf("expected date 2026-03-19, got %q", opts.Date)
	}
	if opts.Theme != "editorial-ai" {
		t.Fatalf("expected default theme editorial-ai, got %q", opts.Theme)
	}
}

func TestParseSampleArgs_AcceptsThemeAfterPositionalDate(t *testing.T) {
	opts, err := run.ParseSampleArgs([]string{"2026-03-19", "--theme", "ai-product-magazine"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Theme != "ai-product-magazine" {
		t.Fatalf("expected theme ai-product-magazine, got %q", opts.Theme)
	}
}
