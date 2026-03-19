package analyze

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func RenderPrompt(promptPath string, data any) (string, error) {
	b, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("read prompt %s: %w", promptPath, err)
	}

	tpl, err := template.New(filepath.Base(promptPath)).Parse(string(b))
	if err != nil {
		return "", fmt.Errorf("parse prompt %s: %w", promptPath, err)
	}

	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("execute prompt %s: %w", promptPath, err)
	}
	return out.String(), nil
}
