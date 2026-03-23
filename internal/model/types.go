package model

import (
	"strings"
	"time"
)

const DefaultThemeID = "editorial-ai"

type DailyPick struct {
	ID              string
	CardType        string
	FallbackReason  string
	Category        string
	Title           string
	Summary         string
	ScoreFinal      float64
	CoverImage      string
	CoverImageLocal string
	SourceName      string
	SourceURL       string
	PublishedAt     time.Time
	ReadingTimeMin  int
	TopicTags       []string
	StyleTags       []string
	CognitiveTags   []string
	Insight         Insight
}

type DailyEdition struct {
	Date        time.Time
	ThemeID     string
	Keywords    []string
	Featured    []DailyPick
	Learning    []string
	GeneratedAt time.Time
}

type ScoreSignals struct {
	Importance        float64
	PersonalRelevance float64
	Credibility       float64
	Novelty           float64
	Freshness         float64
}

type Insight struct {
	SummaryBrief       string
	SummaryDeep        string
	KeyPoints          []string
	Viewpoint          string
	OpportunityRisk    string
	ContrarianTake     string
	LearningSuggestion []string
	Confidence         int
	WhyForYou          string
	TasteGrowthHint    string
	KnowledgeGapHint   string
	ModelName          string
	ModelVersion       string
	PromptVersion      string
	SourceRefs         []string
	EvidenceSnippets   []string
	GeneratedAt        time.Time
}

type ArticleImage struct {
	URL    string
	Source string
}

type Article struct {
	ArticleID         string
	CanonicalURL      string
	Title             string
	Excerpt           string
	CoverImage        string
	ImageCandidates   []ArticleImage
	ContentText       string
	Keywords          []string
	Entities          []string
	CategoryPrimary   string
	CategorySecondary []string
	PublishedAt       time.Time
	IngestedAt        time.Time
	Language          string
}

type RawItem struct {
	ItemID      string
	SourceID    string
	Title       string
	URL         string
	PublishedAt time.Time
	RawContent  string
	Author      string
	Language    string
}

func NormalizeThemeID(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return DefaultThemeID
	}

	var b strings.Builder
	lastDash := false
	for _, ch := range normalized {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
			lastDash = false
			continue
		}
		if b.Len() == 0 || lastDash {
			continue
		}
		b.WriteByte('-')
		lastDash = true
	}

	themeID := strings.Trim(b.String(), "-")
	if themeID == "" {
		return DefaultThemeID
	}
	return themeID
}
