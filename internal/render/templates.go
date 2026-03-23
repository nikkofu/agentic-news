package render

import (
	"html/template"
	"path/filepath"

	"github.com/nikkofu/agentic-news/internal/project"
)

type Renderer struct {
	theme      Theme
	indexTpl   *template.Template
	articleTpl *template.Template
}

func NewRenderer(themeID string) (*Renderer, error) {
	theme, err := ResolveTheme(themeID)
	if err != nil {
		return nil, err
	}

	root, err := project.Root()
	if err != nil {
		return nil, err
	}
	indexTpl, err := template.ParseFiles(filepath.Join(root, theme.IndexTemplate))
	if err != nil {
		return nil, err
	}
	articleTpl, err := template.ParseFiles(filepath.Join(root, theme.ArticleTemplate))
	if err != nil {
		return nil, err
	}
	return &Renderer{theme: theme, indexTpl: indexTpl, articleTpl: articleTpl}, nil
}
