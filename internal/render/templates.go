package render

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

type Renderer struct {
	indexTpl   *template.Template
	articleTpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	root, err := findProjectRoot()
	if err != nil {
		return nil, err
	}
	indexTpl, err := template.ParseFiles(filepath.Join(root, "web", "templates", "index.tmpl"))
	if err != nil {
		return nil, err
	}
	articleTpl, err := template.ParseFiles(filepath.Join(root, "web", "templates", "article.tmpl"))
	if err != nil {
		return nil, err
	}
	return &Renderer{indexTpl: indexTpl, articleTpl: articleTpl}, nil
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		next := filepath.Dir(cur)
		if next == cur {
			return "", fmt.Errorf("project root not found from %s", wd)
		}
		cur = next
	}
}
