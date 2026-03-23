package run

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nikkofu/agentic-news/internal/analyze"
	"github.com/nikkofu/agentic-news/internal/config"
	"github.com/nikkofu/agentic-news/internal/content"
	"github.com/nikkofu/agentic-news/internal/feedback"
	"github.com/nikkofu/agentic-news/internal/model"
	"github.com/nikkofu/agentic-news/internal/output"
	"github.com/nikkofu/agentic-news/internal/profile"
	"github.com/nikkofu/agentic-news/internal/rank"
	"github.com/nikkofu/agentic-news/internal/render"
	"github.com/nikkofu/agentic-news/internal/rss"
	"github.com/nikkofu/agentic-news/internal/verify"
)

type DryRunRequest struct {
	ConfigDir string
	OutputDir string
	StateDir  string
	Date      time.Time
	Mode      string
	Theme     string
}

type DryRunResult struct {
	OutputRoot    string
	PackageRoot   string
	FeaturedCount int
	UsedFallback  bool
}

const maxProcessedCandidates = 24

type DryRunHooks struct {
	LoadConfig     func(string) (config.Config, error)
	FetchFeeds     func(context.Context, []string) ([]model.RawItem, error)
	Dedupe         func([]model.RawItem) []model.RawItem
	Extract        func(context.Context, model.RawItem) (model.Article, content.ExtractionStatus, error)
	AnalyzeInsight func(context.Context, model.Article, profile.UserProfile) (model.Insight, error)
	LoadProfile    func(stateDir string) (profile.UserProfile, error)
	ArchiveImage   func(context.Context, string, string, string, []model.ArticleImage) (string, error)
	Rank           func([]model.DailyPick) []model.DailyPick
	Render         func(string, model.DailyEdition) error
	Verify         func(string) error
	Now            func() time.Time
}

