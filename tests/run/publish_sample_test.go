package run_test

import (
	"path/filepath"
	"testing"
	"time"

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

func TestParsePublishSampleArgs_AcceptsThemeFlag(t *testing.T) {
	opts, err := run.ParsePublishSampleArgs([]string{"--date", "2026-03-19", "--theme", "soft-focus"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if opts.Theme != "soft-focus" {
		t.Fatalf("expected theme soft-focus, got %q", opts.Theme)
	}
}

func TestSampleEditionRoot_UsesThemeScopedDirectory(t *testing.T) {
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	got := run.SampleEditionRoot("output", date, "youth-signal")
	want := filepath.Join("output", "youth-signal", "2026", "03", "19")
	if got != want {
		t.Fatalf("expected sample edition root %q, got %q", want, got)
	}
}
