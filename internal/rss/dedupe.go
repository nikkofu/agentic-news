package rss

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/nikkofu/agentic-news/internal/model"
)

func Dedupe(items []model.RawItem) []model.RawItem {
	seen := make(map[string]struct{}, len(items))
	out := make([]model.RawItem, 0, len(items))

	for _, item := range items {
		key := canonicalURL(item.URL)
		if key == "" {
			key = "title:" + titleHash(item.Title)
		}

		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}

	return out
}

func canonicalURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return raw
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = strings.TrimRight(parsed.RawPath, "/")

	query := parsed.Query()
	filtered := url.Values{}
	for key, values := range query {
		if isTrackingQueryKey(key) {
			continue
		}
		for _, value := range values {
			filtered.Add(key, value)
		}
	}
	parsed.RawQuery = filtered.Encode()

	return parsed.String()
}

func isTrackingQueryKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if strings.HasPrefix(key, "utm_") {
		return true
	}

	switch key {
	case "fbclid", "gclid", "mc_cid", "mc_eid", "ref", "ref_src":
		return true
	default:
		return false
	}
}

func titleHash(title string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(title)), " ")
	h := sha1.Sum([]byte(normalized))
	return hex.EncodeToString(h[:])
}
