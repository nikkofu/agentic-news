package rss_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikkofu/agentic-news/internal/rss"
)

func TestFetchFeeds_ParsesItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Example Feed</title>
    <item>
      <title>Item One</title>
      <link>https://example.com/a</link>
      <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
      <description>Example description</description>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	items, err := rss.FetchFeeds(context.Background(), []string{server.URL})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected items")
	}
	if items[0].Title != "Item One" {
		t.Fatalf("expected title Item One, got %s", items[0].Title)
	}
}
