package content_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	article, err := content.ExtractArticle(context.Background(), raw)
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
}
