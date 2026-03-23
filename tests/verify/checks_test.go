package verify_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	writeValidEditionFixture(t, dir, nil, nil)

	if err := verify.DailyEdition(dir); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestVerifyDailyEdition_RequiresStableHomepageHooks(t *testing.T) {
	dir := t.TempDir()
	writeValidEditionFixture(t, dir, func(indexHTML string) string {
		indexHTML = strings.Replace(indexHTML, `data-article-id="item-a"`, `data-article-id=""`, 1)
		return strings.Replace(indexHTML, `data-theme-id="editorial-ai"`, "", 1)
	}, nil)

	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "homepage") {
		t.Fatalf("expected homepage hook error, got %v", err)
	}
}

func TestVerifyDailyEdition_RequiresArticleFeedbackHooks(t *testing.T) {
	dir := t.TempDir()
	writeValidEditionFixture(t, dir, nil, func(articleHTML string) string {
		articleHTML = strings.Replace(articleHTML, `data-feedback-surface="article"`, "", 1)
		articleHTML = strings.Replace(articleHTML, `data-feedback-value="like"`, "", 1)
		articleHTML = strings.Replace(articleHTML, `data-feedback-value="dislike"`, "", 1)
		return strings.Replace(articleHTML, `data-feedback-value="bookmark"`, "", 1)
	})

	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "feedback") {
		t.Fatalf("expected feedback hook error, got %v", err)
	}
}

func TestDailyEdition_FailsWhenFeaturedCountBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "index.html"), "ok")
	mustWrite(t, filepath.Join(dir, "meta.json"), "{}")
	mustWriteDailyJSON(t, filepath.Join(dir, "data", "daily.json"), featuredIDs(9))

	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "featured") {
		t.Fatalf("expected featured count error, got %v", err)
	}
}

func TestDailyEdition_FailsWhenArticleFileMissing(t *testing.T) {
	dir := t.TempDir()
	writeValidEditionFixture(t, dir, nil, nil)
	ids := featuredIDs(10)
	if err := os.Remove(filepath.Join(dir, "articles", ids[9]+".html")); err != nil {
		t.Fatal(err)
	}

	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "article") {
		t.Fatalf("expected article presence error, got %v", err)
	}
}

func TestDailyEdition_FailsWhenStandardCardCountBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	writeValidEditionFixture(t, dir, nil, nil)

	items := featuredItems(10, "brief")
	items[0].CardType = "standard"
	items[1].CardType = "standard"
	mustWriteDailyItemsJSON(t, filepath.Join(dir, "data", "daily.json"), items)

	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "standard") {
		t.Fatalf("expected standard count error, got %v", err)
	}
}

func TestDailyEdition_FailsWhenBriefCardMissingFallbackReason(t *testing.T) {
	dir := t.TempDir()
	writeValidEditionFixture(t, dir, nil, nil)

	items := featuredItems(10, "standard")
	for i := 3; i < len(items); i++ {
		items[i].CardType = "brief"
		items[i].FallbackReason = ""
	}
	mustWriteDailyItemsJSON(t, filepath.Join(dir, "data", "daily.json"), items)

	err := verify.DailyEdition(dir)
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !strings.Contains(err.Error(), "fallback") {
		t.Fatalf("expected fallback reason error, got %v", err)
	}
}

func TestDailyEdition_AllowsMissingProfileSnapshotByVerifyingFallbackCopy(t *testing.T) {
	dir := t.TempDir()
	writeValidEditionFixture(t, dir, nil, nil)

	items := featuredItems(10, "standard")
	items[0].Insight.WhyForYou = "Selected for broad strategic relevance and signal density."
	items[0].Insight.TasteGrowthHint = "Keep building breadth with evidence-first explainers."
	items[0].Insight.KnowledgeGapHint = "Review market structure basics to sharpen future comparisons."
	mustWriteDailyItemsJSON(t, filepath.Join(dir, "data", "daily.json"), items)

	if err := verify.DailyEdition(dir); err != nil {
		t.Fatalf("expected fallback-safe personalized copy to verify, got %v", err)
	}
}

