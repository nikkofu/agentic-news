package profile

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type UserProfile struct {
	FocusTopics           []string
	PreferredStyles       []string
	CognitivePreferences  []string
	ExplicitFeedback      map[string]string
	BehaviorSignals       map[string]float64
	TopicAffinity         map[string]float64
	StyleAffinity         map[string]float64
	CognitiveAffinity     map[string]float64
	SourceAffinity        map[string]float64
	NegativeSignals       map[string]float64
	RecentFeedbackSummary []string
	LastUpdatedAt         time.Time
}

type EventInput struct {
	EventType     string
	TopicTags     []string
	StyleTags     []string
	CognitiveTags []string
	SourceName    string
	DwellSeconds  float64
	Bookmarked    bool
	ReasonTags    []string
	Timestamp     time.Time
}

type LearningSnapshot struct {
	TasteGrowthHint  string
	KnowledgeGapHint string
	TodayPlan        string
	LearningTracks   []string
	UpdatedAt        time.Time
}

const (
	profileKeyPrefix       = "profile:"
	profileTopicNamespace  = "topic"
	profileStyleNamespace  = "style"
	profileCognitiveNS     = "cognitive"
	profileReasonNamespace = "reason"
)

func ApplyEvent(p UserProfile, event EventInput) UserProfile {
	p = cloneProfile(p)
	p = ensureProfileMaps(p)
	p = seedLegacyAffinity(p)

	explicitBoost := 6.0
	negativeBoost := 4.0
	dwellCap := 180.0
	dwellScore := math.Min(event.DwellSeconds, dwellCap) / 60.0
	if event.Bookmarked {
		dwellScore += 2.0
	}

	normalizedEvent := strings.ToLower(strings.TrimSpace(event.EventType))
	isDislike := normalizedEvent == "dislike" || normalizedEvent == "feedback_dislike"
	isLike := normalizedEvent == "like" || normalizedEvent == "feedback_like"

	if isLike {
		applyAffinity(p.TopicAffinity, event.TopicTags, explicitBoost)
		applyAffinity(p.StyleAffinity, event.StyleTags, explicitBoost)
		applyAffinity(p.CognitiveAffinity, event.CognitiveTags, explicitBoost)
		applyFeedback(p.ExplicitFeedback, event.TopicTags, "like", profileTopicNamespace)
		applyFeedback(p.ExplicitFeedback, event.StyleTags, "like", profileStyleNamespace)
		applyFeedback(p.ExplicitFeedback, event.CognitiveTags, "like", profileCognitiveNS)
	} else if isDislike {
		applyAffinity(p.TopicAffinity, event.TopicTags, -explicitBoost)
		applyAffinity(p.StyleAffinity, event.StyleTags, -explicitBoost)
		applyAffinity(p.CognitiveAffinity, event.CognitiveTags, -explicitBoost)
		applyFeedback(p.ExplicitFeedback, event.TopicTags, "dislike", profileTopicNamespace)
		applyFeedback(p.ExplicitFeedback, event.StyleTags, "dislike", profileStyleNamespace)
		applyFeedback(p.ExplicitFeedback, event.CognitiveTags, "dislike", profileCognitiveNS)
		applyNegative(p.NegativeSignals, event.TopicTags, negativeBoost, profileTopicNamespace)
		applyNegative(p.NegativeSignals, event.StyleTags, negativeBoost, profileStyleNamespace)
		applyNegative(p.NegativeSignals, event.CognitiveTags, negativeBoost, profileCognitiveNS)
		applyNegative(p.NegativeSignals, event.ReasonTags, negativeBoost, profileReasonNamespace)
	}

	if dwellScore > 0 && !isDislike {
		applyAffinity(p.TopicAffinity, event.TopicTags, dwellScore)
		applyAffinity(p.StyleAffinity, event.StyleTags, dwellScore*0.7)
		applyAffinity(p.CognitiveAffinity, event.CognitiveTags, dwellScore*0.6)
		applyBehavior(p.BehaviorSignals, event.TopicTags, dwellScore)
	}

	sourceKey := strings.TrimSpace(event.SourceName)
	if sourceKey != "" {
		switch {
		case isLike:
			p.SourceAffinity[sourceKey] += explicitBoost
		case isDislike:
			p.SourceAffinity[sourceKey] -= negativeBoost
		case dwellScore > 0:
			p.SourceAffinity[sourceKey] += dwellScore
		}
	}

	p.FocusTopics = TopKeys(p.TopicAffinity, 3)
	p.PreferredStyles = TopKeys(p.StyleAffinity, 3)
	p.CognitivePreferences = TopKeys(p.CognitiveAffinity, 3)

	shouldSummarize := isLike || isDislike || event.Bookmarked
	allowSummary := shouldSummarize
	if allowSummary && !event.Timestamp.IsZero() && !p.LastUpdatedAt.IsZero() && event.Timestamp.Before(p.LastUpdatedAt) {
		allowSummary = false
	}
	if allowSummary {
		summary := buildSummaryLine(event, isDislike)
		if summary != "" {
			p.RecentFeedbackSummary = append(p.RecentFeedbackSummary, summary)
			if len(p.RecentFeedbackSummary) > 3 {
				p.RecentFeedbackSummary = p.RecentFeedbackSummary[len(p.RecentFeedbackSummary)-3:]
			}
		}
	}

	if !event.Timestamp.IsZero() && (p.LastUpdatedAt.IsZero() || !event.Timestamp.Before(p.LastUpdatedAt)) {
		p.LastUpdatedAt = event.Timestamp
	}
	return p
}

