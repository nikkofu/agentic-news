package apps_test

import (
	"os"
	"strings"
	"testing"
)

func TestSanctuaryHydrationContractsAreWired(t *testing.T) {
	appJS, err := os.ReadFile("../../apps/js/app.js")
	if err != nil {
		t.Fatalf("expected apps/js/app.js to be readable, got %v", err)
	}

	content := string(appJS)
	requiredSnippets := []string{
		"hydrateReflectionPage",
		"hydrateCommunityPage",
		"hydrateUpgradePage",
		"/api/v1/reflections",
		"/api/v1/community/preview",
		"/api/v1/upgrade/offer",
		"data-reflection-compose-form",
		"data-reflection-list",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected apps/js/app.js to contain %q", snippet)
		}
	}
}
