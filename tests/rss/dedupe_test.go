package rss_test

import (
	"testing"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/rss"
)

func TestDedupe_RemovesCanonicalDuplicates(t *testing.T) {
	items := []model.RawItem{
		{Title: "One", URL: "https://example.com/a"},
		{Title: "One duplicate", URL: "https://example.com/a"},
		{Title: "Two", URL: "https://example.com/b"},
	}

	got := rss.Dedupe(items)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique items, got %d", len(got))
	}
}

func TestDedupe_UsesTitleHashFallback(t *testing.T) {
	items := []model.RawItem{
		{Title: "Same Title", URL: ""},
		{Title: " same   title ", URL: ""},
	}

	got := rss.Dedupe(items)
	if len(got) != 1 {
		t.Fatalf("expected 1 unique item by title hash, got %d", len(got))
	}
}