func BuildLearningSnapshot(p UserProfile) LearningSnapshot {
	p = ensureProfileMaps(p)
	p = cloneProfile(p)
	p = seedLegacyAffinity(p)
	topTopics := TopKeys(p.TopicAffinity, 2)
	topStyles := TopKeys(p.StyleAffinity, 1)
	topCognition := TopKeys(p.CognitiveAffinity, 2)
	topSources := TopKeys(p.SourceAffinity, 1)

	topicHint := "宏观趋势"
	if len(topTopics) > 0 {
		topicHint = strings.Join(topTopics, "、")
	}
	styleHint := "结构化"
	if len(topStyles) > 0 {
		styleHint = strings.Join(topStyles, "、")
	}
	sourceHint := "权威"
	if len(topSources) > 0 {
		sourceHint = strings.Join(topSources, "、")
	}

	tasteHint := fmt.Sprintf("继续深挖 %s，并优先吸收%s内容。", topicHint, styleHint)

	gapHint := "建议补充产业链基础与关键指标的理解。"
	if gapTopic := weakestPositiveKey(p.TopicAffinity, p.ExplicitFeedback, p.NegativeSignals); gapTopic != "" {
		gapHint = fmt.Sprintf("建议补充了解 %s 的基础脉络。", gapTopic)
	}

	learningTracks := []string{
		fmt.Sprintf("追踪 %s 的关键指标与节奏变化", topicHint),
		fmt.Sprintf("保持%s内容与案例解读的输入", styleHint),
	}

	if shouldBalance(topCognition, p.CognitiveAffinity) {
		learningTracks = append(learningTracks, "平衡视角：加入对照型或反思型内容")
		tasteHint = tasteHint + " 同时注意平衡不同认知视角。"
	}

	todayPlan := fmt.Sprintf("今天优先阅读：%s；重点关注%s来源的信号。", topicHint, sourceHint)

	return LearningSnapshot{
		TasteGrowthHint:  tasteHint,
		KnowledgeGapHint: gapHint,
		TodayPlan:        todayPlan,
		LearningTracks:   learningTracks,
		UpdatedAt:        p.LastUpdatedAt,
	}
}

func TopKeys(scores map[string]float64, n int) []string {
	if n <= 0 || len(scores) == 0 {
		return nil
	}
	type pair struct {
		key   string
		score float64
	}
	pairs := make([]pair, 0, len(scores))
	for key, score := range scores {
		if score <= 0 {
			continue
		}
		pairs = append(pairs, pair{key: key, score: score})
	}
	if len(pairs) == 0 {
		return nil
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return pairs[i].key < pairs[j].key
		}
		return pairs[i].score > pairs[j].score
	})
	limit := n
	if len(pairs) < limit {
		limit = len(pairs)
	}
	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, pairs[i].key)
	}
	return result
}

func ensureProfileMaps(p UserProfile) UserProfile {
	if p.ExplicitFeedback == nil {
		p.ExplicitFeedback = map[string]string{}
	}
	if p.BehaviorSignals == nil {
		p.BehaviorSignals = map[string]float64{}
	}
	if p.TopicAffinity == nil {
		p.TopicAffinity = map[string]float64{}
	}
	if p.StyleAffinity == nil {
		p.StyleAffinity = map[string]float64{}
	}
	if p.CognitiveAffinity == nil {
		p.CognitiveAffinity = map[string]float64{}
	}
	if p.SourceAffinity == nil {
		p.SourceAffinity = map[string]float64{}
	}
	if p.NegativeSignals == nil {
		p.NegativeSignals = map[string]float64{}
	}
	if p.RecentFeedbackSummary == nil {
		p.RecentFeedbackSummary = []string{}
	}
	return p
}

