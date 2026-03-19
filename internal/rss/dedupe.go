package rss

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"github.com/nikkofu/agentic-news/internal/model"
)

func Dedupe(items []model.RawItem) []model.RawItem {
	seen := make(map[string]struct{}, len(items))
	out := make([]model.RawItem, 0, len(items))

	for _, item := range items {
		key := strings.TrimSpace(item.URL)
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

func titleHash(title string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(title)), " ")
	h := sha1.Sum([]byte(normalized))
	return hex.EncodeToString(h[:])
}
