package content

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
)

var (
	reTag      = regexp.MustCompile(`(?s)<[^>]*>`)
	reMetaOG   = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["'][^>]*>`)
	reMetaOG2  = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["'][^>]*>`)
	reParagraph = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
)

func ExtractArticle(ctx context.Context, item model.RawItem) (model.Article, error) {
	if strings.TrimSpace(item.URL) == "" {
		return model.Article{}, fmt.Errorf("raw item url is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
	if err != nil {
		return model.Article{}, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.Article{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Article{}, err
	}
	html := string(body)
	text := extractText(html)

	article := model.Article{
		Title:        item.Title,
		CanonicalURL: item.URL,
		CoverImage:   extractOGImage(html),
		ContentText:  text,
		Excerpt:      excerpt(text, 180),
		PublishedAt:  item.PublishedAt,
		IngestedAt:   time.Now(),
		Language:     DetectLanguage(text),
	}

	return article, nil
}

func extractOGImage(html string) string {
	if m := reMetaOG.FindStringSubmatch(html); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := reMetaOG2.FindStringSubmatch(html); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractText(html string) string {
	matches := reParagraph.FindAllStringSubmatch(html, -1)
	if len(matches) > 0 {
		parts := make([]string, 0, len(matches))
		for _, m := range matches {
			p := strings.TrimSpace(reTag.ReplaceAllString(m[1], " "))
			if p != "" {
				parts = append(parts, normalizeSpaces(p))
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n")
		}
	}
	return normalizeSpaces(strings.TrimSpace(reTag.ReplaceAllString(html, " ")))
}

func normalizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func excerpt(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max])
}