func cloneProfile(p UserProfile) UserProfile {
	clone := p
	clone.FocusTopics = cloneStringSlice(p.FocusTopics)
	clone.PreferredStyles = cloneStringSlice(p.PreferredStyles)
	clone.CognitivePreferences = cloneStringSlice(p.CognitivePreferences)
	clone.RecentFeedbackSummary = cloneStringSlice(p.RecentFeedbackSummary)
	clone.ExplicitFeedback = copyStringMap(p.ExplicitFeedback)
	clone.BehaviorSignals = copyFloatMap(p.BehaviorSignals)
	clone.TopicAffinity = copyFloatMap(p.TopicAffinity)
	clone.StyleAffinity = copyFloatMap(p.StyleAffinity)
	clone.CognitiveAffinity = copyFloatMap(p.CognitiveAffinity)
	clone.SourceAffinity = copyFloatMap(p.SourceAffinity)
	clone.NegativeSignals = copyFloatMap(p.NegativeSignals)
	return clone
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func copyStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func copyFloatMap(values map[string]float64) map[string]float64 {
	if values == nil {
		return nil
	}
	clone := make(map[string]float64, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func applyAffinity(target map[string]float64, tags []string, delta float64) {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		target[trimmed] += delta
	}
}

func applyFeedback(target map[string]string, tags []string, sentiment string, namespace string) {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		target[profileKey(namespace, trimmed)] = sentiment
	}
}

func applyBehavior(target map[string]float64, tags []string, delta float64) {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		target[trimmed] += delta
	}
}

func applyNegative(target map[string]float64, tags []string, delta float64, namespace string) {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		target[profileKey(namespace, trimmed)] += delta
	}
}

func buildSummaryLine(event EventInput, isDislike bool) string {
	parts := []string{}
	if tags := normalizedTags(event.TopicTags); len(tags) > 0 {
		parts = append(parts, strings.Join(tags, " + "))
	}
	if tags := normalizedTags(event.StyleTags); len(tags) > 0 {
		parts = append(parts, strings.Join(tags, " + "))
	}
	if tags := normalizedTags(event.CognitiveTags); len(tags) > 0 {
		parts = append(parts, strings.Join(tags, " + "))
	}
	if len(parts) == 0 {
		return ""
	}
	body := strings.Join(parts, " + ")
	if isDislike {
		return fmt.Sprintf("你最近降低了对 %s 的偏好。", body)
	}
	return fmt.Sprintf("你最近强化了 %s 内容。", body)
}

func normalizedTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func seedLegacyAffinity(p UserProfile) UserProfile {
	if len(p.TopicAffinity) == 0 && len(p.FocusTopics) > 0 {
		for _, tag := range normalizedTags(p.FocusTopics) {
			p.TopicAffinity[tag] = 1.0
		}
	}
	if len(p.StyleAffinity) == 0 && len(p.PreferredStyles) > 0 {
		for _, tag := range normalizedTags(p.PreferredStyles) {
			p.StyleAffinity[tag] = 1.0
		}
	}
	if len(p.CognitiveAffinity) == 0 && len(p.CognitivePreferences) > 0 {
		for _, tag := range normalizedTags(p.CognitivePreferences) {
			p.CognitiveAffinity[tag] = 1.0
		}
	}
	return p
}

func weakestPositiveKey(scores map[string]float64, explicit map[string]string, negative map[string]float64) string {
	if len(scores) == 0 {
		return ""
	}
	type pair struct {
		key   string
		score float64
	}
	pairs := make([]pair, 0, len(scores))
	for key, score := range scores {
		if score <= 0 {
			continue
		}
		topicKey := profileKey(profileTopicNamespace, key)
		if explicit != nil && explicit[topicKey] == "dislike" {
			continue
		}
		if negative != nil && negative[topicKey] > 0 {
			continue
		}
		pairs = append(pairs, pair{key: key, score: score})
	}
	if len(pairs) == 0 {
		return ""
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return pairs[i].key < pairs[j].key
		}
		return pairs[i].score < pairs[j].score
	})
	return pairs[0].key
}

func profileKey(namespace, tag string) string {
	return profileKeyPrefix + namespace + ":" + tag
}

func shouldBalance(topCognition []string, scores map[string]float64) bool {
	if len(scores) == 0 {
		return false
	}
	if len(topCognition) == 0 {
		return false
	}
	top := scores[topCognition[0]]
	second := 0.0
	if len(topCognition) > 1 {
		second = scores[topCognition[1]]
	} else {
		for key, value := range scores {
			if key == topCognition[0] {
				continue
			}
			if value > second {
				second = value
			}
		}
	}
	if top <= 0 {
		return false
	}
	if second == 0 {
		return top >= 3
	}
	return top >= second*2.0
}
