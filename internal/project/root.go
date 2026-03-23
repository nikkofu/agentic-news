package project

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func Root() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to resolve project source location")
	}

	cur := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		next := filepath.Dir(cur)
		if next == cur {
			return "", fmt.Errorf("project root not found from %s", file)
		}
		cur = next
	}
}