func RunDryPipeline(ctx context.Context, req DryRunRequest, hooks DryRunHooks) (DryRunResult, error) {
	if strings.TrimSpace(req.Mode) == "" {
		return DryRunResult{}, fmt.Errorf("mode is required")
	}
	req.Theme = model.NormalizeThemeID(req.Theme)

	now := time.Now()
	if hooks.Now != nil {
		now = hooks.Now()
	}
	if req.Date.IsZero() {
		req.Date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	if strings.TrimSpace(req.ConfigDir) == "" {
		req.ConfigDir = "config"
	}
	if strings.TrimSpace(req.OutputDir) == "" {
		req.OutputDir = "output"
	}
	if strings.TrimSpace(req.StateDir) == "" {
		req.StateDir = "state"
	}

	loadConfig := hooks.LoadConfig
	if loadConfig == nil {
		loadConfig = config.LoadConfig
	}
	fetchFeeds := hooks.FetchFeeds
	if fetchFeeds == nil {
		fetchFeeds = rss.FetchFeeds
	}
	dedupe := hooks.Dedupe
	if dedupe == nil {
		dedupe = rss.Dedupe
	}
	extract := hooks.Extract
	if extract == nil {
		extract = content.ExtractArticle
	}
	analyzeInsight := hooks.AnalyzeInsight
	if analyzeInsight == nil {
		analyzeInsight = analyze.RunPipeline
	}
	loadProfile := hooks.LoadProfile
	if loadProfile == nil {
		loadProfile = func(stateDir string) (profile.UserProfile, error) {
			return feedback.NewStore(stateDir).ReadProfileSnapshot()
		}
	}
	archiveImage := hooks.ArchiveImage
	if archiveImage == nil {
		archiveImage = output.ArchivePreferredImage
	}
	rankItems := hooks.Rank
	if rankItems == nil {
		rankItems = rank.RankItems
	}
	renderEdition := hooks.Render
	if renderEdition == nil {
		renderEdition = render.DailyEdition
	}
	verifyEdition := hooks.Verify
	if verifyEdition == nil {
		verifyEdition = verify.DailyEdition
	}

	cfg, err := loadConfig(req.ConfigDir)
	if err != nil {
		return DryRunResult{}, fmt.Errorf("load config: %w", err)
	}

	feedURLs := configuredFeedURLs(cfg)
	if len(feedURLs) == 0 {
		return DryRunResult{}, fmt.Errorf("no rss urls configured")
	}

	rawItems, err := fetchFeeds(ctx, feedURLs)
	if err != nil {
		return DryRunResult{}, fmt.Errorf("fetch feeds: %w", err)
	}

	rawItems = dedupe(rawItems)
	if len(rawItems) == 0 {
		return DryRunResult{}, fmt.Errorf("no candidate items available after dedupe")
	}
	rawItems = limitCandidateWindow(rawItems, maxProcessedCandidates)

	weights := configuredWeights(cfg)
	loadedProfile, err := loadProfile(req.StateDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return DryRunResult{}, fmt.Errorf("load profile snapshot: %w", err)
	}
	analysisProfile := loadedProfile
	scoringProfile := loadedProfile
	if err != nil || isEmptyUserProfile(loadedProfile) {
		analysisProfile = profile.UserProfile{}
		scoringProfile = defaultUserProfile()
	}
	fallbackTopicTags := resolveTags(scoringProfile.FocusTopics, scoringProfile.TopicAffinity, []string{"technology", "policy", "finance"})
	fallbackStyleTags := resolveTags(scoringProfile.PreferredStyles, scoringProfile.StyleAffinity, []string{"data-driven"})
	fallbackCognitiveTags := resolveTags(scoringProfile.CognitivePreferences, scoringProfile.CognitiveAffinity, []string{"systems thinking"})

	sourcesByID := mapSourcesByID(cfg.RSS.Sources)
	usedFallback := ShouldFallback(now)
	editionRoot, err := output.ResetDateDir(req.OutputDir, model.DailyEdition{Date: req.Date, ThemeID: req.Theme})
	if err != nil {
		return DryRunResult{}, fmt.Errorf("prepare edition root: %w", err)
	}

	buildPick := func(i int, item model.RawItem) model.DailyPick {
		source := resolveSource(sourcesByID, item)
		pickID := defaultPickID(item, i)
		category := categoryForItem(source, i)
		baseSummary := firstNonEmpty(item.RawContent, item.Title, "暂无摘要")
		topicTags, styleTags, cognitiveTags := resolvePickTags(source, model.Article{}, category, fallbackTopicTags, fallbackStyleTags, fallbackCognitiveTags)

		article, status, err := extract(ctx, item)
		if err != nil {
			return buildBriefPick(
				pickID,
				category,
				item,
				source,
				baseSummary,
				firstNonEmpty(status.FallbackReason, "extract_failed"),
				40+float64(i%10),
				"",
				"",
				topicTags,
				styleTags,
				cognitiveTags,
			)
		}
		coverImageLocal, archiveErr := archiveImage(ctx, editionRoot, pickID, article.CoverImage, article.ImageCandidates)
		if archiveErr != nil {
			coverImageLocal = ""
		}
		topicTags, styleTags, cognitiveTags = resolvePickTags(source, article, category, fallbackTopicTags, fallbackStyleTags, fallbackCognitiveTags)
		if !status.StandardEligible {
			briefSummary := firstNonEmpty(article.Excerpt, summarizeText(article.ContentText, 180), baseSummary)
			return buildBriefPick(
				pickID,
				category,
				item,
				source,
				briefSummary,
				firstNonEmpty(status.FallbackReason, "extraction_degraded"),
				45+float64(i%10),
				article.CoverImage,
				coverImageLocal,
				topicTags,
				styleTags,
				cognitiveTags,
			)
		}

		insight, err := analyzeInsight(ctx, article, analysisProfile)
		if err != nil {
			briefSummary := firstNonEmpty(article.Excerpt, summarizeText(article.ContentText, 180), baseSummary)
			return buildBriefPick(
				pickID,
				category,
				item,
				source,
				briefSummary,
				"analysis_failed",
				50+float64(i%10),
				article.CoverImage,
				coverImageLocal,
				topicTags,
				styleTags,
				cognitiveTags,
			)
		}

		signals := scoreSignalsForItem(source, item, i, now, scoringProfile, topicTags, styleTags, cognitiveTags)
		return model.DailyPick{
			ID:              pickID,
			CardType:        "standard",
			Category:        category,
			Title:           fallbackTitle(item.Title),
			Summary:         firstNonEmpty(insight.SummaryBrief, article.Excerpt, summarizeText(article.ContentText, 180), baseSummary),
			ScoreFinal:      rank.ScoreItem(signals, weights, category),
			CoverImage:      article.CoverImage,
			CoverImageLocal: coverImageLocal,
			SourceName:      sourceName(source, item),
			SourceURL:       sourceURL(item),
			PublishedAt:     item.PublishedAt,
			ReadingTimeMin:  estimateReadingTime(article.ContentText),
			TopicTags:       slices.Clone(topicTags),
			StyleTags:       slices.Clone(styleTags),
			CognitiveTags:   slices.Clone(cognitiveTags),
			Insight:         insight,
		}
	}

	type pickResult struct {
		index int
		pick  model.DailyPick
	}

	workerLimit := 6
	if len(rawItems) < workerLimit {
		workerLimit = len(rawItems)
	}
	if workerLimit < 1 {
		workerLimit = 1
	}

	resultsCh := make(chan pickResult, len(rawItems))
	sem := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup

	for i, item := range rawItems {
		wg.Add(1)
		go func(i int, item model.RawItem) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			resultsCh <- pickResult{
				index: i,
				pick:  buildPick(i, item),
			}
		}(i, item)
	}

	wg.Wait()
	close(resultsCh)

	pickResults := make([]pickResult, 0, len(rawItems))
	for result := range resultsCh {
		pickResults = append(pickResults, result)
	}
	sort.Slice(pickResults, func(i, j int) bool {
		return pickResults[i].index < pickResults[j].index
	})

	picks := make([]model.DailyPick, 0, len(pickResults))
	for _, result := range pickResults {
		picks = append(picks, result.pick)
	}

	if len(picks) == 0 {
		return DryRunResult{}, fmt.Errorf("no publishable picks could be constructed")
	}

	ranked := rankItems(picks)
	featured := ranked
	if len(featured) > 10 {
		featured = featured[:10]
	}

	usedFallback = usedFallback || containsBrief(featured)
	daily := buildDailyEdition(req.Date, req.Theme, featured, now, usedFallback)

	if err := renderEdition(req.OutputDir, daily); err != nil {
		return DryRunResult{}, fmt.Errorf("render: %w", err)
	}

	packageRoot, err := output.WriteEditionPackage(req.OutputDir, editionRoot, daily)
	if err != nil {
		return DryRunResult{}, fmt.Errorf("write edition package: %w", err)
	}

	root := output.DatePath(req.OutputDir, daily)
	if err := verifyEdition(root); err != nil {
		return DryRunResult{}, fmt.Errorf("verify: %w", err)
	}

	return DryRunResult{
		OutputRoot:    root,
		PackageRoot:   packageRoot,
		FeaturedCount: len(daily.Featured),
		UsedFallback:  usedFallback,
	}, nil
}

