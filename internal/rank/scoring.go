package rank

import (
	"strings"

	"github.com/nikkofu/agentic-news/internal/model"
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
