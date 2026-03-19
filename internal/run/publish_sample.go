package run

import "flag"

type PublishSampleOptions struct {
	Date   string
	DryRun bool
}

func ParsePublishSampleArgs(args []string) (PublishSampleOptions, error) {
	fs := flag.NewFlagSet("publish-sample", flag.ContinueOnError)
	var opts PublishSampleOptions
	fs.StringVar(&opts.Date, "date", "", "sample date in YYYY-MM-DD")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print target paths without uploading")
	if err := fs.Parse(args); err != nil {
		return PublishSampleOptions{}, err
	}
	if opts.Date == "" {
		opts.Date = "today"
	}
	return opts, nil
}
