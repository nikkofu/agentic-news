package profile_test

import (
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/profile"
)

func TestApplyEvent_UpdatesTopicStyleAndCognitiveAffinity(t *testing.T) {
	p := profile.UserProfile{}
	event := profile.EventInput{
		EventType:     "like",
		TopicTags:     []string{"AI 基础设施"},
		StyleTags:     []string{"数据型"},
		CognitiveTags: []string{"结构化"},
	}

	p = profile.ApplyEvent(p, event)

	if p.TopicAffinity["AI 基础设施"] <= 0 {
		t.Fatal("expected topic affinity to increase")
	}
	if p.StyleAffinity["数据型"] <= 0 {
		t.Fatal("expected style affinity to increase")
	}
	if p.CognitiveAffinity["结构化"] <= 0 {
		t.Fatal("expected cognitive affinity to increase")
	}
	if len(p.FocusTopics) == 0 || p.FocusTopics[0] != "AI 基础设施" {
		t.Fatalf("expected focus topics to refresh, got %v", p.FocusTopics)
	}
	if len(p.PreferredStyles) == 0 || p.PreferredStyles[0] != "数据型" {
		t.Fatalf("expected preferred styles to refresh, got %v", p.PreferredStyles)
	}
	if len(p.CognitivePreferences) == 0 || p.CognitivePreferences[0] != "结构化" {
		t.Fatalf("expected cognitive preferences to refresh, got %v", p.CognitivePreferences)
	}
}

func TestApplyEvent_RecordsNegativeSignalsForDislike(t *testing.T) {
	p := profile.UserProfile{}
	event := profile.EventInput{
		EventType: "dislike",
		TopicTags: []string{"AI 基础设施"},
		ReasonTags: []string{
			"标题党",
		},
	}

	p = profile.ApplyEvent(p, event)

	if p.NegativeSignals[profileKey("topic", "AI 基础设施")] <= 0 {
		t.Fatal("expected negative signal for disliked topic")
	}
	if p.NegativeSignals[profileKey("reason", "标题党")] <= 0 {
		t.Fatal("expected negative signal for dislike reason")
	}
}

func TestApplyEvent_FeedbackLikeUpdatesAffinities(t *testing.T) {
	p := profile.UserProfile{}
	event := profile.EventInput{
		EventType:     "feedback_like",
		TopicTags:     []string{"AI 基础设施"},
		StyleTags:     []string{"数据型"},
		CognitiveTags: []string{"结构化"},
	}

	p = profile.ApplyEvent(p, event)

	if p.TopicAffinity["AI 基础设施"] <= 0 {
		t.Fatal("expected topic affinity to increase for feedback_like")
	}
	if p.StyleAffinity["数据型"] <= 0 {
		t.Fatal("expected style affinity to increase for feedback_like")
	}
	if p.CognitiveAffinity["结构化"] <= 0 {
		t.Fatal("expected cognitive affinity to increase for feedback_like")
	}
	if p.ExplicitFeedback[profileKey("topic", "AI 基础设施")] != "like" {
		t.Fatalf("expected explicit feedback like, got %q", p.ExplicitFeedback[profileKey("topic", "AI 基础设施")])
	}
}

func TestApplyEvent_FeedbackDislikeTracksNegativeSignals(t *testing.T) {
	p := profile.UserProfile{}
	event := profile.EventInput{
		EventType:  "feedback_dislike",
		TopicTags:  []string{"云平台"},
		StyleTags:  []string{"研究型"},
		ReasonTags: []string{"标题党"},
	}

	p = profile.ApplyEvent(p, event)

	if p.NegativeSignals[profileKey("topic", "云平台")] <= 0 {
		t.Fatal("expected negative signal for feedback_dislike topic")
	}
	if p.NegativeSignals[profileKey("reason", "标题党")] <= 0 {
		t.Fatal("expected negative signal for feedback_dislike reason")
	}
	if p.ExplicitFeedback[profileKey("topic", "云平台")] != "dislike" {
		t.Fatalf("expected explicit feedback dislike, got %q", p.ExplicitFeedback[profileKey("topic", "云平台")])
	}
}

func TestApplyEvent_BehaviorEventUsesDwellSignals(t *testing.T) {
	p := profile.UserProfile{}
	event := profile.EventInput{
		EventType:    "bookmark",
		TopicTags:    []string{"AI 基础设施"},
		DwellSeconds: 30,
		Bookmarked:   true,
	}

	p = profile.ApplyEvent(p, event)

	if p.BehaviorSignals["AI 基础设施"] <= 0 {
		t.Fatal("expected behavior signal for bookmark event")
	}
	if p.TopicAffinity["AI 基础设施"] <= 0 {
		t.Fatal("expected topic affinity to increase from dwell behavior")
	}
	if _, ok := p.ExplicitFeedback[profileKey("topic", "AI 基础设施")]; ok {
		t.Fatal("expected no explicit feedback for behavior-only event")
	}
}

