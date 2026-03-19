package analyze

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
)

func RunPipeline(ctx context.Context, a model.Article, p profile.UserProfile) (model.Insight, error) {
	if strings.TrimSpace(a.ContentText) == "" {
		return model.Insight{}, errors.New("article content is required")
	}

	_, _ = RenderPrompt(filepath.Join("prompts", "extract_facts.tmpl"), a)
	_, _ = RenderPrompt(filepath.Join("prompts", "deep_analysis.tmpl"), a)
	_, _ = RenderPrompt(filepath.Join("prompts", "personal_advisor.tmpl"), map[string]any{
		"Title":       a.Title,
		"ContentText": a.ContentText,
		"FocusTopics": strings.Join(p.FocusTopics, ", "),
	})

	summaryBrief := oneLine(a.ContentText, 120)
	summaryDeep := oneLine(a.ContentText, 220)
	viewpoint := "This development may reshape competitive positioning and warrants follow-up on execution signals."
	whyForYou := buildWhyForYou(p)

	insight := model.Insight{
		SummaryBrief:       summaryBrief,
		SummaryDeep:        summaryDeep,
		KeyPoints:          []string{summaryBrief},
		Viewpoint:          viewpoint,
		OpportunityRisk:    "Opportunity: early positioning; Risk: policy and execution uncertainty.",
		ContrarianTake:     "Consensus may overprice headline momentum before fundamentals confirm.",
		LearningSuggestion: []string{"Track policy-to-industry transmission", "Compare announced capacity vs demand"},
		Confidence:         78,
		WhyForYou:          whyForYou,
		TasteGrowthHint:    "Prioritize sources with verifiable data over commentary-only takes.",
		KnowledgeGapHint:   "Review semiconductor capex cycle basics for better signal interpretation.",
		ModelName:          "quality-first-simulated",
		ModelVersion:       "v0",
		PromptVersion:      "2026-03-19",
		SourceRefs:         []string{a.CanonicalURL},
		EvidenceSnippets:   []string{summaryBrief},
		GeneratedAt:        time.Now(),
	}

	if err := validateInsight(insight); err != nil {
		return model.Insight{}, err
	}

	return insight, nil
}

func validateInsight(i model.Insight) error {
	if strings.TrimSpace(i.SummaryDeep) == "" {
		return fmt.Errorf("summary_deep is required")
	}
	if strings.TrimSpace(i.Viewpoint) == "" {
		return fmt.Errorf("viewpoint is required")
	}
	if strings.TrimSpace(i.ModelName) == "" || strings.TrimSpace(i.PromptVersion) == "" {
		return fmt.Errorf("traceability fields are required")
	}
	return nil
}

func oneLine(text string, max int) string {
	flat := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if len(flat) <= max {
		return flat
	}
	return strings.TrimSpace(flat[:max])
}

func buildWhyForYou(p profile.UserProfile) string {
	if len(p.FocusTopics) == 0 {
		return "Selected for broad strategic relevance and signal density."
	}
	return fmt.Sprintf("Aligned with your focus on %s and high-impact trend tracking.", strings.Join(p.FocusTopics, ", "))
}
