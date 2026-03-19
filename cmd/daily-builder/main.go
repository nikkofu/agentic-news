package main

import (
	"fmt"
	"os"

	"github.com/nikkofu/agentic-news/internal/run"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: daily-builder run --date today --mode morning [--dry-run]")
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
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
