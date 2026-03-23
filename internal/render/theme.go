package render

import (
	"fmt"
	"path/filepath"

	"github.com/nikkofu/agentic-news/internal/model"
)

type Theme struct {
	ID              string
	IndexTemplate   string
	ArticleTemplate string
}

var themes = map[string]Theme{
	"editorial-ai": {
		ID:              "editorial-ai",
		IndexTemplate:   filepath.Join("web", "templates", "themes", "editorial-ai", "index.tmpl"),
		ArticleTemplate: filepath.Join("web", "templates", "themes", "editorial-ai", "article.tmpl"),
	},
	"ai-product-magazine": {
		ID:              "ai-product-magazine",
		IndexTemplate:   filepath.Join("web", "templates", "themes", "ai-product-magazine", "index.tmpl"),
		ArticleTemplate: filepath.Join("web", "templates", "themes", "ai-product-magazine", "article.tmpl"),
	},
	"youth-signal": {
		ID:              "youth-signal",
		IndexTemplate:   filepath.Join("web", "templates", "themes", "youth-signal", "index.tmpl"),
		ArticleTemplate: filepath.Join("web", "templates", "themes", "youth-signal", "article.tmpl"),
	},
	"soft-focus": {
		ID:              "soft-focus",
		IndexTemplate:   filepath.Join("web", "templates", "themes", "soft-focus", "index.tmpl"),
		ArticleTemplate: filepath.Join("web", "templates", "themes", "soft-focus", "article.tmpl"),
	},
}

func ResolveTheme(raw string) (Theme, error) {
	themeID := model.NormalizeThemeID(raw)
	theme, ok := themes[themeID]
	if !ok {
		return Theme{}, fmt.Errorf("unknown theme: %s", themeID)
	}
	return theme, nil
}