func TestApplyEvent_NormalizesWhitespaceTagsAndSources(t *testing.T) {
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType:  "feedback_like",
		TopicTags:  []string{"  AI 基础设施 "},
		SourceName: " 路透社 ",
	})

	if p.TopicAffinity["AI 基础设施"] <= 0 {
		t.Fatalf("expected normalized topic affinity, got %v", p.TopicAffinity)
	}
	if p.SourceAffinity["路透社"] <= 0 {
		t.Fatalf("expected normalized source affinity, got %v", p.SourceAffinity)
	}
	if _, ok := p.TopicAffinity["  AI 基础设施 "]; ok {
		t.Fatalf("expected trimmed topic key, got %v", p.TopicAffinity)
	}
	if _, ok := p.SourceAffinity[" 路透社 "]; ok {
		t.Fatalf("expected trimmed source key, got %v", p.SourceAffinity)
	}
}

func TestApplyEvent_WhitespaceSourceDoesNotSkipUpdates(t *testing.T) {
	eventTime := time.Date(2026, 3, 19, 18, 0, 0, 0, time.UTC)
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType:     "feedback_like",
		TopicTags:     []string{" AI 基础设施 "},
		StyleTags:     []string{" 数据型 "},
		CognitiveTags: []string{" 结构化 "},
		SourceName:    "   ",
		Timestamp:     eventTime,
	})

	if p.TopicAffinity["AI 基础设施"] <= 0 {
		t.Fatalf("expected topic affinity update, got %v", p.TopicAffinity)
	}
	if len(p.RecentFeedbackSummary) != 1 {
		t.Fatalf("expected summary to be recorded, got %v", p.RecentFeedbackSummary)
	}
	if !p.LastUpdatedAt.Equal(eventTime) {
		t.Fatalf("expected last updated to be set, got %v", p.LastUpdatedAt)
	}
}

func TestApplyEvent_SummaryNormalizesTags(t *testing.T) {
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType:     "feedback_like",
		TopicTags:     []string{"  AI 基础设施  ", "   "},
		StyleTags:     []string{" 数据型 "},
		CognitiveTags: []string{" 结构化 ", ""},
	})

	if len(p.RecentFeedbackSummary) != 1 {
		t.Fatalf("expected summary to be recorded, got %v", p.RecentFeedbackSummary)
	}
	summary := p.RecentFeedbackSummary[0]
	if strings.Contains(summary, "  AI 基础设施  ") || strings.Contains(summary, "   ") {
		t.Fatalf("expected summary to use trimmed tags, got %q", summary)
	}
	if strings.Contains(summary, " +  +") || strings.Contains(summary, "  + ") {
		t.Fatalf("expected summary to avoid blank fragments, got %q", summary)
	}
	if !strings.Contains(summary, "AI 基础设施") || !strings.Contains(summary, "数据型") || !strings.Contains(summary, "结构化") {
		t.Fatalf("expected summary to include normalized tags, got %q", summary)
	}
}

func TestApplyEvent_WeakEventDoesNotCreateSummary(t *testing.T) {
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "article_view",
		TopicTags: []string{"AI 基础设施"},
	})

	if len(p.RecentFeedbackSummary) != 0 {
		t.Fatalf("expected no summary for weak event, got %v", p.RecentFeedbackSummary)
	}
}

func TestApplyEvent_TimestampDoesNotMoveBackward(t *testing.T) {
	newer := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	older := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "feedback_like",
		TopicTags: []string{"AI 基础设施"},
		Timestamp: newer,
	})
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "feedback_like",
		TopicTags: []string{"云平台"},
		Timestamp: older,
	})

	if !p.LastUpdatedAt.Equal(newer) {
		t.Fatalf("expected last updated to remain at newest timestamp, got %v", p.LastUpdatedAt)
	}
}