func ShouldFallback(now time.Time) bool {
	deadlineGuard := time.Date(now.Year(), now.Month(), now.Day(), 6, 50, 0, 0, now.Location())
	return !now.Before(deadlineGuard)
}

func buildDailyEdition(date time.Time, themeID string, picks []model.DailyPick, now time.Time, usedFallback bool) model.DailyEdition {
	learning := []string{
		"Track cross-source confirmation before upgrading an item to a long-lived belief.",
		"Review the highest-scoring item and compare the source claim with the AI viewpoint.",
	}
	if usedFallback {
		learning = append([]string{"本期包含降级简版卡片：请优先阅读标准卡片，再将简版作为补充线索。"}, learning...)
	}

	return model.DailyEdition{
		Date:        date,
		ThemeID:     themeID,
		Keywords:    []string{"AI", "Policy", "Markets"},
		Featured:    picks,
		Learning:    learning,
		GeneratedAt: now,
	}
}

func configuredFeedURLs(cfg config.Config) []string {
	urls := make([]string, 0, len(cfg.RSS.Sources))
	for _, source := range cfg.RSS.Sources {
		if trimmed := strings.TrimSpace(source.RSSURL); trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	return urls
}

func configuredWeights(cfg config.Config) rank.Weights {
	weights := rank.Weights{
		Importance:        cfg.Scoring.Weights.Importance,
		PersonalRelevance: cfg.Scoring.Weights.PersonalRelevance,
		Credibility:       cfg.Scoring.Weights.Credibility,
		Novelty:           cfg.Scoring.Weights.Novelty,
		Freshness:         cfg.Scoring.Weights.Freshness,
	}
	if weights == (rank.Weights{}) {
		return rank.DefaultWeights()
	}
	return weights
}

func mapSourcesByID(sources []config.RSSSource) map[string]config.RSSSource {
	out := make(map[string]config.RSSSource, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source.SourceID) != "" {
			out[source.SourceID] = source
		}
		if strings.TrimSpace(source.RSSURL) != "" {
			out[source.RSSURL] = source
		}
	}
	return out
}

