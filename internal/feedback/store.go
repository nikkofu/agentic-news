package feedback

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikkofu/agentic-news/internal/profile"
)

const (
	eventsDirName        = "events"
	profileSnapshotFile  = "profile_snapshot.json"
	learningSnapshotFile = "learning_snapshot.json"
)

func NewStore(stateDir string) *Store {
	return &Store{root: filepath.Join(stateDir, "feedback")}
}

func (s *Store) AppendEvent(event Event) error {
	month, err := eventMonth(event)
	if err != nil {
		return err
	}
	path := filepath.Join(s.root, eventsDirName, month+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(payload, '\n')); err != nil {
		return err
	}
	return nil
}

func (s *Store) ReadEvents(month string) ([]Event, error) {
	normalizedMonth, err := validateMonth(month)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(s.root, eventsDirName, normalizedMonth+".jsonl")
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	events := make([]Event, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) WriteProfileSnapshot(p profile.UserProfile) error {
	path := filepath.Join(s.root, profileSnapshotFile)
	return writeSnapshot(path, p)
}

func (s *Store) ReadProfileSnapshot() (profile.UserProfile, error) {
	path := filepath.Join(s.root, profileSnapshotFile)
	var snapshot profile.UserProfile
	if err := readSnapshot(path, &snapshot); err != nil {
		return profile.UserProfile{}, err
	}
	return snapshot, nil
}

func (s *Store) WriteLearningSnapshot(v profile.LearningSnapshot) error {
	path := filepath.Join(s.root, learningSnapshotFile)
	return writeSnapshot(path, v)
}

func (s *Store) ReadLearningSnapshot() (profile.LearningSnapshot, error) {
	path := filepath.Join(s.root, learningSnapshotFile)
	var snapshot profile.LearningSnapshot
	if err := readSnapshot(path, &snapshot); err != nil {
		return profile.LearningSnapshot{}, err
	}
	return snapshot, nil
}

func eventMonth(event Event) (string, error) {
	if !event.Timestamp.IsZero() {
		return event.Timestamp.Format("2006-01"), nil
	}
	if event.EditionDate == "" {
		return "", errors.New("feedback: event timestamp missing")
	}
	editionDate, err := time.Parse("2006-01-02", event.EditionDate)
	if err != nil || editionDate.Format("2006-01-02") != event.EditionDate {
		return "", errors.New("feedback: invalid edition date")
	}
	return editionDate.Format("2006-01"), nil
}

func validateMonth(month string) (string, error) {
	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil || parsedMonth.Format("2006-01") != month {
		return "", errors.New("feedback: invalid month")
	}
	return month, nil
}

func writeSnapshot(path string, value any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(payload); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func readSnapshot(path string, target any) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(payload, target)
}
