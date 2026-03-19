package model

import "time"

type DailyPick struct {
	ID              string
	Category        string
	Title           string
	Summary         string
	ScoreFinal      float64
	CoverImage      string
	SourceName      string
	SourceURL       string
	PublishedAt     time.Time
	ReadingTimeMin  int
	Insight         Insight
}

type DailyEdition struct {
	Date            time.Time
	Keywords        []string
	Featured        []DailyPick
	Learning        []string
	GeneratedAt     time.Time
}

type ScoreSignals struct {
	Importance        float64
	PersonalRelevance float64
	Credibility       float64
	Novelty           float64
	Freshness         float64
}

type Insight struct {
	SummaryBrief      string
	SummaryDeep       string
	KeyPoints         []string
	Viewpoint         string
	OpportunityRisk   string
	ContrarianTake    string
	LearningSuggestion []string
	Confidence        int
	WhyForYou         string
	TasteGrowthHint   string
	KnowledgeGapHint  string
	ModelName         string
	ModelVersion      string
	PromptVersion     string
	SourceRefs        []string
	EvidenceSnippets  []string
	GeneratedAt       time.Time
}

type Article struct {
	ArticleID          string
	CanonicalURL       string
	Title              string
	Excerpt            string
	CoverImage         string
	ContentText        string
	Keywords           []string
	Entities           []string
	CategoryPrimary    string
	CategorySecondary  []string
	PublishedAt        time.Time
	IngestedAt         time.Time
	Language           string
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