func resolveSource(sourcesByID map[string]config.RSSSource, item model.RawItem) config.RSSSource {
	if source, ok := sourcesByID[item.SourceID]; ok {
		return source
	}
	return config.RSSSource{
		SourceID: item.SourceID,
		Name:     item.Author,
	}
}

func categoryForItem(source config.RSSSource, index int) string {
	switch strings.ToLower(strings.TrimSpace(source.SourceType)) {
	case "media", "tech", "technology":
		return "tech"
	case "policy", "politics", "public":
		return "policy"
	case "finance", "business":
		return "finance"
	default:
		return defaultCategory(index)
	}
}

func scoreSignalsForItem(
	source config.RSSSource,
	item model.RawItem,
	index int,
	now time.Time,
	userProfile profile.UserProfile,
	topicTags []string,
	styleTags []string,
	cognitiveTags []string,
) model.ScoreSignals {
	credibility := float64(source.CredibilityBase)
	if credibility == 0 {
		credibility = 72
	}
	sourceLabel := sourceName(source, item)
	return model.ScoreSignals{
		Importance:        78 - float64(index%6),
		PersonalRelevance: rank.ScorePersonalRelevance(userProfile, topicTags, styleTags, cognitiveTags, sourceLabel),
		Credibility:       credibility,
		Novelty:           70 - float64(index%5),
		Freshness:         freshnessScore(now, item.PublishedAt),
	}
}

func buildBriefPick(
	id,
	category string,
	item model.RawItem,
	source config.RSSSource,
	summary,
	reason string,
	score float64,
	coverImage string,
	coverImageLocal string,
	topicTags []string,
	styleTags []string,
	cognitiveTags []string,
) model.DailyPick {
	return model.DailyPick{
		ID:              id,
		CardType:        "brief",
		FallbackReason:  firstNonEmpty(reason, "fallback"),
		Category:        category,
		Title:           fallbackTitle(item.Title),
		Summary:         firstNonEmpty(summary, "暂无摘要"),
		ScoreFinal:      score,
		CoverImage:      coverImage,
		CoverImageLocal: coverImageLocal,
		SourceName:      sourceName(source, item),
		SourceURL:       sourceURL(item),
		PublishedAt:     item.PublishedAt,
		TopicTags:       slices.Clone(topicTags),
		StyleTags:       slices.Clone(styleTags),
		CognitiveTags:   slices.Clone(cognitiveTags),
	}
}

func defaultPickID(item model.RawItem, index int) string {
	if trimmed := strings.TrimSpace(item.ItemID); trimmed != "" {
		return trimmed
	}
	return fmt.Sprintf("pick-%02d", index+1)
}

func fallbackTitle(title string) string {
	return firstNonEmpty(title, "未命名条目")
}

func summarizeText(text string, max int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if len(text) <= max {
		return text
	}
	return strings.TrimSpace(text[:max])
}

func sourceName(source config.RSSSource, item model.RawItem) string {
	return firstNonEmpty(source.Name, item.Author, "Unknown Source")
}

func sourceURL(item model.RawItem) string {
	return firstNonEmpty(item.URL, "#")
}

func freshnessScore(now time.Time, published time.Time) float64 {
	if published.IsZero() {
		return 50
	}
	if now.IsZero() {
		now = time.Now()
	}
	hours := now.Sub(published).Hours()
	if hours < 0 {
		hours = 0
	}
	switch {
	case hours <= 6:
		return 90
	case hours <= 24:
		return 80
	case hours <= 48:
		return 70
	default:
		return 60
	}
}

func estimateReadingTime(text string) int {
	words := len(strings.Fields(strings.TrimSpace(text)))
	if words == 0 {
		return 1
	}
	minutes := words / 180
	if words%180 != 0 {
		minutes++
	}
	if minutes < 1 {
		return 1
	}
	return minutes
}

func containsBrief(items []model.DailyPick) bool {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.CardType), "brief") {
			return true
		}
	}
	return false
}

