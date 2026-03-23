package run

import (
	"errors"
	"flag"

	"github.com/nikkofu/agentic-news/internal/model"
)

type RunOptions struct {
	Date string
	Mode string
	Dry  bool
	Theme string
}

func ParseRunArgs(args []string) (RunOptions, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	var opts RunOptions
	fs.StringVar(&opts.Date, "date", "today", "edition date, e.g. today or 2026-03-19")
	fs.StringVar(&opts.Mode, "mode", "morning", "run mode")
	fs.BoolVar(&opts.Dry, "dry-run", false, "run without publish")
	fs.StringVar(&opts.Theme, "theme", model.DefaultThemeID, "render theme ID")
	if err := fs.Parse(args); err != nil {
		return RunOptions{}, err
	}
	if opts.Mode == "" {
		return RunOptions{}, errors.New("mode is required")
	}
	opts.Theme = model.NormalizeThemeID(opts.Theme)
	return opts, nil
}
