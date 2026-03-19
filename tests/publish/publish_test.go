package publish_test

import (
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/publish"
)

func TestBuildRemotePaths_UsesDateAndLatestTargets(t *testing.T) {
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	paths := publish.BuildRemotePaths("/var/www/news", date)
	if paths.Staging == "" || paths.Latest == "" || paths.Dated == "" {
		t.Fatal("expected non-empty remote paths")
	}
	if paths.Latest != "/var/www/news/latest" {
		t.Fatalf("unexpected latest path: %s", paths.Latest)
	}
}
