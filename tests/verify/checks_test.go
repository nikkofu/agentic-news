package verify_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nikkofu/agentic-news/internal/verify"
)

func TestVerifyDailyEdition_RejectsMissingIndex(t *testing.T) {
	dir := t.TempDir()
	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
}

func TestVerifyDailyEdition_PassesWithRequiredFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.html"), "ok")
	mustWrite(t, filepath.Join(dir, "data", "daily.json"), "{}")
	mustWrite(t, filepath.Join(dir, "meta.json"), "{}")

	if err := verify.DailyEdition(dir); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
