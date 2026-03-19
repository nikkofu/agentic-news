package run_test

import (
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/config"
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
