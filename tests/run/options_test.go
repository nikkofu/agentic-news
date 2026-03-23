package run_test

import (
	"testing"

	"github.com/nikkofu/agentic-news/internal/run"
)

func TestParseRunArgs_DefaultsThemeToEditorialAI(t *testing.T) {
	opts, err := run.ParseRunArgs([]string{"--mode", "morning"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Theme != "editorial-ai" {
		t.Fatalf("expected default theme editorial-ai, got %q", opts.Theme)
	}
}

func TestParseRunArgs_AcceptsThemeFlag(t *testing.T) {
	opts, err := run.ParseRunArgs([]string{"--mode", "morning", "--theme", "ai-product-magazine"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Theme != "ai-product-magazine" {
		t.Fatalf("expected theme ai-product-magazine, got %q", opts.Theme)
	}
}

func TestParseRunArgs_AcceptsArbitraryThemeString(t *testing.T) {
	opts, err := run.ParseRunArgs([]string{"--mode", "morning", "--theme", "unknown-theme"})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Theme != "unknown-theme" {
		t.Fatalf("expected theme unknown-theme, got %q", opts.Theme)
	}
}

func TestParseRunArgs_SanitizesThemeToFilesystemSafeID(t *testing.T) {
	opts, err := run.ParseRunArgs([]string{"--mode", "morning", "--theme", "../ A\\B//C.theme "})
	if err != nil {
		t.Fatalf("expected no parse error, got %v", err)
	}
	if opts.Theme != "a-b-c-theme" {
		t.Fatalf("expected sanitized theme a-b-c-theme, got %q", opts.Theme)
	}
}
