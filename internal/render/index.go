package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
)

type featuredCardData struct {
	ID                string
	CardType          string
	BadgeLabel        string
	Category          string
	Title             string
	Summary           string
	CoverImage        string
	ScoreLabel        string
	SourceName        string
	SourceURL         string
	PublishedLabel    string
	TopicTagsJSON     string
	StyleTagsJSON     string
	CognitiveTagsJSON string
}

type indexData struct {
	ThemeID    string
	DateLabel  string
	Keywords   []string
	Featured   []featuredCardData
	Lead       featuredCardData
	Secondary  []featuredCardData
	Supporting []featuredCardData
	Learning   []string
}

func DailyEdition(baseDir string, daily model.DailyEdition) error {
	r, err := NewRenderer(daily.ThemeID)
	if err != nil {
		return err
	}

	root, err := output.EnsureDateDirs(baseDir, daily)
	if err != nil {
		return err
	}

	if err := writeIndex(filepath.Join(root, "index.html"), r, daily); err != nil {
		return err
	}
	if err := writeArticles(filepath.Join(root, "articles"), r, daily); err != nil {
		return err
	}
	if err := copyStaticAssets(filepath.Join(root, "assets")); err != nil {
		return err
	}
	if err := output.WriteJSON(filepath.Join(root, "data", "daily.json"), daily); err != nil {
		return err
	}
	if err := output.WriteJSON(filepath.Join(root, "data", "learning.json"), map[string]any{"learning": daily.Learning}); err != nil {
		return err
	}
	return output.WriteJSON(filepath.Join(root, "meta.json"), map[string]any{
		"generated_at": time.Now().Format(time.RFC3339),
		"featured":     len(daily.Featured),
	})
}

func writeIndex(path string, r *Renderer, daily model.DailyEdition) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	featured := buildFeaturedCards(daily.Featured)
	lead, secondary, supporting := splitHomepageCards(featured)
	data := indexData{
		ThemeID:    r.theme.ID,
		DateLabel:  formatEditionDate(daily.Date),
		Keywords:   daily.Keywords,
		Featured:   featured,
		Lead:       lead,
		Secondary:  secondary,
		Supporting: supporting,
		Learning:   daily.Learning,
	}
	return r.indexTpl.Execute(f, data)
}

func splitHomepageCards(cards []featuredCardData) (featuredCardData, []featuredCardData, []featuredCardData) {
	if len(cards) == 0 {
		return featuredCardData{}, nil, nil
	}
	lead := cards[0]
	if len(cards) == 1 {
		return lead, nil, nil
	}

	secondaryEnd := 2
	if len(cards) > 4 {
		secondaryEnd = 3
	}
	if secondaryEnd > len(cards) {
		secondaryEnd = len(cards)
	}

	secondary := append([]featuredCardData(nil), cards[1:secondaryEnd]...)
	supporting := append([]featuredCardData(nil), cards[secondaryEnd:]...)
	return lead, secondary, supporting
}

func copyStaticAssets(assetDir string) error {
	return output.CopyProjectStaticAssets(assetDir)
}

func buildFeaturedCards(items []model.DailyPick) []featuredCardData {
	cards := make([]featuredCardData, 0, len(items))
	for _, item := range items {
		itemID, ok := normalizeRenderableItemID(item.ID)
		if !ok {
			continue
		}
		cardType := normalizeCardType(item.CardType)
		cards = append(cards, featuredCardData{
			ID:       itemID,
			CardType: cardType,
			BadgeLabel: func() string {
				if cardType == "brief" {
					return "简版"
				}
				return ""
			}(),
			Category:          fallbackString(item.Category, "未分类"),
			Title:             fallbackString(item.Title, "未命名条目"),
			Summary:           fallbackString(item.Summary, "暂无摘要"),
			CoverImage:        coverImageURL(item, "."),
			ScoreLabel:        fmt.Sprintf("%.1f", item.ScoreFinal),
			SourceName:        fallbackString(item.SourceName, "来源待补充"),
			SourceURL:         fallbackString(item.SourceURL, "#"),
			PublishedLabel:    formatPublishedLabel(item.PublishedAt),
			TopicTagsJSON:     marshalStringSliceJSON(item.TopicTags),
			StyleTagsJSON:     marshalStringSliceJSON(item.StyleTags),
			CognitiveTagsJSON: marshalStringSliceJSON(item.CognitiveTags),
		})
	}
	return cards
}

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func formatPublishedLabel(published time.Time) string {
	if published.IsZero() {
		return "时间待定"
	}
	return published.Format("2006-01-02 15:04 MST")
}

func formatEditionDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("2006-01-02")
}

func normalizeCardType(cardType string) string {
	cardType = strings.ToLower(strings.TrimSpace(cardType))
	if cardType == "brief" {
		return "brief"
	}
	return "standard"
}

func marshalStringSliceJSON(values []string) string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return "[]"
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func coverImageURL(item model.DailyPick, relativePrefix string) string {
	local := strings.TrimSpace(item.CoverImageLocal)
	if local == "" {
		return strings.TrimSpace(item.CoverImage)
	}

	normalized := strings.TrimLeft(strings.ReplaceAll(local, "\\", "/"), "/")
	if strings.TrimSpace(relativePrefix) == "" {
		return normalized
	}
	return relativePrefix + "/" + path.Clean(normalized)
}
