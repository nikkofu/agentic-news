package content

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nikkofu/agentic-news/internal/model"
)

var (
	reTag       = regexp.MustCompile(`(?s)<[^>]*>`)
	reMetaOG    = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["'][^>]*>`)
	reMetaOG2   = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["'][^>]*>`)
	reBody      = regexp.MustCompile(`(?is)<body[^>]*>(.*?)</body>`)
	reScriptCSS = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>`)
	reImageSrc  = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["'][^>]*>`)
	reParagraph = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
)

type ExtractionStatus struct {
	StandardEligible bool
	UsedFallbackText bool
	FallbackReason   string
}

func ExtractArticle(ctx context.Context, item model.RawItem) (model.Article, ExtractionStatus, error) {
	if strings.TrimSpace(item.URL) == "" {
		return model.Article{}, ExtractionStatus{}, fmt.Errorf("raw item url is required")
	}

	body, err := readArticleBody(ctx, item.URL)
	if err != nil {
		return model.Article{}, ExtractionStatus{}, err
	}
	html := string(body)
	text, usedFallbackText := extractText(html)
	ogImage := resolveImageURL(extractOGImage(html), item.URL)
	bodyImages := extractBodyImageURLs(html, item.URL)
	imageCandidates := buildImageCandidates(ogImage, bodyImages)

	article := model.Article{
		Title:           item.Title,
		CanonicalURL:    item.URL,
		CoverImage:      ogImage,
		ImageCandidates: imageCandidates,
		ContentText:     text,
		Excerpt:         excerpt(text, 180),
		PublishedAt:     item.PublishedAt,
		IngestedAt:      time.Now(),
		Language:        DetectLanguage(text),
	}

	return article, classifyExtraction(text, usedFallbackText), nil
}

func readArticleBody(ctx context.Context, rawURL string) ([]byte, error) {
	if isRemoteContentURL(rawURL) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return io.ReadAll(resp.Body)
	}

	path, err := contentLocalPath(rawURL)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
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

func extractBodyImageURLs(html string, baseURL string) []string {
	bodyContent := html
	if m := reBody.FindStringSubmatch(html); len(m) > 1 {
		bodyContent = m[1]
	}
	bodyContent = reScriptCSS.ReplaceAllString(bodyContent, " ")

	matches := reImageSrc.FindAllStringSubmatch(bodyContent, -1)
	bodyImages := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		resolved := resolveImageURL(m[1], baseURL)
		if resolved == "" {
			continue
		}
		bodyImages = append(bodyImages, resolved)
	}
	return bodyImages
}

func buildImageCandidates(ogImage string, bodyImages []string) []model.ArticleImage {
	candidates := make([]model.ArticleImage, 0, len(bodyImages)+1)

	if ogImage != "" {
		candidates = append(candidates, model.ArticleImage{URL: ogImage, Source: "og"})
	}

	for _, bodyImage := range bodyImages {
		candidates = append(candidates, model.ArticleImage{URL: bodyImage, Source: "body"})
	}

	return candidates
}

func resolveImageURL(rawImageURL string, baseURL string) string {
	trimmed := strings.TrimSpace(rawImageURL)
	if trimmed == "" {
		return ""
	}

	imageRef, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	if imageRef.IsAbs() {
		return imageRef.String()
	}

	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || !base.IsAbs() {
		return trimmed
	}
	return base.ResolveReference(imageRef).String()
}

func extractText(html string) (string, bool) {
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
			return strings.Join(parts, "\n"), false
		}
	}

	body := html
	if m := reBody.FindStringSubmatch(html); len(m) > 1 {
		body = m[1]
	}
	body = reScriptCSS.ReplaceAllString(body, " ")

	return normalizeSpaces(strings.TrimSpace(reTag.ReplaceAllString(body, " "))), true
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

func classifyExtraction(text string, usedFallbackText bool) ExtractionStatus {
	textLen := utf8.RuneCountInString(strings.TrimSpace(text))
	status := ExtractionStatus{
		StandardEligible: !usedFallbackText && textLen >= 20,
		UsedFallbackText: usedFallbackText,
	}

	switch {
	case textLen < 20 && usedFallbackText:
		status.FallbackReason = "weak_fallback_text"
	case textLen < 20:
		status.FallbackReason = "content_too_short"
	case usedFallbackText:
		status.FallbackReason = "fallback_text_used"
	default:
		status.FallbackReason = ""
	}

	return status
}

func isRemoteContentURL(raw string) bool {
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

func contentLocalPath(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err == nil && parsed.Scheme == "file" {
		return parsed.Path, nil
	}
	return raw, nil
}
