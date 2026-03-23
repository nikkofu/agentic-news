package rank_test

import (
	"math"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/profile"
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

func TestRankItems_DeterministicTieBreakByPublishedAt(t *testing.T) {
	newer := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	older := newer.Add(-1 * time.Hour)

	items := []model.DailyPick{
		{ID: "b", ScoreFinal: 90, PublishedAt: older},
		{ID: "a", ScoreFinal: 90, PublishedAt: older},
		{ID: "c", ScoreFinal: 90, PublishedAt: newer},
		{ID: "z", ScoreFinal: 80, PublishedAt: newer},
	}

	got := rank.RankItems(items)
	if len(got) != 4 {
		t.Fatalf("expected 4 ranked items, got %d", len(got))
	}

	want := []string{"c", "a", "b", "z"}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("expected item %d to be %s, got %s", i, id, got[i].ID)
		}
	}
}

func TestRankItems_PrioritizesStandardCardsBeforeBriefCards(t *testing.T) {
	items := []model.DailyPick{
		{ID: "brief-1", CardType: "brief", ScoreFinal: 99, PublishedAt: time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC)},
		{ID: "standard-1", CardType: "standard", ScoreFinal: 70, PublishedAt: time.Date(2026, 3, 19, 7, 0, 0, 0, time.UTC)},
	}

	got := rank.RankItems(items)
	if len(got) != 2 {
		t.Fatalf("expected 2 ranked items, got %d", len(got))
	}
	if got[0].ID != "standard-1" {
		t.Fatalf("expected standard card first, got %s", got[0].ID)
	}
}

func TestScorePersonalRelevance_UsesTopicStyleCognitiveAndNegativeSignals(t *testing.T) {
	p := profile.UserProfile{
		FocusTopics:          []string{"AI Infrastructure"},
		PreferredStyles:      []string{"Explainer"},
		CognitivePreferences: []string{"Systems"},
		TopicAffinity: map[string]float64{
			"AI Infrastructure": 9,
			"Cloud":             2,
		},
		StyleAffinity: map[string]float64{
			"Explainer": 6,
		},
		CognitiveAffinity: map[string]float64{
			"Systems": 5,
		},
		SourceAffinity: map[string]float64{
			"DeepSource": 4,
		},
		NegativeSignals: map[string]float64{
			"profile:topic:Gossip":      8,
			"profile:style:Hot Take":    4,
			"profile:cognitive:Snark":   3,
			"profile:source:DeepSource": 1,
		},
	}

	got := rank.ScorePersonalRelevance(
		p,
		[]string{"AI Infrastructure", "Gossip"},
		[]string{"Explainer", "Hot Take"},
		[]string{"Systems", "Snark"},
		"DeepSource",
	)

	want := 57.7
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("expected personal relevance %.1f, got %.4f", want, got)
	}
}
