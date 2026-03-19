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
		fmt.Println("usage: daily-builder <run|sample|publish-sample> ...")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		opts, err := run.ParseRunArgs(os.Args[2:])
		if err != nil {
			fmt.Printf("parse args failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("daily-builder run date=%s mode=%s dry_run=%v\n", opts.Date, opts.Mode, opts.Dry)
	case "sample":
		outDir := filepath.Join("output")
		date := time.Now()
		if len(os.Args) >= 3 {
			if parsed, err := time.Parse("2006-01-02", os.Args[2]); err == nil {
				date = parsed
			}
		}
		root, err := run.GenerateSampleEdition(outDir, date)
		if err != nil {
			fmt.Printf("generate sample failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("sample edition generated at %s\n", root)
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
		localDir := filepath.Join("output", date.Format("2006"), date.Format("01"), date.Format("02"))
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
		fmt.Printf("sample published: staging=%s dated=%s latest=%s\n", result.Paths.Staging, result.Paths.Dated, result.Paths.Latest)
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
