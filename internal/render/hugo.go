package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nikkofu/agentic-news/internal/model"
)

type HugoExecRequest struct {
	Source      string
	Destination string
	Theme       string
	ConfigPath  string
	ThemesDir   string
}

type HugoRequest struct {
	PackageRoot string
	OutputRoot  string
	ThemeID     string
	Exec        func(HugoExecRequest) error
}

func RenderHugo(req HugoRequest) error {
	themeID := model.NormalizeThemeID(req.ThemeID)
	if err := validateHugoPaths(req.PackageRoot, req.OutputRoot); err != nil {
		return err
	}
	if _, err := os.Stat(req.PackageRoot); err != nil {
		return fmt.Errorf("package root missing: %w", err)
	}
	if err := os.RemoveAll(req.OutputRoot); err != nil {
		return err
	}

	workspace, err := resolveHugoWorkspace()
	if err != nil {
		return err
	}

	execFn := req.Exec
	if execFn == nil {
		execFn = runHugoExec
	}
	if err := execFn(HugoExecRequest{
		Source:      req.PackageRoot,
		Destination: req.OutputRoot,
		Theme:       themeID,
		ConfigPath:  workspace.ConfigPath,
		ThemesDir:   workspace.ThemesDir,
	}); err != nil {
		return err
	}

	return copyHugoPackageArtifacts(req.PackageRoot, req.OutputRoot)
}

func runHugoExec(req HugoExecRequest) error {
	source, err := filepath.Abs(req.Source)
	if err != nil {
		return err
	}
	destination, err := filepath.Abs(req.Destination)
	if err != nil {
		return err
	}
	configPath, err := filepath.Abs(req.ConfigPath)
	if err != nil {
		return err
	}
	themesDir, err := filepath.Abs(req.ThemesDir)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"hugo",
		"--source", source,
		"--destination", destination,
		"--config", configPath,
		"--themesDir", themesDir,
		"--theme", req.Theme,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
