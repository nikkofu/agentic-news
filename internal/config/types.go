package config

type Config struct {
	AI      AIConfig
	Scoring ScoringConfig
	RSS     RSSConfig
}

type AIConfig struct {
	QualityMode string `yaml:"quality_mode"`
	Provider    string `yaml:"provider"`
	Model       string `yaml:"model"`
}

type ScoringConfig struct {
	Weights ScoreWeights `yaml:"weights"`
}

type ScoreWeights struct {
	Importance        float64 `yaml:"importance"`
	PersonalRelevance float64 `yaml:"personal_relevance"`
	Credibility       float64 `yaml:"credibility"`
	Novelty           float64 `yaml:"novelty"`
	Freshness         float64 `yaml:"freshness"`
}

type RSSConfig struct {
	Sources   []RSSSource `yaml:"sources"`
	OPMLFiles []string    `yaml:"opml_files"`
}

type RSSSource struct {
	SourceID        string `yaml:"source_id"`
	Name            string `yaml:"name"`
	RSSURL          string `yaml:"rss_url"`
	Domain          string `yaml:"domain"`
	SourceType      string `yaml:"source_type"`
	CredibilityBase int    `yaml:"credibility_base"`
}
