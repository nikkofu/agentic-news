package run

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
	"github.com/nikkofu/agentic-news/internal/render"
)

type RenderHugoOptions struct {
	Date  string
	Theme string
}

type RenderHugoResult struct {
	OutputRoot  string
	PackageRoot string
}

func ParseRenderHugoArgs(args []string) (RenderHugoOptions, error) {
	fs := flag.NewFlagSet("render-hugo", flag.ContinueOnError)
	var opts RenderHugoOptions
	fs.StringVar(&opts.Date, "date", "today", "edition date, e.g. today or 2026-03-20")
	fs.StringVar(&opts.Theme, "theme", model.DefaultThemeID, "render theme ID")
	if err := fs.Parse(args); err != nil {
		return RenderHugoOptions{}, err
	}
	opts.Theme = model.NormalizeThemeID(opts.Theme)
	return opts, nil
}

func RenderHugoEdition(baseDir string, date time.Time, theme string) (RenderHugoResult, error) {
	packageRoot := output.PackagePath(baseDir, date)
	if _, err := os.Stat(packageRoot); err != nil {
		return RenderHugoResult{}, fmt.Errorf("package root missing: %w", err)
	}

	outputRoot := output.DatePath(baseDir, model.DailyEdition{
		Date:    date,
		ThemeID: model.NormalizeThemeID(theme),
	})
	if err := render.RenderHugo(render.HugoRequest{
		PackageRoot: packageRoot,
		OutputRoot:  outputRoot,
		ThemeID:     theme,
	}); err != nil {
		return RenderHugoResult{}, err
	}

	return RenderHugoResult{
		OutputRoot:  outputRoot,
		PackageRoot: packageRoot,
	}, nil
}
