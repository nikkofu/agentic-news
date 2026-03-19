package render

import (
	"os"
	"path/filepath"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
)

type articleData struct {
	ID          string
	Category    string
	Title       string
	Summary     string
	ScoreFinal  float64
	CoverImage  string
	SourceName  string
	SourceURL   string
	PublishedAt string
	Viewpoint   string
}

func writeArticles(articlesDir string, r *Renderer, daily model.DailyEdition) error {
	for _, item := range daily.Featured {
		if item.ID == "" {
			continue
		}
		path := filepath.Join(articlesDir, item.ID+".html")
		if err := writeSingleArticle(path, r, item); err != nil {
			return err
		}
	}
	return nil
}

func writeSingleArticle(path string, r *Renderer, item model.DailyPick) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	published := item.PublishedAt
	if published.IsZero() {
		published = time.Now()
	}

	data := articleData{
		ID:          item.ID,
		Category:    item.Category,
		Title:       item.Title,
		Summary:     item.Summary,
		ScoreFinal:  item.ScoreFinal,
		CoverImage:  item.CoverImage,
		SourceName:  item.SourceName,
		SourceURL:   item.SourceURL,
		PublishedAt: published.Format(time.RFC3339),
		Viewpoint:   item.Insight.Viewpoint,
	}
	return r.articleTpl.Execute(f, data)
}
