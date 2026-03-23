package rss_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/rss"
)

func TestFetchFeeds_ContinuesWhenOneFeedFails(t *testing.T) {
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Good Feed</title>
    <item>
      <title>Good Item</title>
      <link>https://example.com/good</link>
      <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
      <description>Good description</description>
    </item>
  </channel>
</rss>`))
	}))
	defer goodServer.Close()

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusInternalServerError)
	}))
	defer badServer.Close()

	items, err := rss.FetchFeeds(context.Background(), []string{goodServer.URL, badServer.URL})
	if err != nil {
		t.Fatalf("expected no error with partial feed failure, got %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected items from healthy feed")
	}
}

func TestFetchFeeds_ReadsLocalRSSFile(t *testing.T) {
	dir := t.TempDir()
	feedPath := filepath.Join(dir, "feed.xml")
	feed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Local Feed</title>
    <item>
      <title>Local Item</title>
      <link>local-article.html</link>
      <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
      <description>Local description</description>
    </item>
  </channel>
</rss>`
	if err := os.WriteFile(feedPath, []byte(feed), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := rss.FetchFeeds(context.Background(), []string{feedPath})
	if err != nil {
		t.Fatalf("expected local feed file to parse, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Local Item" {
		t.Fatalf("expected local item title, got %q", items[0].Title)
	}
}

func TestFetchFeeds_StartsMultipleRemoteRequestsConcurrently(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 8)

	makeServer := func(title string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started <- struct{}{}
			<-release
			w.Header().Set("Content-Type", "application/rss+xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>` + title + `</title>
    <item>
      <title>` + title + ` item</title>
      <link>https://example.com/` + title + `</link>
      <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
      <description>` + title + ` description</description>
    </item>
  </channel>
</rss>`))
		}))
	}

	serverA := makeServer("alpha")
	defer serverA.Close()
	serverB := makeServer("beta")
	defer serverB.Close()
	serverC := makeServer("gamma")
	defer serverC.Close()

	resultCh := make(chan error, 1)
	go func() {
		_, err := rss.FetchFeeds(context.Background(), []string{serverA.URL, serverB.URL, serverC.URL})
		resultCh <- err
	}()

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected first remote request to start")
	}

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		close(release)
		if err := <-resultCh; err != nil {
			t.Fatalf("expected fetch to complete after release, got %v", err)
		}
		t.Fatal("expected a second remote request to start before the first one finished; fetcher is still effectively sequential")
	}

	close(release)
	if err := <-resultCh; err != nil {
		t.Fatalf("expected concurrent fetch to succeed, got %v", err)
	}
}
