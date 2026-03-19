package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type aiFile struct {
	AI AIConfig `yaml:"ai"`
}

type scoringFile struct {
	Weights ScoreWeights `yaml:"weights"`
}

type rssFile struct {
	RSS RSSConfig `yaml:"rss"`
}

func LoadConfig(dir string) (Config, error) {
	cfg := Config{}

	if err := loadYAML(filepath.Join(dir, "ai.yaml"), &aiFile{AI: cfg.AI}, func(v any) {
		cfg.AI = v.(*aiFile).AI
	}); err != nil {
		return Config{}, err
	}

	if err := loadYAML(filepath.Join(dir, "scoring.yaml"), &scoringFile{Weights: cfg.Scoring.Weights}, func(v any) {
		cfg.Scoring = ScoringConfig{Weights: v.(*scoringFile).Weights}
	}); err != nil {
		return Config{}, err
	}

	if err := loadYAML(filepath.Join(dir, "rss_sources.yaml"), &rssFile{RSS: cfg.RSS}, func(v any) {
		cfg.RSS = v.(*rssFile).RSS
	}); err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadYAML(path string, target any, set func(any)) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	if err := yaml.Unmarshal(b, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	set(target)
	return nil
}

func (c Config) validate() error {
	if c.AI.QualityMode == "" {
		return errors.New("ai quality_mode is required")
	}
	if len(c.RSS.Sources) == 0 {
		return errors.New("at least one rss source is required")
	}
	return nil
}
