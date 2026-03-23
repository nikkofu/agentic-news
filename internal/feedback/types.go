package feedback

import (
	"time"

	"github.com/nikkofu/agentic-news/internal/profile"
)

type Event struct {
	EventID       string
	EventType     string
	Timestamp     time.Time
	EditionDate   string
	ArticleID     string
	ArticleTitle  string
	ArticleURL    string
	SourceName    string
	TopicTags     []string
	StyleTags     []string
	CognitiveTags []string
	Metadata      map[string]any
}

type Store struct {
	root string
}

type LearningSnapshot = profile.LearningSnapshot
