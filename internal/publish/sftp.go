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
	Status    string
	Message   string
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
	return PublishResult{
		Paths:     paths,
		Status:    "transport_deferred",
		Message:   "SFTP transport is deferred in the current mainline iteration; remote paths were planned only.",
		Published: false,
	}, nil
}
