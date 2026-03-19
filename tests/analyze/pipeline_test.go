package analyze_test

import (
	"context"
	"testing"

	"github.com/nikkofu/agentic-news/internal/analyze"
	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
)

func TestAnalyzeArticle_ReturnsRequiredInsightFields(t *testing.T) {
	article := model.Article{
		Title:       "Global chip supply shifts",
		ContentText: "Semiconductor leaders announced new capacity plans with policy support.",
		Language:    "en",
	}
	p := profile.UserProfile{FocusTopics: []string{"technology", "semiconductor"}}

	insight, err := analyze.RunPipeline(context.Background(), article, p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if insight.SummaryDeep == "" || insight.Viewpoint == "" {
		t.Fatal("missing required insight fields")
	}
	if insight.PromptVersion == "" || insight.ModelName == "" {
		t.Fatal("missing traceability fields")
	}
}
