package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikkofu/agentic-news/internal/model"
)

func DatePath(baseDir string, date model.DailyEdition) string {
	y, m, d := date.Date.Date()
	return filepath.Join(baseDir, fmt.Sprintf("%04d", y), fmt.Sprintf("%02d", int(m)), fmt.Sprintf("%02d", d))
}

func EnsureDateDirs(baseDir string, daily model.DailyEdition) (string, error) {
	root := DatePath(baseDir, daily)
	if err := os.MkdirAll(filepath.Join(root, "articles"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "data"), 0o755); err != nil {
		return "", err
	}
	return root, nil
}

func WriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
