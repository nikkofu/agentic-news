package feedback_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/feedback"
	"github.com/nikkofu/agentic-news/internal/profile"
)

func TestStore_AppendEventWritesJSONLRecord(t *testing.T) {
	store := feedback.NewStore(t.TempDir())
	eventTime := time.Date(2026, 3, 19, 10, 30, 0, 0, time.UTC)
	event := feedback.Event{
		EventID:       "evt-1",
		EventType:     "feedback_like",
		Timestamp:     eventTime,
		EditionDate:   "2026-03-19",
		ArticleID:     "article-1",
		ArticleTitle:  "AI 基础设施",
		ArticleURL:    "https://example.com/ai",
		SourceName:    "Example News",
		TopicTags:     []string{"AI 基础设施"},
		StyleTags:     []string{"数据型"},
		CognitiveTags: []string{"结构化"},
		Metadata: map[string]any{
			"score": 1.5,
			"flag":  true,
		},
	}

	if err := store.AppendEvent(event); err != nil {
		t.Fatalf("append event: %v", err)
	}

	events, err := store.ReadEvents("2026-03")
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !reflect.DeepEqual(events[0], event) {
		t.Fatalf("expected event round-trip, got %#v", events[0])
	}
}

func TestStore_AppendEventPreservesOrder(t *testing.T) {
	store := feedback.NewStore(t.TempDir())
	first := feedback.Event{
		EventID:     "evt-1",
		EventType:   "feedback_like",
		Timestamp:   time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC),
		EditionDate: "2026-03-19",
		ArticleID:   "article-1",
		Metadata: map[string]any{
			"index": float64(1),
		},
	}
	second := feedback.Event{
		EventID:     "evt-2",
		EventType:   "feedback_like",
		Timestamp:   time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC),
		EditionDate: "2026-03-19",
		ArticleID:   "article-2",
		Metadata: map[string]any{
			"index": float64(2),
		},
	}

	if err := store.AppendEvent(first); err != nil {
		t.Fatalf("append first event: %v", err)
	}
	if err := store.AppendEvent(second); err != nil {
		t.Fatalf("append second event: %v", err)
	}

	events, err := store.ReadEvents("2026-03")
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if !reflect.DeepEqual(events[0], first) || !reflect.DeepEqual(events[1], second) {
		t.Fatalf("expected ordered events, got %#v", events)
	}
}

func TestStore_ReadEventsMissingMonthReturnsEmpty(t *testing.T) {
	store := feedback.NewStore(t.TempDir())

	events, err := store.ReadEvents("2026-02")
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestStore_ReadEventsRejectsInvalidMonth(t *testing.T) {
	store := feedback.NewStore(t.TempDir())
	invalidMonths := []string{
		"../2026-03",
		"2026-3",
		"2026-13",
		"2026/03",
	}

	for _, month := range invalidMonths {
		month := month
		t.Run(month, func(t *testing.T) {
			events, err := store.ReadEvents(month)
			if err == nil {
				t.Fatalf("expected error for invalid month %q, got events %#v", month, events)
			}
		})
	}
}

func TestStore_AppendEventRejectsMalformedEditionDateWithoutTimestamp(t *testing.T) {
	store := feedback.NewStore(t.TempDir())
	event := feedback.Event{
		EventID:     "evt-invalid-edition-date",
		EventType:   "feedback_like",
		EditionDate: "2026-3-19",
		ArticleID:   "article-invalid",
	}

	if err := store.AppendEvent(event); err == nil {
		t.Fatal("expected malformed edition date to be rejected when timestamp is missing")
	}
}

func TestStore_WriteAndReadProfileSnapshot(t *testing.T) {
	stateDir := t.TempDir()
	store := feedback.NewStore(stateDir)
	updatedAt := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	profileSnapshot := profile.UserProfile{
		FocusTopics:          []string{"AI 基础设施"},
		PreferredStyles:      []string{"数据型"},
		CognitivePreferences: []string{"结构化"},
		ExplicitFeedback: map[string]string{
			"profile:topic:AI 基础设施": "like",
		},
		BehaviorSignals: map[string]float64{
			"AI 基础设施": 2.5,
		},
		TopicAffinity: map[string]float64{
			"AI 基础设施": 6,
		},
		StyleAffinity: map[string]float64{
			"数据型": 3,
		},
		CognitiveAffinity: map[string]float64{
			"结构化": 2,
		},
		SourceAffinity: map[string]float64{
			"Example News": 1.2,
		},
		NegativeSignals: map[string]float64{
			"profile:topic:标题党": 4,
		},
		RecentFeedbackSummary: []string{"喜欢 AI 基础设施"},
		LastUpdatedAt:         updatedAt,
	}

	if err := store.WriteProfileSnapshot(profileSnapshot); err != nil {
		t.Fatalf("write profile snapshot: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(stateDir, "feedback", "profile_snapshot.json"))
	if err != nil {
		t.Fatalf("read profile snapshot file: %v", err)
	}
	if !strings.Contains(string(content), "\n  \"") {
		t.Fatalf("expected indented json, got %q", string(content))
	}

	loaded, err := store.ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if !reflect.DeepEqual(loaded, profileSnapshot) {
		t.Fatalf("expected profile snapshot round-trip, got %#v", loaded)
	}
}

func TestStore_ReadProfileSnapshotMissingReturnsZero(t *testing.T) {
	store := feedback.NewStore(t.TempDir())

	loaded, err := store.ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if !reflect.DeepEqual(loaded, profile.UserProfile{}) {
		t.Fatalf("expected zero profile snapshot, got %#v", loaded)
	}
}

func TestStore_WriteAndReadLearningSnapshot(t *testing.T) {
	store := feedback.NewStore(t.TempDir())
	updatedAt := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	snapshot := profile.LearningSnapshot{
		TasteGrowthHint:  "继续深挖 AI 基础设施",
		KnowledgeGapHint: "补充产业链基础",
		TodayPlan:        "今天优先阅读 AI 基础设施",
		LearningTracks:   []string{"追踪 AI 基础设施", "平衡视角"},
		UpdatedAt:        updatedAt,
	}

	if err := store.WriteLearningSnapshot(snapshot); err != nil {
		t.Fatalf("write learning snapshot: %v", err)
	}

	loaded, err := store.ReadLearningSnapshot()
	if err != nil {
		t.Fatalf("read learning snapshot: %v", err)
	}
	if !reflect.DeepEqual(loaded, snapshot) {
		t.Fatalf("expected learning snapshot round-trip, got %#v", loaded)
	}
}

func TestStore_ReadLearningSnapshotMissingReturnsZero(t *testing.T) {
	store := feedback.NewStore(t.TempDir())

	loaded, err := store.ReadLearningSnapshot()
	if err != nil {
		t.Fatalf("read learning snapshot: %v", err)
	}
	if !reflect.DeepEqual(loaded, profile.LearningSnapshot{}) {
		t.Fatalf("expected zero learning snapshot, got %#v", loaded)
	}
}
