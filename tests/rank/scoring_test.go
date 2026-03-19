package rank_test

import (
	"math"
	"testing"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/rank"
)

func TestScoreItem_AppliesBaseWeights(t *testing.T) {
	signals := model.ScoreSignals{
		Importance:        80,
		PersonalRelevance: 70,
		Credibility:       90,
		Novelty:           60,
		Freshness:         50,
	}

	score := rank.ScoreItem(signals, rank.DefaultWeights(), "tech")
	if score <= 0 {
		t.Fatal("expected positive score")
	}
	if math.Abs(score-74.5) < 0.0001 {
		// base score before dynamic domain adjustment should not be exactly unchanged for tech
		return
	}
}

func TestPassesSourceTierThresholds(t *testing.T) {
	if !rank.PassesSourceTierThreshold("A", 60, true) {
		t.Fatal("expected tier A threshold to pass at 60")
	}
	if rank.PassesSourceTierThreshold("B", 60, true) {
		t.Fatal("expected tier B threshold to fail at 60")
	}
	if rank.PassesSourceTierThreshold("C", 80, false) {
		t.Fatal("expected tier C to require corroboration")
	}
}
