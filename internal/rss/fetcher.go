package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
)

type rssDoc struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func FetchFeeds(ctx context.Context, urls []string) ([]model.RawItem, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	if len(urls) == 0 {
		return nil, nil
	}

	type feedResult struct {
		index int
		items []model.RawItem
		err   error
	}

	results := make([]feedResult, 0, len(urls))
	resultsCh := make(chan feedResult, len(urls))
	workerLimit := min(8, len(urls))
	sem := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup

	for index, feedURL := range urls {
		wg.Add(1)
		go func(index int, feedURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			items, err := fetchSingleFeed(ctx, client, feedURL)
			resultsCh <- feedResult{
				index: index,
				items: items,
				err:   err,
			}
		}(index, feedURL)
	}

	wg.Wait()
	close(resultsCh)

	for result := range resultsCh {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	out := make([]model.RawItem, 0)
	failCount := 0
	var lastErr error
	for _, result := range results {
		if result.err != nil {
			failCount++
			lastErr = result.err
			continue
		}
		out = append(out, result.items...)
	}

	if len(out) == 0 && failCount > 0 {
		if lastErr == nil {
			lastErr = fmt.Errorf("all feeds failed")
		}
		return nil, lastErr
	}

	return out, nil
}

func fetchSingleFeed(ctx context.Context, client *http.Client, feedURL string) ([]model.RawItem, error) {
	body, err := openFeed(ctx, client, feedURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var doc rssDoc
	if err := xml.NewDecoder(body).Decode(&doc); err != nil {
		return nil, err
	}

	items := make([]model.RawItem, 0, len(doc.Channel.Items))
	for _, item := range doc.Channel.Items {
		published := parsePubDate(item.PubDate)
		items = append(items, model.RawItem{
			SourceID:    feedURL,
			Title:       strings.TrimSpace(item.Title),
			URL:         resolveItemLink(feedURL, strings.TrimSpace(item.Link)),
			PublishedAt: published,
			RawContent:  strings.TrimSpace(item.Description),
		})
	}
	return items, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func openFeed(ctx context.Context, client *http.Client, feedURL string) (io.ReadCloser, error) {
	if isRemoteURL(feedURL) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("fetch %s failed: %s", feedURL, resp.Status)
		}
		return resp.Body, nil
	}

	path, err := localPath(feedURL)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func resolveItemLink(feedURL, itemLink string) string {
	itemLink = strings.TrimSpace(itemLink)
	if itemLink == "" {
		return ""
	}
	if isRemoteURL(itemLink) {
		return itemLink
	}
	if parsed, err := url.Parse(itemLink); err == nil && parsed.Scheme == "file" {
		return itemLink
	}
	if isRemoteURL(feedURL) {
		base, err := url.Parse(feedURL)
		if err != nil {
			return itemLink
		}
		ref, err := url.Parse(itemLink)
		if err != nil {
			return itemLink
		}
		return base.ResolveReference(ref).String()
	}
	path, err := localPath(feedURL)
	if err != nil {
		return itemLink
	}
	if filepath.IsAbs(itemLink) {
		return itemLink
	}
	return filepath.Clean(filepath.Join(filepath.Dir(path), itemLink))
}

func localPath(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err == nil && parsed.Scheme == "file" {
		return parsed.Path, nil
	}
	return raw, nil
}

func isRemoteURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return true
	default:
		return false
	}
}

func parsePubDate(v string) time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC1123Z, v); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC1123, v); err == nil {
		return t
	}
	return time.Time{}
}
