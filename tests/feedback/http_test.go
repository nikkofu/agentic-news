package feedback_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/agentic-news/internal/feedback"
	"github.com/nikkofu/agentic-news/internal/profile"
)

func TestFeedbackAPI_PostEventAppendsAndReturnsUpdatedProfile(t *testing.T) {
	stateDir := t.TempDir()
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(stateDir))))
	defer server.Close()

	payload := `{
	  "event_type":"feedback_like",
	  "timestamp":"2026-03-19T09:00:00Z",
	  "edition_date":"2026-03-19",
	  "article_id":"a1",
	  "article_title":"AI 基础设施观察",
	  "source_name":"Example News",
	  "topic_tags":["AI 基础设施"],
	  "style_tags":["深度分析"],
	  "cognitive_tags":["风险"],
	  "metadata":{
	    "dwell_seconds":90,
	    "bookmarked":true,
	    "reason_tags":["想看更多数据"]
	  }
	}`

	resp, err := http.Post(server.URL+"/api/v1/feedback/events", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var got struct {
		FocusTopics           []string `json:"focus_topics"`
		PreferredStyles       []string `json:"preferred_styles"`
		CognitivePreferences  []string `json:"cognitive_preferences"`
		RecentFeedbackSummary []string `json:"recent_feedback_summary"`
		LastUpdatedAt         string   `json:"last_updated_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !reflect.DeepEqual(got.FocusTopics, []string{"AI 基础设施"}) {
		t.Fatalf("unexpected focus topics: %#v", got.FocusTopics)
	}
	if !reflect.DeepEqual(got.PreferredStyles, []string{"深度分析"}) {
		t.Fatalf("unexpected preferred styles: %#v", got.PreferredStyles)
	}
	if !reflect.DeepEqual(got.CognitivePreferences, []string{"风险"}) {
		t.Fatalf("unexpected cognitive preferences: %#v", got.CognitivePreferences)
	}
	if len(got.RecentFeedbackSummary) == 0 {
		t.Fatal("expected recent feedback summary")
	}
	if got.LastUpdatedAt != "2026-03-19T09:00:00Z" {
		t.Fatalf("unexpected last_updated_at: %q", got.LastUpdatedAt)
	}

	storedEvents, err := feedback.NewStore(stateDir).ReadEvents("2026-03")
	if err != nil {
		t.Fatalf("read stored events: %v", err)
	}
	if len(storedEvents) != 1 {
		t.Fatalf("expected 1 stored event, got %d", len(storedEvents))
	}
	if strings.TrimSpace(storedEvents[0].EventID) == "" {
		t.Fatalf("expected generated event id, got %#v", storedEvents[0])
	}
	if storedEvents[0].Metadata == nil {
		t.Fatal("expected metadata to be persisted")
	}

	storedProfile, err := feedback.NewStore(stateDir).ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if !reflect.DeepEqual(storedProfile.FocusTopics, []string{"AI 基础设施"}) {
		t.Fatalf("unexpected stored focus topics: %#v", storedProfile.FocusTopics)
	}
	if storedProfile.BehaviorSignals["AI 基础设施"] <= 0 {
		t.Fatalf("expected metadata dwell_seconds to affect behavior signals, got %#v", storedProfile.BehaviorSignals)
	}
	if len(storedProfile.RecentFeedbackSummary) == 0 {
		t.Fatal("expected stored recent feedback summary")
	}

	storedLearning, err := feedback.NewStore(stateDir).ReadLearningSnapshot()
	if err != nil {
		t.Fatalf("read learning snapshot: %v", err)
	}
	if storedLearning.TodayPlan == "" {
		t.Fatal("expected learning snapshot to be written")
	}
}

func TestFeedbackAPI_GetProfileReturnsSnapshot(t *testing.T) {
	stateDir := t.TempDir()
	store := feedback.NewStore(stateDir)
	updatedAt := time.Date(2026, 3, 19, 11, 30, 0, 0, time.UTC)
	if err := store.WriteProfileSnapshot(profile.UserProfile{
		FocusTopics:           []string{"半导体", "AI 基础设施"},
		PreferredStyles:       []string{"数据型"},
		CognitivePreferences:  []string{"结构化"},
		RecentFeedbackSummary: []string{"你最近强化了半导体内容。"},
		LastUpdatedAt:         updatedAt,
	}); err != nil {
		t.Fatalf("seed profile snapshot: %v", err)
	}

	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(store)))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expected := map[string]any{
		"focus_topics":            []any{"半导体", "AI 基础设施"},
		"preferred_styles":        []any{"数据型"},
		"cognitive_preferences":   []any{"结构化"},
		"recent_feedback_summary": []any{"你最近强化了半导体内容。"},
		"last_updated_at":         "2026-03-19T11:30:00Z",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected profile response: %#v", got)
	}
}

func TestFeedbackAPI_GetLearningReturnsTracksAndTodayPlan(t *testing.T) {
	stateDir := t.TempDir()
	store := feedback.NewStore(stateDir)
	updatedAt := time.Date(2026, 3, 19, 12, 15, 0, 0, time.UTC)
	if err := store.WriteLearningSnapshot(profile.LearningSnapshot{
		TasteGrowthHint:  "继续深挖 AI 基础设施",
		KnowledgeGapHint: "补充供应链基础",
		TodayPlan:        "今天优先阅读 AI 基础设施",
		LearningTracks:   []string{"追踪 AI 基础设施", "平衡视角"},
		UpdatedAt:        updatedAt,
	}); err != nil {
		t.Fatalf("seed learning snapshot: %v", err)
	}

	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(store)))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/profile/learning")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expected := map[string]any{
		"taste_growth_hint":  "继续深挖 AI 基础设施",
		"knowledge_gap_hint": "补充供应链基础",
		"today_plan":         "今天优先阅读 AI 基础设施",
		"learning_tracks":    []any{"追踪 AI 基础设施", "平衡视角"},
		"updated_at":         "2026-03-19T12:15:00Z",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected learning response: %#v", got)
	}
}

func TestFeedbackAPI_RejectsMalformedEventPayload(t *testing.T) {
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(t.TempDir()))))
	defer server.Close()

	payload := []byte(`{"timestamp":"2026-03-19T09:00:00Z","edition_date":"2026-03-19"}`)
	resp, err := http.Post(server.URL+"/api/v1/feedback/events", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestFeedbackAPI_RejectsInvalidJSONPayload(t *testing.T) {
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(t.TempDir()))))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/v1/feedback/events", "application/json", strings.NewReader(`{"event_type":`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestFeedbackAPI_PostEventAcceptsReasonTagsFromTopLevelMetadata(t *testing.T) {
	stateDir := t.TempDir()
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(stateDir))))
	defer server.Close()

	payload := `{
	  "event_type":"feedback_dislike",
	  "timestamp":"2026-03-19T10:00:00Z",
	  "edition_date":"2026-03-19",
	  "article_id":"a2",
	  "topic_tags":["标题党"],
	  "metadata":{"reason_tags":["信息密度低"]}
	}`

	resp, err := http.Post(server.URL+"/api/v1/feedback/events", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	storedProfile, err := feedback.NewStore(stateDir).ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if storedProfile.NegativeSignals["profile:reason:信息密度低"] <= 0 {
		t.Fatalf("expected reason_tags metadata to affect negative signals, got %#v", storedProfile.NegativeSignals)
	}
}

func TestFeedbackAPI_PostEventRetryWithSameEventIDDoesNotDoubleApplyAcrossMonths(t *testing.T) {
	stateDir := t.TempDir()
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(stateDir))))
	defer server.Close()

	firstPayload := `{
	  "event_id":"evt-retry-001",
	  "event_type":"feedback_like",
	  "timestamp":"2026-03-19T09:00:00Z",
	  "edition_date":"2026-03-19",
	  "article_id":"a-retry",
	  "topic_tags":["AI 基础设施"]
	}`
	secondPayload := `{
	  "event_id":"evt-retry-001",
	  "event_type":"feedback_like",
	  "timestamp":"2026-04-01T09:00:00Z",
	  "edition_date":"2026-04-01",
	  "article_id":"a-retry",
	  "topic_tags":["AI 基础设施"]
	}`

	for _, payload := range []string{firstPayload, secondPayload} {
		resp, err := http.Post(server.URL+"/api/v1/feedback/events", "application/json", strings.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}

	marchEvents, err := feedback.NewStore(stateDir).ReadEvents("2026-03")
	if err != nil {
		t.Fatalf("read March events: %v", err)
	}
	aprilEvents, err := feedback.NewStore(stateDir).ReadEvents("2026-04")
	if err != nil {
		t.Fatalf("read April events: %v", err)
	}
	if len(marchEvents) != 1 {
		t.Fatalf("expected 1 March event, got %d", len(marchEvents))
	}
	if len(aprilEvents) != 0 {
		t.Fatalf("expected retry to avoid appending April duplicate, got %d", len(aprilEvents))
	}

	storedProfile, err := feedback.NewStore(stateDir).ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if got := storedProfile.TopicAffinity["AI 基础设施"]; got != 6 {
		t.Fatalf("expected single-applied topic affinity, got %v", got)
	}
}

func TestFeedbackAPI_PostEventRejectsWrongMethod(t *testing.T) {
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(t.TempDir()))))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/feedback/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestFeedbackAPI_GetProfileDoesNotExposeHeavySnapshotFields(t *testing.T) {
	stateDir := t.TempDir()
	store := feedback.NewStore(stateDir)
	if err := store.WriteProfileSnapshot(profile.UserProfile{
		FocusTopics:     []string{"AI"},
		TopicAffinity:   map[string]float64{"AI": 9},
		BehaviorSignals: map[string]float64{"AI": 3},
	}); err != nil {
		t.Fatalf("seed profile snapshot: %v", err)
	}

	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(store)))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, exists := got["topic_affinity"]; exists {
		t.Fatalf("expected lightweight response, got %#v", got)
	}
	if _, exists := got["behavior_signals"]; exists {
		t.Fatalf("expected lightweight response, got %#v", got)
	}

	path := filepath.Join(stateDir, "feedback", "profile_snapshot.json")
	if path == "" {
		t.Fatal("expected seeded path")
	}
}

func TestFeedbackAPI_GetProfileMissingSnapshotReturnsEmptyLists(t *testing.T) {
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(t.TempDir()))))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expected := map[string]any{
		"focus_topics":            []any{},
		"preferred_styles":        []any{},
		"cognitive_preferences":   []any{},
		"recent_feedback_summary": []any{},
		"last_updated_at":         "",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected profile response: %#v", got)
	}
}

func TestFeedbackAPI_GetLearningMissingSnapshotReturnsEmptyLists(t *testing.T) {
	server := httptest.NewServer(feedback.NewHandler(feedback.NewService(feedback.NewStore(t.TempDir()))))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/profile/learning")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expected := map[string]any{
		"taste_growth_hint":  "",
		"knowledge_gap_hint": "",
		"today_plan":         "",
		"learning_tracks":    []any{},
		"updated_at":         "",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected learning response: %#v", got)
	}
}

func TestService_RecordEventRetryWithSameLogicalEventDoesNotDuplicateOrDoubleApply(t *testing.T) {
	stateDir := t.TempDir()
	store := feedback.NewStore(stateDir)
	svc := feedback.NewService(store)
	event := feedback.Event{
		EventID:     "evt-retry",
		EventType:   "feedback_like",
		Timestamp:   time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC),
		EditionDate: "2026-03-19",
		ArticleID:   "retry-article",
		TopicTags:   []string{"AI 基础设施"},
	}

	for attempt := 0; attempt < 2; attempt++ {
		if _, _, err := svc.RecordEvent(event); err != nil {
			t.Fatalf("attempt %d record event: %v", attempt+1, err)
		}
	}

	storedEvents, err := store.ReadEvents("2026-03")
	if err != nil {
		t.Fatalf("read stored events: %v", err)
	}
	if len(storedEvents) != 1 {
		t.Fatalf("expected deduped event log, got %d events", len(storedEvents))
	}

	snapshot, err := store.ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if got := snapshot.TopicAffinity["AI 基础设施"]; got != 6 {
		t.Fatalf("expected single application of feedback_like, got affinity %v", got)
	}
	if len(snapshot.RecentFeedbackSummary) != 1 {
		t.Fatalf("expected one summary entry, got %#v", snapshot.RecentFeedbackSummary)
	}
}

func TestService_RecordEventRetryRebuildsSnapshotsFromRawLog(t *testing.T) {
	stateDir := t.TempDir()
	store := feedback.NewStore(stateDir)
	svc := feedback.NewService(store)
	event := feedback.Event{
		EventID:     "evt-retry",
		EventType:   "feedback_like",
		Timestamp:   time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC),
		EditionDate: "2026-03-19",
		ArticleID:   "retry-article",
		TopicTags:   []string{"AI 基础设施"},
	}
	if err := store.AppendEvent(event); err != nil {
		t.Fatalf("seed raw event: %v", err)
	}
	if _, _, err := svc.RecordEvent(event); err != nil {
		t.Fatalf("retry record event: %v", err)
	}

	storedEvents, err := store.ReadEvents("2026-03")
	if err != nil {
		t.Fatalf("read stored events: %v", err)
	}
	if len(storedEvents) != 1 {
		t.Fatalf("expected no duplicate append on retry, got %d events", len(storedEvents))
	}

	snapshot, err := store.ReadProfileSnapshot()
	if err != nil {
		t.Fatalf("read profile snapshot: %v", err)
	}
	if got := snapshot.TopicAffinity["AI 基础设施"]; got != 6 {
		t.Fatalf("expected rebuild from raw log, got affinity %v", got)
	}

	learning, err := store.ReadLearningSnapshot()
	if err != nil {
		t.Fatalf("read learning snapshot: %v", err)
	}
	if learning.TodayPlan == "" {
		t.Fatal("expected learning snapshot to be rebuilt from raw log")
	}
}
