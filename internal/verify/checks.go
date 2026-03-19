package verify

import (
	"fmt"
	"os"
	"path/filepath"
)

func DailyEdition(dir string) error {
	required := []string{
		filepath.Join(dir, "index.html"),
		filepath.Join(dir, "data", "daily.json"),
		filepath.Join(dir, "meta.json"),
	}

	for _, p := range required {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("missing required file: %s", p)
		}
	}
	return nil
}