func TestApplyEvent_OlderEventDoesNotDisplaceRecentSummary(t *testing.T) {
	newer := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	older := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "feedback_like",
		TopicTags: []string{"新事件"},
		Timestamp: newer,
	})
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "feedback_like",
		TopicTags: []string{"旧事件"},
		Timestamp: older,
	})

	if len(p.RecentFeedbackSummary) != 1 {
		t.Fatalf("expected only newest summary, got %v", p.RecentFeedbackSummary)
	}
	if !strings.Contains(p.RecentFeedbackSummary[0], "新事件") {
		t.Fatalf("expected summary to reflect newest event, got %v", p.RecentFeedbackSummary)
	}
	if strings.Contains(p.RecentFeedbackSummary[0], "旧事件") {
		t.Fatalf("expected older event not to displace summary, got %v", p.RecentFeedbackSummary)
	}
}

func TestApplyEvent_PreservesLegacySlicePreferences(t *testing.T) {
	p := profile.UserProfile{
		FocusTopics:          []string{"AI 基础设施"},
		PreferredStyles:      []string{"数据型"},
		CognitivePreferences: []string{"结构化"},
	}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "feedback_like",
		TopicTags: []string{"云平台"},
		StyleTags: []string{"研究型"},
		CognitiveTags: []string{
			"对照型",
		},
	})

	if !containsString(p.FocusTopics, "AI 基础设施") {
		t.Fatalf("expected legacy focus topic preserved, got %v", p.FocusTopics)
	}
	if !containsString(p.PreferredStyles, "数据型") {
		t.Fatalf("expected legacy preferred style preserved, got %v", p.PreferredStyles)
	}
	if !containsString(p.CognitivePreferences, "结构化") {
		t.Fatalf("expected legacy cognitive preference preserved, got %v", p.CognitivePreferences)
	}
	if !containsString(p.FocusTopics, "云平台") || !containsString(p.PreferredStyles, "研究型") || !containsString(p.CognitivePreferences, "对照型") {
		t.Fatalf("expected new tags to refine preferences, got %v / %v / %v", p.FocusTopics, p.PreferredStyles, p.CognitivePreferences)
	}
}

func TestBuildLearningSnapshot_UsesLegacySlicePreferences(t *testing.T) {
	p := profile.UserProfile{
		FocusTopics:          []string{"AI 基础设施"},
		PreferredStyles:      []string{"数据型"},
		CognitivePreferences: []string{"结构化"},
	}

	snapshot := profile.BuildLearningSnapshot(p)

	if !strings.Contains(snapshot.TasteGrowthHint, "AI 基础设施") {
		t.Fatalf("expected taste hint to reflect legacy topic, got %q", snapshot.TasteGrowthHint)
	}
	if !strings.Contains(snapshot.TasteGrowthHint, "数据型") {
		t.Fatalf("expected taste hint to reflect legacy style, got %q", snapshot.TasteGrowthHint)
	}
	if !strings.Contains(snapshot.LearningTracks[0], "AI 基础设施") {
		t.Fatalf("expected learning tracks to reflect legacy topic, got %v", snapshot.LearningTracks)
	}
}

func TestBuildLearningSnapshot_DoesNotMutateEmptyAffinityMaps(t *testing.T) {
	p := profile.UserProfile{
		FocusTopics:          []string{"AI 基础设施"},
		PreferredStyles:      []string{"数据型"},
		CognitivePreferences: []string{"结构化"},
		TopicAffinity:        map[string]float64{},
		StyleAffinity:        map[string]float64{},
		CognitiveAffinity:    map[string]float64{},
	}

	_ = profile.BuildLearningSnapshot(p)

	if len(p.TopicAffinity) != 0 || len(p.StyleAffinity) != 0 || len(p.CognitiveAffinity) != 0 {
		t.Fatalf("expected snapshot to avoid mutating affinity maps, got %v / %v / %v", p.TopicAffinity, p.StyleAffinity, p.CognitiveAffinity)
	}
}

func TestTopKeys_IgnoresNonPositiveScores(t *testing.T) {
	scores := map[string]float64{
		"积极": 2,
		"中性": 0,
		"消极": -1,
	}

	top := profile.TopKeys(scores, 3)

	if len(top) != 1 || top[0] != "积极" {
		t.Fatalf("expected only positive keys, got %v", top)
	}
}

