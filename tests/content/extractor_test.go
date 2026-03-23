package content_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikkofu/agentic-news/internal/content"
	"github.com/nikkofu/agentic-news/internal/model"
)

func TestExtractArticle_ReturnsCleanTextAndCover(t *testing.T) {
	html := `<!doctype html><html><head><meta property="og:image" content="https://img.example.com/cover.jpg"></head><body><article><h1>Title</h1><p>第一段内容</p><p>Second paragraph</p></article></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	article, status, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if article.ContentText == "" {
		t.Fatal("expected text")
	}
	if article.CoverImage != "https://img.example.com/cover.jpg" {
		t.Fatalf("expected cover image, got %s", article.CoverImage)
	}
	if article.Language == "" {
		t.Fatal("expected language detection")
	}
	if !status.StandardEligible {
		t.Fatalf("expected standard-eligible extraction, got %+v", status)
	}
}

func TestExtractArticle_FallsBackToBodyTextWhenNoParagraphTags(t *testing.T) {
	html := `<!doctype html><html><head><title>Noise Title</title><script>console.log("ignore me")</script></head><body><article><div>First body block</div><div>Second body block</div></article></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	article, status, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(article.ContentText, "First body block") || !strings.Contains(article.ContentText, "Second body block") {
		t.Fatalf("expected body fallback text, got %q", article.ContentText)
	}
	if strings.Contains(article.ContentText, "Noise Title") || strings.Contains(article.ContentText, "ignore me") {
		t.Fatalf("expected head/script noise to be excluded, got %q", article.ContentText)
	}
	if !status.UsedFallbackText {
		t.Fatalf("expected fallback extraction status, got %+v", status)
	}
}

func TestExtractArticle_LeavesCoverImageEmptyWhenOGImageMissing(t *testing.T) {
	html := `<!doctype html><html><body><article><img src="https://img.example.com/body-1.jpg" alt="cover"><div>Body</div><img src="https://img.example.com/body-2.jpg"></article></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	article, _, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if article.CoverImage != "" {
		t.Fatalf("expected empty cover image when og:image is missing, got %q", article.CoverImage)
	}
	candidates := article.ImageCandidates
	if len(candidates) != 2 {
		t.Fatalf("expected 2 body image candidates, got %d", len(candidates))
	}
	if candidates[0].URL != "https://img.example.com/body-1.jpg" || candidates[0].Source != "body" {
		t.Fatalf("expected first body candidate, got %+v", candidates[0])
	}
	if candidates[1].URL != "https://img.example.com/body-2.jpg" || candidates[1].Source != "body" {
		t.Fatalf("expected second body candidate, got %+v", candidates[1])
	}
}

func TestExtractArticle_FlagsWeakFallbackContent(t *testing.T) {
	html := `<!doctype html><html><body><div>tiny</div></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	_, status, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status.StandardEligible {
		t.Fatalf("expected weak fallback content to be downgrade-only, got %+v", status)
	}
	if !status.UsedFallbackText {
		t.Fatalf("expected fallback text path, got %+v", status)
	}
	if status.FallbackReason == "" {
		t.Fatalf("expected fallback reason, got %+v", status)
	}
}

func TestExtractArticle_FlagsStrongParagraphExtractionAsStandardEligible(t *testing.T) {
	html := `<!doctype html><html><body><article><p>First paragraph contains enough signal to be useful.</p><p>Second paragraph adds supporting context and detail.</p></article></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	_, status, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !status.StandardEligible {
		t.Fatalf("expected paragraph extraction to remain standard-eligible, got %+v", status)
	}
	if status.UsedFallbackText {
		t.Fatalf("expected paragraph extraction to avoid fallback path, got %+v", status)
	}
}

func TestExtractArticle_ReadsLocalHTMLFile(t *testing.T) {
	dir := t.TempDir()
	articlePath := filepath.Join(dir, "article.html")
	html := `<!doctype html><html><body><article><p>Local paragraph one with enough content.</p><p>Local paragraph two with enough content.</p></article></body></html>`
	if err := os.WriteFile(articlePath, []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := model.RawItem{Title: "Local Test", URL: articlePath}
	article, status, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(article.ContentText, "Local paragraph one") {
		t.Fatalf("expected local article content, got %q", article.ContentText)
	}
	if !status.StandardEligible {
		t.Fatalf("expected local html file to be standard-eligible, got %+v", status)
	}
}

func TestExtractArticle_PrefersOGImageButRetainsOrderedBodyCandidates(t *testing.T) {
	html := `<!doctype html><html><head><meta property="og:image" content="https://img.example.com/og.jpg"></head><body><article><img src="https://img.example.com/body-1.jpg"><p>Body paragraph.</p><img src="https://img.example.com/body-1.jpg"><img src="https://img.example.com/body-2.jpg"></article></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	article, _, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if article.CoverImage != "https://img.example.com/og.jpg" {
		t.Fatalf("expected og cover image, got %q", article.CoverImage)
	}

	candidates := article.ImageCandidates
	if len(candidates) != 4 {
		t.Fatalf("expected 4 image candidates (og + 3 body, including duplicate), got %d", len(candidates))
	}
	if candidates[0].URL != "https://img.example.com/og.jpg" || candidates[0].Source != "og" {
		t.Fatalf("expected first candidate to be og image, got %+v", candidates[0])
	}
	if candidates[1].URL != "https://img.example.com/body-1.jpg" || candidates[1].Source != "body" {
		t.Fatalf("expected first body image candidate, got %+v", candidates[1])
	}
	if candidates[2].URL != "https://img.example.com/body-1.jpg" || candidates[2].Source != "body" {
		t.Fatalf("expected duplicate body image candidate to be preserved, got %+v", candidates[2])
	}
	if candidates[3].URL != "https://img.example.com/body-2.jpg" || candidates[3].Source != "body" {
		t.Fatalf("expected third body image candidate, got %+v", candidates[3])
	}
}

func TestExtractArticle_ResolvesRelativeBodyImageURLsAgainstArticleURL(t *testing.T) {
	html := `<!doctype html><html><body><article><img src="/images/hero.jpg"><p>Body paragraph.</p></article></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL + "/posts/a"}
	article, _, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	candidates := article.ImageCandidates
	if len(candidates) != 1 {
		t.Fatalf("expected 1 body candidate, got %d", len(candidates))
	}
	if candidates[0].URL != server.URL+"/images/hero.jpg" {
		t.Fatalf("expected resolved absolute body image url, got %q", candidates[0].URL)
	}
	if candidates[0].Source != "body" {
		t.Fatalf("expected body source, got %q", candidates[0].Source)
	}
}

func TestExtractArticle_IgnoresScriptAndStyleImageLikeSnippets(t *testing.T) {
	html := `<!doctype html><html><body>
<script>var fake = '<img src="https://img.example.com/fake-script.jpg">';</script>
<style>.hero:before { content: '<img src="https://img.example.com/fake-style.jpg">'; }</style>
<article><img src="https://img.example.com/real.jpg"><p>Body paragraph.</p></article>
</body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	raw := model.RawItem{Title: "Test", URL: server.URL}
	article, _, err := content.ExtractArticle(context.Background(), raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	candidates := article.ImageCandidates
	if len(candidates) != 1 {
		t.Fatalf("expected only 1 real body image candidate, got %d", len(candidates))
	}
	if candidates[0].URL != "https://img.example.com/real.jpg" || candidates[0].Source != "body" {
		t.Fatalf("expected only real body image candidate, got %+v", candidates[0])
	}
}
