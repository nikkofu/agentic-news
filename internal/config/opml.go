package config

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultOPMLSourceType      = "blog"
	defaultOPMLCredibilityBase = 70
)

type opmlDocument struct {
	Body opmlBody `xml:"body"`
}

type opmlBody struct {
	Outlines []opmlOutline `xml:"outline"`
}

type opmlOutline struct {
	Text     string        `xml:"text,attr"`
	Title    string        `xml:"title,attr"`
	Type     string        `xml:"type,attr"`
	XMLURL   string        `xml:"xmlUrl,attr"`
	HTMLURL  string        `xml:"htmlUrl,attr"`
	Outlines []opmlOutline `xml:"outline"`
}

func expandOPMLSources(baseDir string, rss *RSSConfig) error {
	if rss == nil || len(rss.OPMLFiles) == 0 {
		return nil
	}

	merged := append([]RSSSource(nil), rss.Sources...)
	seenURLs := make(map[string]struct{}, len(merged))
	seenIDs := make(map[string]int, len(merged))
	for _, source := range merged {
		if normalized := normalizeSourceURL(source.RSSURL); normalized != "" {
			seenURLs[normalized] = struct{}{}
		}
		if trimmed := strings.TrimSpace(source.SourceID); trimmed != "" {
			seenIDs[trimmed]++
		}
	}

	for _, rawPath := range rss.OPMLFiles {
		opmlPath := resolveConfigPath(baseDir, rawPath)
		imported, err := parseOPMLSources(opmlPath)
		if err != nil {
			return err
		}
		for _, source := range imported {
			normalized := normalizeSourceURL(source.RSSURL)
			if normalized == "" {
				continue
			}
			if _, exists := seenURLs[normalized]; exists {
				continue
			}
			seenURLs[normalized] = struct{}{}
			source.SourceID = uniqueSourceID(source.SourceID, seenIDs)
			merged = append(merged, source)
		}
	}

	rss.Sources = merged
	return nil
}

func parseOPMLSources(path string) ([]RSSSource, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read opml %s: %w", path, err)
	}

	var doc opmlDocument
	if err := xml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse opml %s: %w", path, err)
	}

	sources := make([]RSSSource, 0)
	seenIDs := make(map[string]int)
	for _, outline := range doc.Body.Outlines {
		collectOPMLSources(outline, &sources, seenIDs)
	}
	return sources, nil
}

func collectOPMLSources(outline opmlOutline, sources *[]RSSSource, seenIDs map[string]int) {
	if feedURL := strings.TrimSpace(outline.XMLURL); feedURL != "" {
		name := firstNonEmpty(strings.TrimSpace(outline.Title), strings.TrimSpace(outline.Text), hostFromURL(outline.HTMLURL), hostFromURL(feedURL), "Imported Feed")
		domain := firstNonEmpty(hostFromURL(outline.HTMLURL), hostFromURL(feedURL))
		sourceID := slugifySourceID(firstNonEmpty(name, domain, feedURL))
		if sourceID == "" {
			sourceID = "imported-feed"
		}
		*sources = append(*sources, RSSSource{
			SourceID:        uniqueSourceID(sourceID, seenIDs),
			Name:            name,
			RSSURL:          feedURL,
			Domain:          domain,
			SourceType:      defaultOPMLSourceType,
			CredibilityBase: defaultOPMLCredibilityBase,
		})
	}

	for _, child := range outline.Outlines {
		collectOPMLSources(child, sources, seenIDs)
	}
}

func resolveConfigPath(baseDir, raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) {
		return trimmed
	}
	if _, err := os.Stat(trimmed); err == nil {
		return trimmed
	}
	if baseDir == "" {
		return trimmed
	}
	candidate := filepath.Join(baseDir, trimmed)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return trimmed
}

func normalizeSourceURL(raw string) string {
	return strings.TrimSpace(raw)
}

func hostFromURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Host)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func slugifySourceID(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, ch := range raw {
		switch {
		case ch >= 'a' && ch <= 'z':
			b.WriteRune(ch)
			lastDash = false
		case ch >= '0' && ch <= '9':
			b.WriteRune(ch)
			lastDash = false
		default:
			if b.Len() == 0 || lastDash {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}

func uniqueSourceID(base string, seen map[string]int) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		trimmed = "source"
	}
	if seen[trimmed] == 0 {
		seen[trimmed] = 1
		return trimmed
	}
	index := seen[trimmed]
	seen[trimmed] = index + 1
	return fmt.Sprintf("%s-%d", trimmed, index+1)
}
