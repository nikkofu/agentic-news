package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikkofu/agentic-news/internal/model"
)

type articleData struct {
	ThemeID           string
	ID                string
	EditionDate       string
	CardType          string
	IsBrief           bool
	FallbackReason    string
	Category          string
	Title             string
	Summary           string
	ScoreFinal        float64
	ScoreLabel        string
	CoverImage        string
	SourceName        string
	SourceURL         string
	PublishedAt       string
	Viewpoint         string
	WhyForYou         string
	TasteGrowthHint   string
	KnowledgeGapHint  string
	TopicTags         []string
	StyleTags         []string
	CognitiveTags     []string
	TopicTagsJSON     string
	StyleTagsJSON     string
	CognitiveTagsJSON string
}

func writeArticles(articlesDir string, r *Renderer, daily model.DailyEdition) error {
	for _, item := range daily.Featured {
		itemID, ok := normalizeRenderableItemID(item.ID)
		if !ok {
			continue
		}
		path := filepath.Join(articlesDir, itemID+".html")
		if err := writeSingleArticle(path, r, daily.Date, itemID, item); err != nil {
			return err
		}
	}
	return nil
}

func writeSingleArticle(path string, r *Renderer, editionDate time.Time, itemID string, item model.DailyPick) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	published := item.PublishedAt
	cardType := normalizeCardType(item.CardType)

	data := articleData{
		ThemeID:           r.theme.ID,
		ID:                itemID,
		EditionDate:       formatEditionDate(editionDate),
		CardType:          cardType,
		IsBrief:           cardType == "brief",
		FallbackReason:    fallbackString(item.FallbackReason, "unknown"),
		Category:          fallbackString(item.Category, "未分类"),
		Title:             fallbackString(item.Title, "未命名条目"),
		Summary:           fallbackString(item.Summary, "暂无摘要"),
		ScoreFinal:        item.ScoreFinal,
		ScoreLabel:        fmt.Sprintf("%.1f", item.ScoreFinal),
		CoverImage:        coverImageURL(item, ".."),
		SourceName:        fallbackString(item.SourceName, "来源待补充"),
		SourceURL:         fallbackString(item.SourceURL, "#"),
		PublishedAt:       formatPublishedLabel(published),
		Viewpoint:         fallbackString(item.Insight.Viewpoint, "观点待补充"),
		WhyForYou:         fallbackString(item.Insight.WhyForYou, "这篇内容与你近期关注的主题和理解方式相关。"),
		TasteGrowthHint:   fallbackString(item.Insight.TasteGrowthHint, "反馈后会在这里刷新口味拓展建议。"),
		KnowledgeGapHint:  fallbackString(item.Insight.KnowledgeGapHint, "反馈后会在这里刷新知识补位建议。"),
		TopicTags:         copyStringSlice(item.TopicTags),
		StyleTags:         copyStringSlice(item.StyleTags),
		CognitiveTags:     copyStringSlice(item.CognitiveTags),
		TopicTagsJSON:     marshalStringSliceJSON(item.TopicTags),
		StyleTagsJSON:     marshalStringSliceJSON(item.StyleTags),
		CognitiveTagsJSON: marshalStringSliceJSON(item.CognitiveTags),
	}
	return r.articleTpl.Execute(f, data)
}

func copyStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
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
