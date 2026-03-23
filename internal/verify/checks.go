package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DailyEdition(dir string) error {
	dailyJSONPath := filepath.Join(dir, "data", "daily.json")
	indexPath := filepath.Join(dir, "index.html")
	required := []string{
		indexPath,
		dailyJSONPath,
		filepath.Join(dir, "meta.json"),
	}

	for _, p := range required {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("missing required file: %s", p)
		}
	}

	type insight struct {
		Viewpoint        string `json:"Viewpoint"`
		WhyForYou        string `json:"WhyForYou"`
		TasteGrowthHint  string `json:"TasteGrowthHint"`
		KnowledgeGapHint string `json:"KnowledgeGapHint"`
	}
	type featuredItem struct {
		ID             string  `json:"ID"`
		CardType       string  `json:"CardType"`
		FallbackReason string  `json:"FallbackReason"`
		Summary        string  `json:"Summary"`
		SourceName     string  `json:"SourceName"`
		SourceURL      string  `json:"SourceURL"`
		Insight        insight `json:"Insight"`
	}
	var daily struct {
		Featured []featuredItem `json:"Featured"`
	}

	data, err := os.ReadFile(dailyJSONPath)
	if err != nil {
		return fmt.Errorf("read daily.json: %w", err)
	}
	if err := json.Unmarshal(data, &daily); err != nil {
		return fmt.Errorf("parse daily.json: %w", err)
	}
	if len(daily.Featured) < 10 {
		return fmt.Errorf("featured count below threshold: got %d, want at least 10", len(daily.Featured))
	}
	if err := verifyHomepageHooks(indexPath); err != nil {
		return err
	}

	standardCount := 0
	firstArticlePath := ""
	for _, item := range daily.Featured {
		if strings.TrimSpace(item.ID) == "" {
			return fmt.Errorf("featured item id is required")
		}
		if strings.TrimSpace(item.Summary) == "" {
			return fmt.Errorf("featured item summary is required: %s", item.ID)
		}
		if strings.TrimSpace(item.SourceName) == "" {
			return fmt.Errorf("featured item source_name is required: %s", item.ID)
		}
		if strings.TrimSpace(item.SourceURL) == "" {
			return fmt.Errorf("featured item source_url is required: %s", item.ID)
		}

		cardType := strings.ToLower(strings.TrimSpace(item.CardType))
		switch cardType {
		case "", "standard":
			standardCount++
			if strings.TrimSpace(item.Insight.Viewpoint) == "" {
				return fmt.Errorf("standard card viewpoint is required: %s", item.ID)
			}
			if strings.TrimSpace(item.Insight.WhyForYou) == "" {
				return fmt.Errorf("standard card why_for_you is required: %s", item.ID)
			}
			if strings.TrimSpace(item.Insight.TasteGrowthHint) == "" {
				return fmt.Errorf("standard card taste_growth_hint is required: %s", item.ID)
			}
			if strings.TrimSpace(item.Insight.KnowledgeGapHint) == "" {
				return fmt.Errorf("standard card knowledge_gap_hint is required: %s", item.ID)
			}
		case "brief":
			if strings.TrimSpace(item.FallbackReason) == "" {
				return fmt.Errorf("brief card fallback_reason is required: %s", item.ID)
			}
		default:
			return fmt.Errorf("unknown card type %q for item %s", item.CardType, item.ID)
		}

		articlePath := filepath.Join(dir, "articles", item.ID+".html")
		if _, err := os.Stat(articlePath); err != nil {
			return fmt.Errorf("missing article file: %s", articlePath)
		}
		if firstArticlePath == "" {
			firstArticlePath = articlePath
		}
	}
	if standardCount < 3 {
		return fmt.Errorf("standard card count below threshold: got %d, want at least 3", standardCount)
	}
	if err := verifyArticleHooks(firstArticlePath); err != nil {
		return err
	}
	return nil
}

func verifyHomepageHooks(path string) error {
	html, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read homepage html: %w", err)
	}
	body := string(html)
	for _, required := range []string{
		`data-page-kind="index"`,
		`data-theme-id="`,
		`data-article-id="`,
		`data-card-type="`,
	} {
		if !strings.Contains(body, required) {
			return fmt.Errorf("homepage missing required hook: %s", required)
		}
	}
	return nil
}

func verifyArticleHooks(path string) error {
	html, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read article html: %w", err)
	}
	body := string(html)
	for _, required := range []string{
		`data-theme-id="`,
		`data-feedback-surface="article"`,
		`data-feedback-value="like"`,
		`data-feedback-value="dislike"`,
		`data-feedback-value="bookmark"`,
	} {
		if !strings.Contains(body, required) {
			return fmt.Errorf("article missing required feedback hook: %s", required)
		}
	}
	return nil
}
