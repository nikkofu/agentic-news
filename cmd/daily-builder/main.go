package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nikkofu/agentic-news/internal/publish"
	"github.com/nikkofu/agentic-news/internal/run"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: daily-builder <run|sample|render-hugo|publish-sample> ...")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		opts, err := run.ParseRunArgs(os.Args[2:])
		if err != nil {
			fmt.Printf("parse args failed: %v\n", err)
			os.Exit(1)
		}
		date, err := parseEditionDate(opts.Date)
		if err != nil {
			fmt.Printf("invalid run date: %v\n", err)
			os.Exit(1)
		}
		result, err := run.RunDryPipeline(context.Background(), run.DryRunRequest{
			ConfigDir: "config",
			OutputDir: "output",
			StateDir:  "state",
			Date:      date,
			Mode:      opts.Mode,
			Theme:     opts.Theme,
		}, run.DryRunHooks{})
		if err != nil {
			fmt.Printf("run pipeline failed: %v\n", err)
			os.Exit(1)
		}
		if opts.Dry {
			fmt.Printf("dry-run daily pipeline generated at %s featured=%d fallback=%v\n", result.OutputRoot, result.FeaturedCount, result.UsedFallback)
			return
		}
		fmt.Printf("daily pipeline generated locally at %s featured=%d fallback=%v publish=deferred\n", result.OutputRoot, result.FeaturedCount, result.UsedFallback)
	case "sample":
		outDir := filepath.Join("output")
		opts, err := run.ParseSampleArgs(os.Args[2:])
		if err != nil {
			fmt.Printf("parse sample args failed: %v\n", err)
			os.Exit(1)
		}
		date, err := parseEditionDate(opts.Date)
		if err != nil {
			fmt.Printf("invalid sample date: %v\n", err)
			os.Exit(1)
		}
		result, err := run.GenerateSampleEdition(outDir, date, opts.Theme)
		if err != nil {
			fmt.Printf("generate sample failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("sample edition generated\noutput=%s\npackage=%s\n", result.OutputRoot, result.PackageRoot)
	case "render-hugo":
		opts, err := run.ParseRenderHugoArgs(os.Args[2:])
		if err != nil {
			fmt.Printf("parse render-hugo args failed: %v\n", err)
			os.Exit(1)
		}
		date, err := parseEditionDate(opts.Date)
		if err != nil {
			fmt.Printf("invalid render-hugo date: %v\n", err)
			os.Exit(1)
		}
		result, err := run.RenderHugoEdition(filepath.Join("output"), date, opts.Theme)
		if err != nil {
			fmt.Printf("render hugo failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("hugo render generated\noutput=%s\npackage=%s\n", result.OutputRoot, result.PackageRoot)
	case "publish-sample":
		opts, err := run.ParsePublishSampleArgs(os.Args[2:])
		if err != nil {
			fmt.Printf("parse publish-sample args failed: %v\n", err)
			os.Exit(1)
		}
		date := time.Now()
		if opts.Date != "today" {
			if parsed, err := time.Parse("2006-01-02", opts.Date); err == nil {
				date = parsed
			} else {
				fmt.Printf("invalid date: %s\n", opts.Date)
				os.Exit(1)
			}
		}
		paths := publish.BuildRemotePaths(os.Getenv("SFTP_REMOTE_DIR"), date)
		localDir := run.SampleEditionRoot(filepath.Join("output"), date, opts.Theme)
		if opts.DryRun {
			fmt.Printf("dry-run publish sample\nlocal=%s\nstaging=%s\ndated=%s\nlatest=%s\n", localDir, paths.Staging, paths.Dated, paths.Latest)
			return
		}
		cfg := publish.SFTPConfig{
			Host:      os.Getenv("SFTP_HOST"),
			User:      os.Getenv("SFTP_USER"),
			KeyPath:   os.Getenv("SFTP_KEY_PATH"),
			RemoteDir: os.Getenv("SFTP_REMOTE_DIR"),
			Port:      22,
		}
		result, err := publish.PublishEdition(context.Background(), localDir, date, cfg)
		if err != nil {
			fmt.Printf("publish sample failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("sample publish status=%s published=%v\nmessage=%s\nstaging=%s\ndated=%s\nlatest=%s\n", result.Status, result.Published, result.Message, result.Paths.Staging, result.Paths.Dated, result.Paths.Latest)
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func parseEditionDate(raw string) (time.Time, error) {
	if raw == "" || raw == "today" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
