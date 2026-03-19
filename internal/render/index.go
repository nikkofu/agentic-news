package render

import (
	"os"
	"path/filepath"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
)

type indexData struct {
	DateLabel string
	Keywords  []string
	Featured  []model.DailyPick
	Learning  []string
}

func DailyEdition(baseDir string, daily model.DailyEdition) error {
	r, err := NewRenderer()
	if err != nil {
		return err
	}

	root, err := output.EnsureDateDirs(baseDir, daily)
	if err != nil {
		return err
	}

	if err := writeIndex(filepath.Join(root, "index.html"), r, daily); err != nil {
		return err
	}
	if err := writeArticles(filepath.Join(root, "articles"), r, daily); err != nil {
		return err
	}
	if err := copyStaticAssets(filepath.Join(root, "assets")); err != nil {
		return err
	}
	if err := output.WriteJSON(filepath.Join(root, "data", "daily.json"), daily); err != nil {
		return err
	}
	if err := output.WriteJSON(filepath.Join(root, "data", "learning.json"), map[string]any{"learning": daily.Learning}); err != nil {
		return err
	}
	return output.WriteJSON(filepath.Join(root, "meta.json"), map[string]any{
		"generated_at": time.Now().Format(time.RFC3339),
		"featured":     len(daily.Featured),
	})
}

func writeIndex(path string, r *Renderer, daily model.DailyEdition) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data := indexData{
		DateLabel: daily.Date.Format("2006-01-02"),
		Keywords:  daily.Keywords,
		Featured:  daily.Featured,
		Learning:  daily.Learning,
	}
	return r.indexTpl.Execute(f, data)
}

func copyStaticAssets(assetDir string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	css, err := os.ReadFile(filepath.Join(root, "web", "static", "styles.css"))
	if err != nil {
		return err
	}
	js, err := os.ReadFile(filepath.Join(root, "web", "static", "app.js"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetDir, "styles.css"), css, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(assetDir, "app.js"), js, 0o644)
}
