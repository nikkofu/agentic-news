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
	learningSnapshot := profile.BuildLearningSnapshot(p)

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
		TasteGrowthHint:    firstNonEmpty(learningSnapshot.TasteGrowthHint, "Prioritize sources with verifiable data over commentary-only takes."),
		KnowledgeGapHint:   firstNonEmpty(learningSnapshot.KnowledgeGapHint, "Review semiconductor capex cycle basics for better signal interpretation."),
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
	if strings.TrimSpace(i.SummaryBrief) == "" {
		return fmt.Errorf("summary_brief is required")
	}
	if strings.TrimSpace(i.SummaryDeep) == "" {
		return fmt.Errorf("summary_deep is required")
	}
	if strings.TrimSpace(i.Viewpoint) == "" {
		return fmt.Errorf("viewpoint is required")
	}
	if i.Confidence < 0 || i.Confidence > 100 {
		return fmt.Errorf("confidence must be between 0 and 100")
	}
	if strings.TrimSpace(i.ModelName) == "" || strings.TrimSpace(i.ModelVersion) == "" || strings.TrimSpace(i.PromptVersion) == "" {
		return fmt.Errorf("traceability fields are required")
	}
	if len(i.SourceRefs) == 0 {
		return fmt.Errorf("source_refs are required")
	}
	for _, ref := range i.SourceRefs {
		if strings.TrimSpace(ref) == "" {
			return fmt.Errorf("source_refs must not be empty")
		}
	}
	if len(i.EvidenceSnippets) == 0 {
		return fmt.Errorf("evidence_snippets are required")
	}
	for _, snippet := range i.EvidenceSnippets {
		if strings.TrimSpace(snippet) == "" {
			return fmt.Errorf("evidence_snippets must not be empty")
		}
	}
	if i.GeneratedAt.IsZero() {
		return fmt.Errorf("generated_at is required")
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
	topics := normalized(p.FocusTopics)
	styles := normalized(p.PreferredStyles)
	cognition := normalized(p.CognitivePreferences)

	if len(topics) == 0 && len(styles) == 0 && len(cognition) == 0 {
		return "Selected for broad strategic relevance and signal density."
	}

	parts := make([]string, 0, 4)
	if len(topics) > 0 {
		parts = append(parts, fmt.Sprintf("Aligned with your focus on %s.", strings.Join(topics, ", ")))
	}
	if len(styles) > 0 {
		parts = append(parts, fmt.Sprintf("Format match: %s.", strings.Join(styles, ", ")))
	}
	if len(cognition) > 0 {
		parts = append(parts, fmt.Sprintf("Thinking style match: %s.", strings.Join(cognition, ", ")))
	}
	if len(p.RecentFeedbackSummary) > 0 {
		parts = append(parts, firstNonEmpty(p.RecentFeedbackSummary[len(p.RecentFeedbackSummary)-1]))
	}
	return strings.Join(parts, " ")
}

func normalized(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
