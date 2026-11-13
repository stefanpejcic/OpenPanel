// Package botcontrol handles per-domain AI crawler blocking/rate-limiting and traffic breakdown by bot, network, and URL.
package botcontrol

import "strings"

// BotCategory groups detected user-agents into AI, scraper, search, or normal traffic.
type BotCategory string

const (
	CatAITraining   BotCategory = "AI Training"
	CatAISearch     BotCategory = "AI Search"
	CatScraper      BotCategory = "Scraper"
	CatSearchEngine BotCategory = "Search Engine"
	CatHuman        BotCategory = "Browser / Other"
)

// BotSignature describes one known crawler or bot family.
type BotSignature struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Operator string      `json:"operator"`
	Category BotCategory `json:"category"`
	Token    string      `json:"-"`
}

// knownBots lists AI crawlers first (targeted by Caddy rules) followed by search engines for traffic classification.
var knownBots = []BotSignature{
	{ID: "gptbot", Name: "GPTBot", Operator: "OpenAI", Category: CatAITraining, Token: "gptbot"},
	{ID: "chatgpt-user", Name: "ChatGPT-User", Operator: "OpenAI", Category: CatAISearch, Token: "chatgpt-user"},
	{ID: "oai-searchbot", Name: "OAI-SearchBot", Operator: "OpenAI", Category: CatAISearch, Token: "oai-searchbot"},
	{ID: "claudebot", Name: "ClaudeBot", Operator: "Anthropic", Category: CatAITraining, Token: "claudebot"},
	{ID: "claude-web", Name: "Claude-Web", Operator: "Anthropic", Category: CatAISearch, Token: "claude-web"},
	{ID: "anthropic-ai", Name: "anthropic-ai", Operator: "Anthropic", Category: CatAITraining, Token: "anthropic-ai"},
	{ID: "bytespider", Name: "Bytespider", Operator: "ByteDance", Category: CatAITraining, Token: "bytespider"},
	{ID: "ccbot", Name: "CCBot", Operator: "Common Crawl", Category: CatAITraining, Token: "ccbot"},
	{ID: "perplexitybot", Name: "PerplexityBot", Operator: "Perplexity", Category: CatAISearch, Token: "perplexitybot"},
	{ID: "perplexity-user", Name: "Perplexity-User", Operator: "Perplexity", Category: CatAISearch, Token: "perplexity-user"},
	{ID: "amazonbot", Name: "Amazonbot", Operator: "Amazon", Category: CatAITraining, Token: "amazonbot"},
	{ID: "meta-externalagent", Name: "Meta-ExternalAgent", Operator: "Meta", Category: CatAITraining, Token: "meta-externalagent"},
	{ID: "meta-externalfetcher", Name: "Meta-ExternalFetcher", Operator: "Meta", Category: CatAISearch, Token: "meta-externalfetcher"},
	{ID: "facebookbot", Name: "FacebookBot", Operator: "Meta", Category: CatAITraining, Token: "facebookbot"},
	{ID: "diffbot", Name: "Diffbot", Operator: "Diffbot", Category: CatScraper, Token: "diffbot"},
	{ID: "omgilibot", Name: "Omgilibot", Operator: "Webz.io", Category: CatScraper, Token: "omgilibot"},
	{ID: "cohere-ai", Name: "cohere-ai", Operator: "Cohere", Category: CatAITraining, Token: "cohere-ai"},
	{ID: "youbot", Name: "YouBot", Operator: "You.com", Category: CatAISearch, Token: "youbot"},
	{ID: "duckassistbot", Name: "DuckAssistBot", Operator: "DuckDuckGo", Category: CatAISearch, Token: "duckassistbot"},
	{ID: "timpibot", Name: "Timpibot", Operator: "Timpi", Category: CatAISearch, Token: "timpibot"},
	{ID: "scrapy", Name: "Scrapy", Operator: "Scraper", Category: CatScraper, Token: "scrapy"},
	{ID: "googlebot", Name: "Googlebot", Operator: "Google", Category: CatSearchEngine, Token: "googlebot"},
	{ID: "bingbot", Name: "Bingbot", Operator: "Microsoft", Category: CatSearchEngine, Token: "bingbot"},
	{ID: "duckduckbot", Name: "DuckDuckBot", Operator: "DuckDuckGo", Category: CatSearchEngine, Token: "duckduckbot"},
	{ID: "yandexbot", Name: "YandexBot", Operator: "Yandex", Category: CatSearchEngine, Token: "yandexbot"},
	{ID: "baiduspider", Name: "Baiduspider", Operator: "Baidu", Category: CatSearchEngine, Token: "baiduspider"},
}

// KnownAICrawlers returns only the AI and scraper signatures enforced by Caddy rules.
func KnownAICrawlers() []BotSignature {
	out := make([]BotSignature, 0, len(knownBots))
	for _, b := range knownBots {
		if b.Category != CatSearchEngine {
			out = append(out, b)
		}
	}
	return out
}

// classifyUserAgent matches a User-Agent string to a known bot or returns a generic label.
func classifyUserAgent(ua string) (name string, operator string, category BotCategory, isAI bool) {
	lower := strings.ToLower(strings.TrimSpace(ua))
	if lower == "" {
		return "Empty User-Agent", "Unknown", CatScraper, false
	}
	for _, b := range knownBots {
		if strings.Contains(lower, b.Token) {
			return b.Name, b.Operator, b.Category, b.Category != CatSearchEngine
		}
	}
	if strings.Contains(lower, "curl/") || strings.Contains(lower, "python-requests") || strings.Contains(lower, "go-http-client") || strings.Contains(lower, "wget/") {
		return "CLI / Script (" + shortUA(ua) + ")", "Scripted", CatScraper, false
	}
	return "Browser / Other", "-", CatHuman, false
}

func shortUA(ua string) string {
	fields := strings.Fields(ua)
	if len(fields) == 0 {
		return "unknown"
	}
	if len(fields[0]) > 28 {
		return fields[0][:28]
	}
	return fields[0]
}
