package render

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nikkofu/agentic-news/internal/project"
)

type HugoWorkspace struct {
	ConfigPath string
	ThemesDir  string
}

func resolveHugoWorkspace() (HugoWorkspace, error) {
	root, err := project.Root()
	if err != nil {
		return HugoWorkspace{}, err
	}
	return HugoWorkspace{
		ConfigPath: filepath.Join(root, "hugo", "config.toml"),
		ThemesDir:  filepath.Join(root, "hugo", "themes"),
	}, nil
}

func copyHugoPackageArtifacts(packageRoot, outputRoot string) error {
	if err := os.MkdirAll(filepath.Join(outputRoot, "data"), 0o755); err != nil {
		return err
	}

	if err := copyHugoFile(
		filepath.Join(packageRoot, "data", "daily.json"),
		filepath.Join(outputRoot, "data", "daily.json"),
	); err != nil {
		return err
	}
	if err := copyHugoFile(
		filepath.Join(packageRoot, "data", "learning.json"),
		filepath.Join(outputRoot, "data", "learning.json"),
	); err != nil {
		return err
	}
	return copyHugoFile(
		filepath.Join(packageRoot, "meta", "edition.json"),
		filepath.Join(outputRoot, "meta.json"),
	)
}

func copyHugoFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return nil
}

func validateHugoPaths(packageRoot, outputRoot string) error {
	if packageRoot == "" {
		return fmt.Errorf("package root is required")
	}
	if outputRoot == "" {
		return fmt.Errorf("output root is required")
	}
	return nil
}