func TestDailyEdition_FailsWhenStandardCardMissingPersonalizedCopy(t *testing.T) {
	cases := []struct {
		name        string
		clearFields func(*insightPayload)
		wantError   string
	}{
		{
			name: "missing why_for_you",
			clearFields: func(i *insightPayload) {
				i.WhyForYou = ""
			},
			wantError: "why_for_you",
		},
		{
			name: "missing taste_growth_hint",
			clearFields: func(i *insightPayload) {
				i.TasteGrowthHint = ""
			},
			wantError: "taste_growth_hint",
		},
		{
			name: "missing knowledge_gap_hint",
			clearFields: func(i *insightPayload) {
				i.KnowledgeGapHint = ""
			},
			wantError: "knowledge_gap_hint",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeValidEditionFixture(t, dir, nil, nil)

			items := featuredItems(10, "standard")
			tc.clearFields(&items[0].Insight)
			mustWriteDailyItemsJSON(t, filepath.Join(dir, "data", "daily.json"), items)

			err := verify.DailyEdition(dir)
			if err == nil {
				t.Fatal("expected verification error")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("expected %s validation error, got %v", tc.wantError, err)
			}
		})
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

func mustWriteDailyJSON(t *testing.T, path string, ids []string) {
	t.Helper()
	items := make([]featuredItemPayload, 0, len(ids))
	for _, id := range ids {
		items = append(items, featuredItemPayload{
			ID:         id,
			CardType:   "standard",
			Summary:    "Summary",
			SourceName: "Example",
			SourceURL:  "https://example.com",
			Insight: insightPayload{
				Viewpoint:        "Viewpoint",
				WhyForYou:        "Selected for your interest in long-term strategy.",
				TasteGrowthHint:  "Keep balancing explainers with primary-source reporting.",
				KnowledgeGapHint: "Review the core market mechanics behind this topic.",
			},
			FallbackReason: "",
		})
	}
	mustWriteDailyItemsJSON(t, path, items)
}

func featuredIDs(n int) []string {
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, "item-"+string(rune('a'+i)))
	}
	return ids
}

type featuredItemPayload struct {
	ID             string         `json:"ID"`
	CardType       string         `json:"CardType"`
	FallbackReason string         `json:"FallbackReason"`
	Summary        string         `json:"Summary"`
	SourceName     string         `json:"SourceName"`
	SourceURL      string         `json:"SourceURL"`
	Insight        insightPayload `json:"Insight"`
}

type insightPayload struct {
	Viewpoint        string `json:"Viewpoint"`
	WhyForYou        string `json:"WhyForYou"`
	TasteGrowthHint  string `json:"TasteGrowthHint"`
	KnowledgeGapHint string `json:"KnowledgeGapHint"`
}

func mustWriteDailyItemsJSON(t *testing.T, path string, items []featuredItemPayload) {
	t.Helper()
	payload := struct {
		Featured []featuredItemPayload `json:"Featured"`
	}{
		Featured: items,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, path, string(data))
}

func featuredItems(n int, cardType string) []featuredItemPayload {
	items := make([]featuredItemPayload, 0, n)
	for _, id := range featuredIDs(n) {
		item := featuredItemPayload{
			ID:         id,
			CardType:   cardType,
			Summary:    "Summary",
			SourceName: "Example",
			SourceURL:  "https://example.com/" + id,
		}
		if cardType == "brief" {
			item.FallbackReason = "analysis_failed"
		} else {
			item.Insight = insightPayload{
				Viewpoint:        "Viewpoint",
				WhyForYou:        "Selected for your interest in long-term strategy.",
				TasteGrowthHint:  "Keep balancing explainers with primary-source reporting.",
				KnowledgeGapHint: "Review the core market mechanics behind this topic.",
			}
		}
		items = append(items, item)
	}
	return items
}

func writeValidEditionFixture(t *testing.T, dir string, mutateIndex func(string) string, mutateArticle func(string) string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, "meta.json"), "{}")

	items := featuredItems(10, "standard")
	items[3].CardType = "brief"
	items[3].FallbackReason = "analysis_failed"
	mustWriteDailyItemsJSON(t, filepath.Join(dir, "data", "daily.json"), items)

	indexHTML := `<body data-page-kind="index" data-theme-id="editorial-ai"><main data-layout="editorial-homepage"><article data-article-id="item-a" data-card-type="standard"></article></main></body>`
	if mutateIndex != nil {
		indexHTML = mutateIndex(indexHTML)
	}
	mustWrite(t, filepath.Join(dir, "index.html"), indexHTML)

	articleHTML := `<body data-theme-id="editorial-ai"><main data-page-kind="article"><article data-feedback-surface="article"><button data-feedback-value="like"></button><button data-feedback-value="dislike"></button><button data-feedback-value="bookmark"></button></article></main></body>`
	if mutateArticle != nil {
		articleHTML = mutateArticle(articleHTML)
	}
	for _, item := range items {
		mustWrite(t, filepath.Join(dir, "articles", item.ID+".html"), articleHTML)
	}
}
