package run

import (
	"errors"
	"flag"
)

type RunOptions struct {
	Date string
	Mode string
	Dry  bool
}

func ParseRunArgs(args []string) (RunOptions, error) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	var opts RunOptions
	fs.StringVar(&opts.Date, "date", "today", "edition date, e.g. today or 2026-03-19")
	fs.StringVar(&opts.Mode, "mode", "morning", "run mode")
	fs.BoolVar(&opts.Dry, "dry-run", false, "run without publish")
	if err := fs.Parse(args); err != nil {
		return RunOptions{}, err
	}
	if opts.Mode == "" {
		return RunOptions{}, errors.New("mode is required")
	}
	return opts, nil
}
