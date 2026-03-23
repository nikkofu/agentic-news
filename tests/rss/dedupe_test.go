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


func TestDedupe_NormalizesQueryAndTrailingSlash(t *testing.T) {
	items := []model.RawItem{
		{Title: "A", URL: "https://example.com/news/a/?utm_source=x"},
		{Title: "A dup", URL: "https://example.com/news/a"},
	}

	got := rss.Dedupe(items)
	if len(got) != 1 {
		t.Fatalf("expected URL variants to dedupe into 1 item, got %d", len(got))
	}
}

func TestDedupe_PreservesMeaningfulQueryParameters(t *testing.T) {
	items := []model.RawItem{
		{Title: "A", URL: "https://example.com/news?id=1"},
		{Title: "B", URL: "https://example.com/news?id=2"},
	}

	got := rss.Dedupe(items)
	if len(got) != 2 {
		t.Fatalf("expected meaningful query params to remain distinct, got %d", len(got))
	}
}