func defaultCategory(i int) string {
	switch i % 3 {
	case 0:
		return "tech"
	case 1:
		return "policy"
	default:
		return "finance"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func defaultUserProfile() profile.UserProfile {
	return profile.UserProfile{
		FocusTopics:          []string{"technology", "policy", "finance"},
		PreferredStyles:      []string{"data-driven"},
		CognitivePreferences: []string{"systems thinking"},
		TopicAffinity: map[string]float64{
			"technology": 8,
			"policy":     6,
			"finance":    5,
		},
		StyleAffinity: map[string]float64{
			"data-driven": 6,
		},
		CognitiveAffinity: map[string]float64{
			"systems thinking": 6,
		},
	}
}

func isEmptyUserProfile(p profile.UserProfile) bool {
	return len(p.FocusTopics) == 0 &&
		len(p.PreferredStyles) == 0 &&
		len(p.CognitivePreferences) == 0 &&
		len(p.TopicAffinity) == 0 &&
		len(p.StyleAffinity) == 0 &&
		len(p.CognitiveAffinity) == 0 &&
		len(p.SourceAffinity) == 0 &&
		len(p.NegativeSignals) == 0 &&
		len(p.RecentFeedbackSummary) == 0
}

func resolveTags(primary []string, affinity map[string]float64, fallback []string) []string {
	primary = normalizeTags(primary)
	if len(primary) > 0 {
		return primary
	}
	if len(affinity) > 0 {
		ranked := topPositiveAffinityKeys(affinity, 3)
		if len(ranked) > 0 {
			return ranked
		}
	}
	return normalizeTags(fallback)
}

func limitCandidateWindow(items []model.RawItem, limit int) []model.RawItem {
	if limit <= 0 || len(items) <= limit {
		return items
	}

	trimmed := append([]model.RawItem(nil), items...)
	sort.Slice(trimmed, func(i, j int) bool {
		if trimmed[i].PublishedAt.Equal(trimmed[j].PublishedAt) {
			if trimmed[i].ItemID == trimmed[j].ItemID {
				return trimmed[i].URL < trimmed[j].URL
			}
			return trimmed[i].ItemID < trimmed[j].ItemID
		}
		return trimmed[i].PublishedAt.After(trimmed[j].PublishedAt)
	})
	return trimmed[:limit]
}

func normalizeTags(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func topPositiveAffinityKeys(values map[string]float64, limit int) []string {
	if len(values) == 0 || limit <= 0 {
		return nil
	}
	type scorePair struct {
		key   string
		score float64
	}
	pairs := make([]scorePair, 0, len(values))
	for key, score := range values {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" || score <= 0 {
			continue
		}
		pairs = append(pairs, scorePair{key: trimmed, score: score})
	}
	if len(pairs) == 0 {
		return nil
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return strings.ToLower(pairs[i].key) < strings.ToLower(pairs[j].key)
		}
		return pairs[i].score > pairs[j].score
	})
	if len(pairs) < limit {
		limit = len(pairs)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, pairs[i].key)
	}
	return out
}

func resolvePickTags(
	source config.RSSSource,
	article model.Article,
	category string,
	fallbackTopicTags []string,
	fallbackStyleTags []string,
	fallbackCognitiveTags []string,
) ([]string, []string, []string) {
	topicCandidates := normalizeTags(append([]string{}, article.Keywords...))
	if len(topicCandidates) == 0 {
		topicCandidates = normalizeTags(append(topicCandidates, article.CategoryPrimary))
	}
	if len(topicCandidates) == 0 {
		sourceType := strings.TrimSpace(source.SourceType)
		if sourceType != "" {
			topicCandidates = normalizeTags(append(topicCandidates, sourceType, category))
		}
	}
	styleCandidates := normalizeTags(append([]string{}, article.CategorySecondary...))
	cognitiveCandidates := normalizeTags(append([]string{}, article.CategorySecondary...))

	topicTags := normalizeTags(fallbackTopicTags)
	if len(topicCandidates) > 0 {
		topicTags = topicCandidates
	}
	styleTags := normalizeTags(fallbackStyleTags)
	if len(styleCandidates) > 0 {
		styleTags = styleCandidates
	}
	cognitiveTags := normalizeTags(fallbackCognitiveTags)
	if len(cognitiveCandidates) > 0 {
		cognitiveTags = cognitiveCandidates
	}

	return topicTags, styleTags, cognitiveTags
}
