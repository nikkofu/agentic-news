package publish_test

import (
	"context"
	"strings"
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

func TestPublishEdition_ReturnsDeferredTransportHint(t *testing.T) {
	date := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	result, err := publish.PublishEdition(context.Background(), "/tmp/output/2026/03/19", date, publish.SFTPConfig{
		RemoteDir: "/var/www/news",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Published {
		t.Fatal("expected publish transport to remain deferred")
	}
	if result.Status != "transport_deferred" {
		t.Fatalf("expected transport_deferred status, got %q", result.Status)
	}
	if !strings.Contains(result.Message, "deferred") {
		t.Fatalf("expected deferred message, got %q", result.Message)
	}
	if result.Paths.Dated == "" || result.Paths.Latest == "" || result.Paths.Staging == "" {
		t.Fatal("expected remote paths to still be planned")
	}
}