func TestApplyEvent_SourceAffinityWeightsExplicitAndBehavior(t *testing.T) {
	p := profile.UserProfile{}
	explicit := profile.EventInput{
		EventType:  "feedback_like",
		SourceName: "路透社",
	}
	behavior := profile.EventInput{
		EventType:    "bookmark",
		SourceName:   "彭博",
		DwellSeconds: 30,
	}
	unknown := profile.EventInput{
		EventType:  "mystery_event",
		SourceName: "未知来源",
	}

	p = profile.ApplyEvent(p, explicit)
	p = profile.ApplyEvent(p, behavior)
	p = profile.ApplyEvent(p, unknown)

	if p.SourceAffinity["路透社"] <= p.SourceAffinity["彭博"] {
		t.Fatalf("expected explicit source affinity stronger than behavior, got %v vs %v", p.SourceAffinity["路透社"], p.SourceAffinity["彭博"])
	}
	if p.SourceAffinity["彭博"] <= 0 {
		t.Fatalf("expected behavior source affinity positive, got %v", p.SourceAffinity["彭博"])
	}
	if value, ok := p.SourceAffinity["未知来源"]; ok && value != 0 {
		t.Fatalf("expected unknown event to avoid drifting source affinity, got %v", value)
	}
}

func TestBuildLearningSnapshot_ReturnsTasteAndKnowledgeHints(t *testing.T) {
	updatedAt := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	p := profile.UserProfile{
		TopicAffinity: map[string]float64{
			"AI 基础设施": 6,
			"芯片供应链":  3,
		},
		StyleAffinity: map[string]float64{
			"数据型": 4,
		},
		CognitiveAffinity: map[string]float64{
			"结构化": 10,
			"反思型": 2,
		},
		SourceAffinity: map[string]float64{
			"路透社": 5,
		},
		LastUpdatedAt: updatedAt,
	}

	snapshot := profile.BuildLearningSnapshot(p)

	if snapshot.TasteGrowthHint == "" || snapshot.KnowledgeGapHint == "" {
		t.Fatal("expected taste and knowledge hints to be non-empty")
	}
	if snapshot.TodayPlan == "" {
		t.Fatal("expected today plan to be non-empty")
	}
	if !strings.Contains(snapshot.TodayPlan, "路透社") {
		t.Fatalf("expected today plan to mention source affinity, got %q", snapshot.TodayPlan)
	}
	if strings.Contains(snapshot.TodayPlan, "数据型来源") {
		t.Fatalf("expected today plan to use source affinity, got %q", snapshot.TodayPlan)
	}
	if len(snapshot.LearningTracks) == 0 {
		t.Fatal("expected learning tracks to be non-empty")
	}
	if !snapshot.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated at to match profile, got %v", snapshot.UpdatedAt)
	}
	hasBalance := strings.Contains(snapshot.TasteGrowthHint, "平衡") ||
		strings.Contains(snapshot.KnowledgeGapHint, "平衡")
	if !hasBalance {
		for _, track := range snapshot.LearningTracks {
			if strings.Contains(track, "平衡") {
				hasBalance = true
				break
			}
		}
	}
	if !hasBalance {
		t.Fatal("expected balancing guidance when one cognitive lens dominates")
	}
}

func TestBuildLearningSnapshot_KnowledgeGapAvoidsDislikedTopics(t *testing.T) {
	p := profile.UserProfile{
		TopicAffinity: map[string]float64{
			"AI 基础设施": 2,
			"低质话题":    -3,
		},
	}

	snapshot := profile.BuildLearningSnapshot(p)

	if strings.Contains(snapshot.KnowledgeGapHint, "低质话题") {
		t.Fatalf("expected knowledge gap to avoid disliked topic, got %q", snapshot.KnowledgeGapHint)
	}
	if !strings.Contains(snapshot.KnowledgeGapHint, "AI 基础设施") {
		t.Fatalf("expected knowledge gap to use positive topic, got %q", snapshot.KnowledgeGapHint)
	}
}

func TestBuildLearningSnapshot_KnowledgeGapSkipsExplicitDislike(t *testing.T) {
	p := profile.UserProfile{
		TopicAffinity: map[string]float64{
			"AI 基础设施": 2,
			"云平台":    1,
		},
		ExplicitFeedback: map[string]string{
			profileKey("topic", "云平台"): "dislike",
		},
		NegativeSignals: map[string]float64{
			profileKey("topic", "云平台"): 1,
		},
	}

	snapshot := profile.BuildLearningSnapshot(p)

	if strings.Contains(snapshot.KnowledgeGapHint, "云平台") {
		t.Fatalf("expected knowledge gap to avoid disliked topic, got %q", snapshot.KnowledgeGapHint)
	}
	if !strings.Contains(snapshot.KnowledgeGapHint, "AI 基础设施") {
		t.Fatalf("expected knowledge gap to use acceptable topic, got %q", snapshot.KnowledgeGapHint)
	}
}

