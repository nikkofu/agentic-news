package feedback

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nikkofu/agentic-news/internal/profile"
)

type Service struct {
	store *Store
	mu    sync.Mutex
}

type ProfileResponse struct {
	FocusTopics           []string `json:"focus_topics"`
	PreferredStyles       []string `json:"preferred_styles"`
	CognitivePreferences  []string `json:"cognitive_preferences"`
	RecentFeedbackSummary []string `json:"recent_feedback_summary"`
	LastUpdatedAt         string   `json:"last_updated_at"`
}

type LearningResponse struct {
	TasteGrowthHint  string   `json:"taste_growth_hint"`
	KnowledgeGapHint string   `json:"knowledge_gap_hint"`
	TodayPlan        string   `json:"today_plan"`
	LearningTracks   []string `json:"learning_tracks"`
	UpdatedAt        string   `json:"updated_at"`
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) RecordEvent(event Event) (ProfileResponse, LearningResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event = normalizedEventIdentity(event)
	duplicate, err := s.hasEvent(event)
	if err != nil {
		return ProfileResponse{}, LearningResponse{}, err
	}
	if !duplicate {
		if err := s.store.AppendEvent(event); err != nil {
			return ProfileResponse{}, LearningResponse{}, err
		}
	}

	updatedProfile, updatedLearning, err := s.rebuildSnapshotsFromRawEvents()
	if err != nil {
		return ProfileResponse{}, LearningResponse{}, err
	}
	if err := s.store.WriteProfileSnapshot(updatedProfile); err != nil {
		return ProfileResponse{}, LearningResponse{}, err
	}
	if err := s.store.WriteLearningSnapshot(updatedLearning); err != nil {
		return ProfileResponse{}, LearningResponse{}, err
	}

	return profileResponseFromSnapshot(updatedProfile), learningResponseFromSnapshot(updatedLearning), nil
}

func (s *Service) GetProfile() (ProfileResponse, error) {
	snapshot, err := s.store.ReadProfileSnapshot()
	if err != nil {
		return ProfileResponse{}, err
	}
	return profileResponseFromSnapshot(snapshot), nil
}

func (s *Service) GetLearning() (LearningResponse, error) {
	snapshot, err := s.store.ReadLearningSnapshot()
	if err != nil {
		return LearningResponse{}, err
	}
	return learningResponseFromSnapshot(snapshot), nil
}

func (s *Service) eventInputFromEvent(event Event) profile.EventInput {
	input := profile.EventInput{
		EventType:     event.EventType,
		TopicTags:     cloneStrings(event.TopicTags),
		StyleTags:     cloneStrings(event.StyleTags),
		CognitiveTags: cloneStrings(event.CognitiveTags),
		SourceName:    event.SourceName,
		Timestamp:     event.Timestamp,
	}
	if event.Metadata == nil {
		return input
	}
	if dwellSeconds, ok := float64Value(event.Metadata["dwell_seconds"]); ok {
		input.DwellSeconds = dwellSeconds
	}
	if bookmarked, ok := boolValue(event.Metadata["bookmarked"]); ok {
		input.Bookmarked = bookmarked
	}
	if reasonTags, ok := stringSliceValue(event.Metadata["reason_tags"]); ok {
		input.ReasonTags = reasonTags
	}
	return input
}

func (s *Service) hasEvent(event Event) (bool, error) {
	eventID := eventIdentityKey(event)
	if eventID == "" {
		return false, nil
	}
	events, err := s.readAllEvents()
	if err != nil {
		return false, err
	}
	for _, existing := range events {
		if eventIdentityKey(existing) == eventID {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) rebuildSnapshotsFromRawEvents() (profile.UserProfile, profile.LearningSnapshot, error) {
	events, err := s.readAllEvents()
	if err != nil {
		return profile.UserProfile{}, profile.LearningSnapshot{}, err
	}
	current := profile.UserProfile{}
	for _, event := range events {
		current = profile.ApplyEvent(current, s.eventInputFromEvent(event))
	}
	return current, profile.BuildLearningSnapshot(current), nil
}

func (s *Service) readAllEvents() ([]Event, error) {
	eventsPath := filepath.Join(s.store.root, eventsDirName)
	entries, err := os.ReadDir(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, err
	}
	months := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".jsonl" {
			continue
		}
		months = append(months, strings.TrimSuffix(name, ".jsonl"))
	}
	sort.Strings(months)

	allEvents := make([]Event, 0)
	seenEventIDs := map[string]struct{}{}
	for _, month := range months {
		monthEvents, err := s.store.ReadEvents(month)
		if err != nil {
			return nil, err
		}
		for _, event := range monthEvents {
			event = normalizedEventIdentity(event)
			eventID := eventIdentityKey(event)
			if eventID != "" {
				if _, seen := seenEventIDs[eventID]; seen {
					continue
				}
				seenEventIDs[eventID] = struct{}{}
			}
			allEvents = append(allEvents, event)
		}
	}
	return allEvents, nil
}

func profileResponseFromSnapshot(snapshot profile.UserProfile) ProfileResponse {
	return ProfileResponse{
		FocusTopics:           responseStrings(snapshot.FocusTopics),
		PreferredStyles:       responseStrings(snapshot.PreferredStyles),
		CognitivePreferences:  responseStrings(snapshot.CognitivePreferences),
		RecentFeedbackSummary: responseStrings(snapshot.RecentFeedbackSummary),
		LastUpdatedAt:         formatTime(snapshot.LastUpdatedAt),
	}
}

func learningResponseFromSnapshot(snapshot profile.LearningSnapshot) LearningResponse {
	return LearningResponse{
		TasteGrowthHint:  snapshot.TasteGrowthHint,
		KnowledgeGapHint: snapshot.KnowledgeGapHint,
		TodayPlan:        snapshot.TodayPlan,
		LearningTracks:   responseStrings(snapshot.LearningTracks),
		UpdatedAt:        formatTime(snapshot.UpdatedAt),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func responseStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func float64Value(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case jsonNumberLike:
		parsed, err := strconv.ParseFloat(string(typed), 64)
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

type jsonNumberLike string

func boolValue(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return false, false
	}
}

func stringSliceValue(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		result := normalizedStringSlice(typed)
		return result, len(result) > 0
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				continue
			}
			trimmed := strings.TrimSpace(text)
			if trimmed == "" {
				continue
			}
			result = append(result, trimmed)
		}
		return result, len(result) > 0
	default:
		return nil, false
	}
}

func normalizedStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func normalizedEventIdentity(event Event) Event {
	event.EventID = strings.TrimSpace(event.EventID)
	if event.EventID != "" {
		return event
	}

	stamp := event.EditionDate
	if !event.Timestamp.IsZero() {
		stamp = event.Timestamp.UTC().Format(time.RFC3339Nano)
	}
	parts := []string{stamp, strings.TrimSpace(event.EventType), strings.TrimSpace(event.ArticleID)}
	event.EventID = strings.Join(parts, "-")
	return event
}

func eventIdentityKey(event Event) string {
	return strings.TrimSpace(normalizedEventIdentity(event).EventID)
}
