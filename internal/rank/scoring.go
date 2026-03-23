package rank

import (
	"slices"
	"strings"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
)

type Weights struct {
	Importance        float64
	PersonalRelevance float64
	Credibility       float64
	Novelty           float64
	Freshness         float64
}

func DefaultWeights() Weights {
	return Weights{
		Importance:        0.30,
		PersonalRelevance: 0.25,
		Credibility:       0.20,
		Novelty:           0.15,
		Freshness:         0.10,
	}
}

func ScoreItem(s model.ScoreSignals, w Weights, category string) float64 {
	base := s.Importance*w.Importance +
		s.PersonalRelevance*w.PersonalRelevance +
		s.Credibility*w.Credibility +
		s.Novelty*w.Novelty +
		s.Freshness*w.Freshness

	cat := strings.ToLower(strings.TrimSpace(category))
	switch cat {
	case "tech", "technology":
		base += s.Novelty * 0.02
	case "public", "policy", "politics":
		base += s.Credibility * 0.02
	case "finance", "business":
		base += s.Importance * 0.02
	}

	if base < 0 {
		return 0
	}
	if base > 100 {
		return 100
	}
	return base
}

func PassesSourceTierThreshold(tier string, confidence int, corroborated bool) bool {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case "A":
		return confidence >= 60
	case "B":
		return confidence >= 70
	case "C":
		return confidence >= 80 && corroborated
	default:
		return confidence >= 75
	}
}

func ScorePersonalRelevance(
	p profile.UserProfile,
	topicTags []string,
	styleTags []string,
	cognitiveTags []string,
	sourceName string,
) float64 {
	normalizedTopicAffinity := normalizeAffinityMap(p.TopicAffinity)
	normalizedStyleAffinity := normalizeAffinityMap(p.StyleAffinity)
	normalizedCognitiveAffinity := normalizeAffinityMap(p.CognitiveAffinity)
	normalizedSourceAffinity := normalizeAffinityMap(p.SourceAffinity)
	normalizedNegative := normalizeAffinityMap(p.NegativeSignals)

	score := 52.0
	score += scoreFromTags(normalizedTopicAffinity, topicTags) * 1.0
	score += scoreFromTags(normalizedStyleAffinity, styleTags) * 0.7
	score += scoreFromTags(normalizedCognitiveAffinity, cognitiveTags) * 0.6

	sourceKey := strings.ToLower(strings.TrimSpace(sourceName))
	if sourceKey != "" {
		score += normalizedSourceAffinity[sourceKey] * 0.5
	}

	score -= scoreFromNegativeTags(normalizedNegative, "profile:topic:", topicTags) * 0.9
	score -= scoreFromNegativeTags(normalizedNegative, "profile:style:", styleTags) * 0.7
	score -= scoreFromNegativeTags(normalizedNegative, "profile:cognitive:", cognitiveTags) * 0.6
	if sourceKey != "" {
		score -= normalizedNegative["profile:source:"+sourceKey] * 0.7
	}

	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func RankItems(items []model.DailyPick) []model.DailyPick {
	ranked := slices.Clone(items)
	slices.SortStableFunc(ranked, func(a, b model.DailyPick) int {
		if cardTypePriority(a.CardType) != cardTypePriority(b.CardType) {
			if cardTypePriority(a.CardType) < cardTypePriority(b.CardType) {
				return -1
			}
			return 1
		}
		if a.ScoreFinal != b.ScoreFinal {
			if a.ScoreFinal > b.ScoreFinal {
				return -1
			}
			return 1
		}
		if !a.PublishedAt.Equal(b.PublishedAt) {
			if a.PublishedAt.After(b.PublishedAt) {
				return -1
			}
			return 1
		}
		return strings.Compare(a.ID, b.ID)
	})
	return ranked
}

func cardTypePriority(cardType string) int {
	switch strings.ToLower(strings.TrimSpace(cardType)) {
	case "brief":
		return 1
	default:
		return 0
	}
}

func normalizeAffinityMap(source map[string]float64) map[string]float64 {
	if len(source) == 0 {
		return map[string]float64{}
	}
	normalized := make(map[string]float64, len(source))
	for key, value := range source {
		trimmedKey := strings.ToLower(strings.TrimSpace(key))
		if trimmedKey == "" {
			continue
		}
		normalized[trimmedKey] += value
	}
	return normalized
}

func scoreFromTags(affinity map[string]float64, tags []string) float64 {
	score := 0.0
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		score += affinity[normalized]
	}
	return score
}

func scoreFromNegativeTags(negative map[string]float64, prefix string, tags []string) float64 {
	score := 0.0
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		score += negative[prefix+normalized]
	}
	return score
}