func TestBuildLearningSnapshot_KnowledgeGapIgnoresStyleLabelDislike(t *testing.T) {
	p := profile.UserProfile{}
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType: "feedback_like",
		TopicTags: []string{"重叠标签"},
	})
	p = profile.ApplyEvent(p, profile.EventInput{
		EventType:  "feedback_dislike",
		StyleTags:  []string{"重叠标签"},
		ReasonTags: []string{"重叠标签"},
	})

	snapshot := profile.BuildLearningSnapshot(p)

	if !strings.Contains(snapshot.KnowledgeGapHint, "重叠标签") {
		t.Fatalf("expected knowledge gap to allow topic label despite style/reason dislike, got %q", snapshot.KnowledgeGapHint)
	}
}

func profileKey(namespace, tag string) string {
	return "profile:" + namespace + ":" + tag
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestApplyEvent_BuildsRecentFeedbackSummary(t *testing.T) {
	p := profile.UserProfile{}
	events := []profile.EventInput{
		{
			EventType:     "like",
			TopicTags:     []string{"AI 基础设施"},
			StyleTags:     []string{"数据型"},
			CognitiveTags: []string{"结构化"},
		},
		{
			EventType:     "like",
			TopicTags:     []string{"芯片供应链"},
			StyleTags:     []string{"研究型"},
			CognitiveTags: []string{"对照型"},
		},
		{
			EventType:     "like",
			TopicTags:     []string{"AI 基础设施"},
			StyleTags:     []string{"简报型"},
			CognitiveTags: []string{"故事型"},
		},
		{
			EventType:     "like",
			TopicTags:     []string{"云平台"},
			StyleTags:     []string{"数据型"},
			CognitiveTags: []string{"结构化"},
		},
	}

	for _, event := range events {
		p = profile.ApplyEvent(p, event)
	}

	if len(p.RecentFeedbackSummary) != 3 {
		t.Fatalf("expected 3 summary entries, got %d", len(p.RecentFeedbackSummary))
	}
	summary := p.RecentFeedbackSummary[len(p.RecentFeedbackSummary)-1]
	if !strings.Contains(summary, "云平台") || !strings.Contains(summary, "数据型") {
		t.Fatalf("expected summary to include tags, got %q", summary)
	}
	if !strings.Contains(summary, "强化") {
		t.Fatalf("expected summary to describe positive feedback, got %q", summary)
	}
}

func TestApplyEvent_DoesNotMutateInputProfile(t *testing.T) {
	originalTime := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	original := profile.UserProfile{
		FocusTopics:           []string{"AI 基础设施"},
		TopicAffinity:         map[string]float64{"AI 基础设施": 2},
		RecentFeedbackSummary: []string{"你最近强化了 AI 基础设施 内容。"},
		LastUpdatedAt:         originalTime,
	}
	event := profile.EventInput{
		EventType:  "like",
		TopicTags:  []string{"云平台"},
		Timestamp:  time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC),
		StyleTags:  []string{"数据型"},
		ReasonTags: []string{"简报型"},
	}

	_ = profile.ApplyEvent(original, event)

	if original.TopicAffinity["AI 基础设施"] != 2 {
		t.Fatalf("expected original topic affinity untouched, got %v", original.TopicAffinity["AI 基础设施"])
	}
	if len(original.RecentFeedbackSummary) != 1 {
		t.Fatalf("expected original summaries untouched, got %v", original.RecentFeedbackSummary)
	}
	if len(original.FocusTopics) != 1 || original.FocusTopics[0] != "AI 基础设施" {
		t.Fatalf("expected original focus topics untouched, got %v", original.FocusTopics)
	}
	if !original.LastUpdatedAt.Equal(originalTime) {
		t.Fatalf("expected original last updated unchanged, got %v", original.LastUpdatedAt)
	}
}

func TestTimestamps_AreDeterministic(t *testing.T) {
	eventTime := time.Date(2026, 3, 19, 10, 30, 0, 0, time.UTC)
	p := profile.UserProfile{}
	event := profile.EventInput{
		EventType: "like",
		TopicTags: []string{"AI 基础设施"},
		Timestamp: eventTime,
	}

	p = profile.ApplyEvent(p, event)
	if !p.LastUpdatedAt.Equal(eventTime) {
		t.Fatalf("expected last updated to match event time, got %v", p.LastUpdatedAt)
	}

	p.LastUpdatedAt = eventTime.Add(2 * time.Hour)
	snapshot := profile.BuildLearningSnapshot(p)
	if !snapshot.UpdatedAt.Equal(p.LastUpdatedAt) {
		t.Fatalf("expected snapshot updated at to match profile, got %v", snapshot.UpdatedAt)
	}
}
