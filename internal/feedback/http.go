package feedback

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type eventRequest struct {
	EventID       string         `json:"event_id"`
	EventType     string         `json:"event_type"`
	Timestamp     string         `json:"timestamp"`
	EditionDate   string         `json:"edition_date"`
	ArticleID     string         `json:"article_id"`
	ArticleTitle  string         `json:"article_title"`
	ArticleURL    string         `json:"article_url"`
	SourceName    string         `json:"source_name"`
	TopicTags     []string       `json:"topic_tags"`
	StyleTags     []string       `json:"style_tags"`
	CognitiveTags []string       `json:"cognitive_tags"`
	Metadata      map[string]any `json:"metadata"`
}

func NewHandler(svc *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/feedback/events", postEventHandler(svc))
	mux.HandleFunc("/api/v1/profile", getProfileHandler(svc))
	mux.HandleFunc("/api/v1/profile/learning", getLearningHandler(svc))
	return mux
}

func postEventHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req eventRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "malformed event payload")
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			writeJSONError(w, http.StatusBadRequest, "malformed event payload")
			return
		}

		event, err := eventFromRequest(req)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		profileResp, _, err := svc.RecordEvent(event)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to record event")
			return
		}
		writeJSON(w, http.StatusOK, profileResp)
	}
}

func getProfileHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		resp, err := svc.GetProfile()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load profile")
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func getLearningHandler(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		resp, err := svc.GetLearning()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to load learning profile")
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func eventFromRequest(req eventRequest) (Event, error) {
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		return Event{}, errors.New("event_type is required")
	}
	articleID := strings.TrimSpace(req.ArticleID)
	if articleID == "" {
		return Event{}, errors.New("article_id is required")
	}
	timestampText := strings.TrimSpace(req.Timestamp)
	if timestampText == "" {
		return Event{}, errors.New("timestamp is required")
	}
	timestamp, err := time.Parse(time.RFC3339, timestampText)
	if err != nil {
		return Event{}, errors.New("timestamp must be RFC3339")
	}

	editionDate := strings.TrimSpace(req.EditionDate)
	if editionDate == "" {
		editionDate = timestamp.UTC().Format("2006-01-02")
	}

	return Event{
		EventID:       firstNonEmpty(strings.TrimSpace(req.EventID), generatedEventID(eventType, articleID, timestamp.UTC())),
		EventType:     eventType,
		Timestamp:     timestamp.UTC(),
		EditionDate:   editionDate,
		ArticleID:     articleID,
		ArticleTitle:  strings.TrimSpace(req.ArticleTitle),
		ArticleURL:    strings.TrimSpace(req.ArticleURL),
		SourceName:    strings.TrimSpace(req.SourceName),
		TopicTags:     normalizedStringSlice(req.TopicTags),
		StyleTags:     normalizedStringSlice(req.StyleTags),
		CognitiveTags: normalizedStringSlice(req.CognitiveTags),
		Metadata:      cloneMap(req.Metadata),
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	clone := make(map[string]any, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func generatedEventID(eventType, articleID string, timestamp time.Time) string {
	return strings.Join([]string{
		timestamp.UTC().Format("20060102T150405.000000000Z"),
		eventType,
		articleID,
	}, "-")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
