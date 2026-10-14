package aiassistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	maxMessages   = 20
	maxMessageLen = 4000
	chatTimeout   = 3 * time.Minute
)

// Register wires the drawer's info and chat endpoints
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "ai_assistant")(h)
	}
	mux.Handle("GET /ai/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInfo(a, w, r) }))
	mux.Handle("POST /ai/chat", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleChat(a, w, r) }))

	// warm the docs index so the first question doesn't wait on the download
	if s := LoadSettings(a.Config); s.Configured() {
		go docs.get(context.Background(), s.DocsURL)
	}
}

func handleInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	s := LoadSettings(a.Config)
	_, username, _, err := auth.Injected(a, r)
	if err != nil || !safeUsername(username) {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"name":       s.Name,
		"configured": s.Configured(),
		"used":       readUsage(username).Requests,
		"limit":      s.MonthlyLimit,
	})
}

type chatRequest struct {
	Messages []message `json:"messages"`
	Page     struct {
		Path  string `json:"path"`
		Title string `json:"title"`
	} `json:"page"`
}

func handleChat(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	s := LoadSettings(a.Config)
	if !s.Configured() {
		web.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "The AI Assistant is not configured yet, please contact your hosting provider."})
		return
	}

	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256*1024)).Decode(&req); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	msgs, err := cleanMessages(req.Messages)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(r.Context(), userID)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	username, _ := injected["current_username"].(string)
	if !safeUsername(username) {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	u, ok := takeRequest(username, s.MonthlyLimit)
	if !ok {
		web.WriteJSON(w, http.StatusTooManyRequests, map[string]any{"error": "You've used all your AI Assistant requests for this month.", "used": u.Requests, "limit": s.MonthlyLimit})
		return
	}

	allowed, _ := injected["user_allowed"].([]string)
	plan, _ := injected["hosting_plan_name"].(string)
	question := msgs[len(msgs)-1].Content
	// the page the user is on helps with vague questions like "how do I use this"
	query := question + " " + strings.ReplaceAll(req.Page.Path, "/", " ") + " " + req.Page.Title
	excerpts := docs.get(r.Context(), s.DocsURL).search(query)
	system := buildSystemPrompt(s, plan, allowed, req.Page.Path, req.Page.Title, excerpts)

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { web.WriteNDJSON(w, flusher, canFlush, v) }

	ctx, cancel := context.WithTimeout(r.Context(), chatTimeout)
	defer cancel()
	err = streamChat(ctx, s, system, msgs, func(text string) { emit(map[string]any{"text": text}) })
	if err != nil {
		log.Printf("AI - chat for %s via %s (%s) failed: %v", username, s.Provider, s.Model, err)
		msg := "The AI Assistant couldn't answer right now, please try again later."
		if errors.Is(err, errRefused) {
			msg = "The AI Assistant can't help with that request."
		}
		emit(map[string]any{"error": msg})
		return
	}
	emit(map[string]any{"done": true, "used": u.Requests, "limit": s.MonthlyLimit})
}

// cleanMessages keeps the last few turns, drops empty ones and makes sure the user spoke last
func cleanMessages(in []message) ([]message, error) {
	var out []message
	for _, m := range in {
		m.Content = strings.TrimSpace(m.Content)
		if m.Content == "" || (m.Role != "user" && m.Role != "assistant") {
			continue
		}
		if len(m.Content) > maxMessageLen {
			if m.Role == "user" {
				return nil, fmt.Errorf("Message is too long, keep it under %d characters.", maxMessageLen)
			}
			m.Content = m.Content[:maxMessageLen]
		}
		out = append(out, m)
	}
	if len(out) > maxMessages {
		out = out[len(out)-maxMessages:]
	}
	// anthropic wants the conversation to start with the user
	for len(out) > 0 && out[0].Role != "user" {
		out = out[1:]
	}
	if len(out) == 0 || out[len(out)-1].Role != "user" {
		return nil, errors.New("Ask a question first.")
	}
	return out, nil
}

func safeUsername(u string) bool {
	return u != "" && !strings.ContainsAny(u, "/\\") && u != "." && u != ".."
}

func buildSystemPrompt(s Settings, plan string, allowed []string, pagePath, pageTitle string, excerpts []chunk) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are %s, the assistant built into the %s hosting control panel. ", s.Name, s.BrandName)
	b.WriteString("You help a hosting customer use their control panel: websites, domains, DNS, email, databases, files, backups, cron jobs, applications and account settings.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Answer from the documentation excerpts below. If they don't cover the question, say so briefly and give your best general guidance, clearly marked as general.\n")
	b.WriteString("- You can't see the account's data and can't make changes. Explain the steps and which page in the panel to use.\n")
	b.WriteString("- The user is a customer without root or SSH access to the server. Never suggest opencli, OpenAdmin, or server-wide changes; for those, tell them to contact their hosting provider.\n")
	b.WriteString("- Only suggest features from the list of features on their plan. If something isn't on it, say their plan doesn't include it and suggest contacting their hosting provider.\n")
	fmt.Fprintf(&b, "- Call the panel %s. Don't link to external documentation websites; refer to pages by their menu names instead.\n", s.BrandName)
	b.WriteString("- Be concise. Use short Markdown: numbered steps, `code` for values, and code blocks for commands or config.\n")
	b.WriteString("- Reply in the language the user writes in.\n")
	b.WriteString("- If the user asks about anything unrelated to web hosting or this panel, politely steer back.\n")
	if s.Instructions != "" {
		fmt.Fprintf(&b, "\nInstructions from the hosting provider:\n%s\n", s.Instructions)
	}

	b.WriteString("\nAbout this user:\n")
	if plan != "" {
		fmt.Fprintf(&b, "- Hosting plan: %s\n", plan)
	}
	features := append([]string{}, allowed...)
	sort.Strings(features)
	fmt.Fprintf(&b, "- Features on their plan: %s\n", strings.Join(features, ", "))
	if pagePath != "" {
		fmt.Fprintf(&b, "- They are currently on the page %s", pagePath)
		if pageTitle != "" {
			fmt.Fprintf(&b, " (%s)", pageTitle)
		}
		b.WriteString("\n")
	}

	if len(excerpts) > 0 {
		b.WriteString("\nDocumentation excerpts:\n")
		for _, c := range excerpts {
			fmt.Fprintf(&b, "\n<doc path=%q title=%q>\n%s\n</doc>\n", c.Path, c.Title, c.Text)
		}
	}
	return b.String()
}
