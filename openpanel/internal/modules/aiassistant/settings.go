// Package aiassistant is the chat drawer users open from the top bar: it answers questions using the OpenPanel docs and, if the admin allows it, looks up or changes things in the account through the panel's own /api/ routes. The provider is set in openpanel.config's [AI] section.
package aiassistant

import (
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

const (
	defaultDocsURL   = "https://openpanel.com/llms-full.txt"
	defaultName      = "AI Assistant"
	defaultMaxTokens = 2048
	defaultLimit     = 100
	defaultDocsChars = 14000
)

// Settings is the [AI] section of openpanel.config
type Settings struct {
	Provider     string // openrouter, openai, anthropic or custom (any OpenAI-compatible API)
	APIKey       string
	Model        string
	BaseURL      string
	Name         string
	Instructions string
	DocsURL      string
	DocsChars    int // documentation sent with each question, the prompt is what makes CPU-only models slow
	MonthlyLimit int // requests per user per month, 0 is unlimited
	MaxTokens    int
	BrandName    string
	Actions      string // off, read or all
}

func LoadSettings(cfg config.Config) Settings {
	s := Settings{
		Provider:     strings.ToLower(strings.TrimSpace(unquote(cfg.Get("ai_provider", "openrouter")))),
		APIKey:       strings.TrimSpace(unquote(cfg.Get("ai_api_key", ""))),
		Model:        strings.TrimSpace(unquote(cfg.Get("ai_model", ""))),
		BaseURL:      strings.TrimRight(strings.TrimSpace(unquote(cfg.Get("ai_base_url", ""))), "/"),
		Name:         strings.TrimSpace(unquote(cfg.Get("ai_name", ""))),
		Instructions: strings.TrimSpace(unquote(cfg.Get("ai_instructions", ""))),
		DocsURL:      strings.TrimSpace(unquote(cfg.Get("ai_docs_url", ""))),
		DocsChars:    atoiDefault(cfg.Get("ai_docs_chars", ""), defaultDocsChars),
		MonthlyLimit: atoiDefault(cfg.Get("ai_monthly_limit", ""), defaultLimit),
		MaxTokens:    atoiDefault(cfg.Get("ai_max_tokens", ""), defaultMaxTokens),
		BrandName:    strings.TrimSpace(unquote(cfg.Get("brand_name", ""))),
		Actions:      strings.ToLower(strings.TrimSpace(unquote(cfg.Get("ai_actions", modeRead)))),
	}
	if s.Actions != modeOff && s.Actions != modeAll {
		s.Actions = modeRead
	}
	if s.Provider == "" {
		s.Provider = "openrouter"
	}
	if s.Name == "" {
		s.Name = defaultName
	}
	if s.DocsURL == "" {
		s.DocsURL = defaultDocsURL
	}
	if s.DocsChars <= 0 {
		s.DocsChars = defaultDocsChars
	}
	if s.MaxTokens <= 0 {
		s.MaxTokens = defaultMaxTokens
	}
	if s.BrandName == "" {
		s.BrandName = "OpenPanel"
	}
	if s.Model == "" {
		s.Model = defaultModel(s.Provider)
	}
	if s.BaseURL == "" {
		s.BaseURL = defaultBaseURL(s.Provider)
	}
	return s
}

func defaultModel(provider string) string {
	switch provider {
	case "openrouter":
		return "openrouter/auto"
	case "anthropic":
		return "claude-opus-5-5"
	}
	return ""
}

func defaultBaseURL(provider string) string {
	switch provider {
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "openai":
		return "https://api.openai.com/v1"
	}
	return ""
}

// Configured reports whether there's enough set to actually call the provider, custom endpoints (ollama etc.) may not need a key
func (s Settings) Configured() bool {
	switch s.Provider {
	case "openrouter", "openai", "anthropic":
		return s.APIKey != "" && s.Model != ""
	case "custom":
		return s.BaseURL != "" && s.Model != ""
	}
	return false
}

func atoiDefault(v string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(unquote(v)))
	if err != nil || n < 0 {
		return def
	}
	return n
}

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
		return v[1 : len(v)-1]
	}
	return v
}
