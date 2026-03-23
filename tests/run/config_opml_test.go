package run_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nikkofu/agentic-news/internal/config"
)

func TestLoadConfig_MergesOPMLSources(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "ai.yaml"), []byte("ai:\n  quality_mode: high\n  provider: test\n  model: test-model\n"), 0o644); err != nil {
		t.Fatalf("write ai config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scoring.yaml"), []byte("weights:\n  importance: 0.30\n  personal_relevance: 0.25\n  credibility: 0.20\n  novelty: 0.15\n  freshness: 0.10\n"), 0o644); err != nil {
		t.Fatalf("write scoring config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rss_sources.yaml"), []byte("rss:\n  sources:\n    - source_id: curated_1\n      name: Curated Source\n      rss_url: https://example.com/feed.xml\n      domain: example.com\n      source_type: media\n      credibility_base: 82\n  opml_files:\n    - feeds.opml\n"), 0o644); err != nil {
		t.Fatalf("write rss config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "feeds.opml"), []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <body>
    <outline text="Blogs">
      <outline type="rss" text="Example Blog" title="Example Blog" xmlUrl="https://blog.example.com/rss.xml" htmlUrl="https://blog.example.com"/>
      <outline type="rss" text="Another Feed" xmlUrl="https://another.example.com/feed"/>
    </outline>
  </body>
</opml>`), 0o644); err != nil {
		t.Fatalf("write opml file: %v", err)
	}

	cfg, err := config.LoadConfig(dir)
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}

	if len(cfg.RSS.Sources) != 3 {
		t.Fatalf("expected 3 rss sources after opml import, got %d", len(cfg.RSS.Sources))
	}

	if cfg.RSS.Sources[1].RSSURL != "https://blog.example.com/rss.xml" {
		t.Fatalf("expected first imported feed url to be preserved, got %q", cfg.RSS.Sources[1].RSSURL)
	}
	if cfg.RSS.Sources[1].Name != "Example Blog" {
		t.Fatalf("expected first imported feed name, got %q", cfg.RSS.Sources[1].Name)
	}
	if cfg.RSS.Sources[1].Domain != "blog.example.com" {
		t.Fatalf("expected first imported feed domain, got %q", cfg.RSS.Sources[1].Domain)
	}
	if cfg.RSS.Sources[1].SourceType == "" {
		t.Fatal("expected imported feed source type to be populated")
	}
	if cfg.RSS.Sources[1].CredibilityBase == 0 {
		t.Fatal("expected imported feed credibility to be populated")
	}

	if cfg.RSS.Sources[2].Name != "Another Feed" {
		t.Fatalf("expected second imported feed name, got %q", cfg.RSS.Sources[2].Name)
	}
	if cfg.RSS.Sources[2].Domain != "another.example.com" {
		t.Fatalf("expected second imported feed domain, got %q", cfg.RSS.Sources[2].Domain)
	}
}
