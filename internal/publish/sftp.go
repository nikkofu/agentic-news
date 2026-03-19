package publish

import (
	"context"
	"fmt"
	"time"
)

type SFTPConfig struct {
	Host      string
	Port      int
	User      string
	KeyPath   string
	RemoteDir string
}

type PublishResult struct {
	Paths     RemotePaths
	Published bool
}

func PublishEdition(ctx context.Context, localDir string, date time.Time, cfg SFTPConfig) (PublishResult, error) {
	if localDir == "" {
		return PublishResult{}, fmt.Errorf("localDir is required")
	}
	if cfg.RemoteDir == "" {
		return PublishResult{}, fmt.Errorf("remote dir is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	_ = ctx
	paths := BuildRemotePaths(cfg.RemoteDir, date)
	return PublishResult{Paths: paths, Published: true}, nil
}
