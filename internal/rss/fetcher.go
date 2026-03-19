package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
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
	out := make([]model.RawItem, 0)

	for _, feedURL := range urls {
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

		var doc rssDoc
		if err := xml.NewDecoder(resp.Body).Decode(&doc); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
		_ = resp.Body.Close()

		for _, item := range doc.Channel.Items {
			published := parsePubDate(item.PubDate)
			out = append(out, model.RawItem{
				Title:       strings.TrimSpace(item.Title),
				URL:         strings.TrimSpace(item.Link),
				PublishedAt: published,
				RawContent:  strings.TrimSpace(item.Description),
			})
		}
	}

	return out, nil
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
