package run

import (
	"flag"

	"github.com/nikkofu/agentic-news/internal/model"
)

type PublishSampleOptions struct {
	Date   string
	DryRun bool
	Theme  string
}

func ParsePublishSampleArgs(args []string) (PublishSampleOptions, error) {
	fs := flag.NewFlagSet("publish-sample", flag.ContinueOnError)
	var opts PublishSampleOptions
	fs.StringVar(&opts.Date, "date", "", "sample date in YYYY-MM-DD")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print target paths without uploading")
	fs.StringVar(&opts.Theme, "theme", model.DefaultThemeID, "render theme ID")
	if err := fs.Parse(args); err != nil {
		return PublishSampleOptions{}, err
	}
	if opts.Date == "" {
		opts.Date = "today"
	}
	opts.Theme = model.NormalizeThemeID(opts.Theme)
	return opts, nil
}
