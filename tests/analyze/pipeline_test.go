package analyze_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nikkofu/agentic-news/internal/analyze"
	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
)

func TestAnalyzeArticle_ReturnsRequiredInsightFields(t *testing.T) {
	article := model.Article{
		Title:        "Global chip supply shifts",
		ContentText:  "Semiconductor leaders announced new capacity plans with policy support.",
		CanonicalURL: "https://example.com/chips",
		Language:     "en",
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

func TestRunPipeline_RejectsMissingTraceabilityFields(t *testing.T) {
	article := model.Article{
		Title:        "Global chip supply shifts",
		ContentText:  "Semiconductor leaders announced new capacity plans with policy support.",
		CanonicalURL: "",
		Language:     "en",
	}

	_, err := analyze.RunPipeline(context.Background(), article, profile.UserProfile{})
	if err == nil {
		t.Fatal("expected validation error for missing traceability fields")
	}
	if !strings.Contains(err.Error(), "source_refs") {
		t.Fatalf("expected source_refs validation error, got %v", err)
	}
}

func TestAnalyzeRunPipeline_BuildsWhyForYouFromExpandedProfile(t *testing.T) {
	article := model.Article{
		Title:        "Chip foundries revisit expansion plans",
		ContentText:  "Foundries are revisiting expansion plans while enterprise buyers compare capex efficiency and supply resilience.",
		CanonicalURL: "https://example.com/foundries",
		Language:     "en",
	}
	p := profile.UserProfile{
		FocusTopics:           []string{"semiconductors", "industrial policy"},
		PreferredStyles:       []string{"data-driven"},
		CognitivePreferences:  []string{"systems thinking"},
		RecentFeedbackSummary: []string{"You recently reinforced semiconductors + data-driven + systems thinking content."},
		TopicAffinity: map[string]float64{
			"semiconductors":    8,
			"industrial policy": 4,
		},
		StyleAffinity: map[string]float64{
			"data-driven": 6,
		},
		CognitiveAffinity: map[string]float64{
			"systems thinking": 5,
		},
	}

	insight, err := analyze.RunPipeline(context.Background(), article, p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(insight.WhyForYou, "semiconductors") {
		t.Fatalf("expected why_for_you to include focus topic, got %q", insight.WhyForYou)
	}
	if !strings.Contains(insight.WhyForYou, "data-driven") {
		t.Fatalf("expected why_for_you to include preferred style, got %q", insight.WhyForYou)
	}
	if !strings.Contains(insight.WhyForYou, "systems thinking") {
		t.Fatalf("expected why_for_you to include cognitive preference, got %q", insight.WhyForYou)
	}
	if !strings.Contains(insight.TasteGrowthHint, "semiconductors") {
		t.Fatalf("expected taste hint to reflect profile snapshot, got %q", insight.TasteGrowthHint)
	}
	if !strings.Contains(insight.KnowledgeGapHint, "industrial policy") {
		t.Fatalf("expected knowledge gap hint to reflect weakest positive topic, got %q", insight.KnowledgeGapHint)
	}
}
