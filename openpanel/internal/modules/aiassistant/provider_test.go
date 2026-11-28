package aiassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"github.com/anthropics/anthropic-sdk-go"
)

// fakeChat answers /chat/completions, reject decides per request body whether to send a 400 instead
func fakeChat(t *testing.T, reject func(body map[string]any) string, stream string) (*httptest.Server, *[]map[string]any) {
	var seen []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		seen = append(seen, body)
		if msg := reject(body); msg != "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":{"message":%q}}`, msg)
			return
		}
		fmt.Fprint(w, stream)
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

const okStream = "data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\ndata: [DONE]\n\n"

func TestOpenAISendsMaxCompletionTokens(t *testing.T) {
	srv, seen := fakeChat(t, func(b map[string]any) string {
		if _, ok := b["max_tokens"]; ok {
			return "Unsupported parameter: 'max_tokens' is not supported with this model. Use 'max_completion_tokens' instead."
		}
		return ""
	}, okStream)
	out, err := turnOpenAI(context.Background(), Settings{Provider: "openai", BaseURL: srv.URL, Model: "m", MaxTokens: 100}, "", []turnMsg{{Role: "user", Text: "q"}}, nil, func(string) {})
	if err != nil || out.Text != "hi" {
		t.Fatalf("got %q, %v", out.Text, err)
	}
	if len(*seen) != 1 || (*seen)[0]["max_completion_tokens"] != float64(100) {
		t.Fatalf("requests: %v", *seen)
	}
}

func TestCustomRetriesWithOtherTokensField(t *testing.T) {
	srv, seen := fakeChat(t, func(b map[string]any) string {
		if _, ok := b["max_tokens"]; ok {
			return "Unsupported parameter: 'max_tokens'"
		}
		return ""
	}, okStream)
	out, err := turnOpenAI(context.Background(), Settings{Provider: "custom", BaseURL: srv.URL, Model: "m", MaxTokens: 100}, "", []turnMsg{{Role: "user", Text: "q"}}, nil, func(string) {})
	if err != nil || out.Text != "hi" || len(*seen) != 2 {
		t.Fatalf("got %q, %v after %d requests", out.Text, err, len(*seen))
	}
}

func TestOtherErrorsAreNotRetried(t *testing.T) {
	srv, seen := fakeChat(t, func(map[string]any) string { return "The model `nope` does not exist" }, okStream)
	_, err := turnOpenAI(context.Background(), Settings{Provider: "openai", BaseURL: srv.URL, Model: "nope", MaxTokens: 100}, "", []turnMsg{{Role: "user", Text: "q"}}, nil, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "does not exist") || len(*seen) != 1 {
		t.Fatalf("got %v after %d requests", err, len(*seen))
	}
}

const lunaToolsError = "Function tools with reasoning_effort are not supported for gpt-6-luna in /v1/chat/completions. To use function tools, use /v1/responses or set reasoning_effort to 'none'."

var testTools = []toolSpec{{Name: "list_domains", Description: "d", Properties: map[string]any{}}}

func TestReasoningEffortRetriedWithNoneWhenUnset(t *testing.T) {
	srv, seen := fakeChat(t, func(b map[string]any) string {
		if b["reasoning_effort"] != "none" {
			return lunaToolsError
		}
		return ""
	}, okStream)
	out, err := turnOpenAI(context.Background(), Settings{Provider: "openai", BaseURL: srv.URL, Model: "gpt-6-luna", MaxTokens: 100}, "", []turnMsg{{Role: "user", Text: "q"}}, testTools, func(string) {})
	if err != nil || out.Text != "hi" || len(*seen) != 2 {
		t.Fatalf("got %q, %v after %d requests", out.Text, err, len(*seen))
	}
	if _, ok := (*seen)[0]["reasoning_effort"]; ok {
		t.Fatalf("first request should leave reasoning_effort out: %v", (*seen)[0])
	}
}

func TestReasoningEffortSentAndNotOverridden(t *testing.T) {
	srv, seen := fakeChat(t, func(map[string]any) string { return lunaToolsError }, okStream)
	_, err := turnOpenAI(context.Background(), Settings{Provider: "openai", BaseURL: srv.URL, Model: "gpt-6-luna", MaxTokens: 100, Reasoning: "high"}, "", []turnMsg{{Role: "user", Text: "q"}}, testTools, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "reasoning_effort") || len(*seen) != 1 || (*seen)[0]["reasoning_effort"] != "high" {
		t.Fatalf("got %v after %v", err, *seen)
	}
}

func TestBothRetriesInOneQuestion(t *testing.T) {
	srv, seen := fakeChat(t, func(b map[string]any) string {
		if _, ok := b["max_tokens"]; ok {
			return "Unsupported parameter: 'max_tokens'"
		}
		if b["reasoning_effort"] != "none" {
			return lunaToolsError
		}
		return ""
	}, okStream)
	out, err := turnOpenAI(context.Background(), Settings{Provider: "custom", BaseURL: srv.URL, Model: "m", MaxTokens: 100}, "", []turnMsg{{Role: "user", Text: "q"}}, testTools, func(string) {})
	if err != nil || out.Text != "hi" || len(*seen) != 3 {
		t.Fatalf("got %q, %v after %d requests", out.Text, err, len(*seen))
	}
}

func TestReasoningSettingNormalized(t *testing.T) {
	for in, want := range map[string]string{"": "", "None": "none", "xhigh": "xhigh", "max": "", "bogus": ""} {
		if got := LoadSettings(config.Config{"ai_reasoning_effort": in}).Reasoning; got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
	if anthropicEffort("") != anthropic.BetaOutputConfigEffortLow || anthropicEffort("none") != anthropic.BetaOutputConfigEffortLow || anthropicEffort("high") != anthropic.BetaOutputConfigEffortHigh {
		t.Error("anthropic effort mapping")
	}
}

func TestOutOfTokensBeforeAnswerIsAnError(t *testing.T) {
	srv, _ := fakeChat(t, func(map[string]any) string { return "" }, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"length\"}]}\n\ndata: [DONE]\n\n")
	_, err := turnOpenAI(context.Background(), Settings{Provider: "openai", BaseURL: srv.URL, Model: "m", MaxTokens: 100}, "", []turnMsg{{Role: "user", Text: "q"}}, nil, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "Max tokens per answer") {
		t.Fatalf("got %v", err)
	}
}

func TestRefundAndLastError(t *testing.T) {
	usersDir = t.TempDir()
	takeRequest("u", 0)
	takeRequest("u", 0)
	refundRequest("u")
	setLastError("u", "provider returned 401: bad key")
	f := loadUsage("u")
	if f.month().Requests != 1 || f.LastError != "provider returned 401: bad key" || f.LastErrorAt == "" {
		t.Fatalf("got %+v %+v", f, f.month())
	}
	setLastError("u", "")
	if f = loadUsage("u"); f.LastError != "" || f.LastErrorAt != "" {
		t.Fatalf("not cleared: %+v", f)
	}
}

// the same error again moves last_error_at, OpenAdmin showed the first failure's time for every later one (discussion #1392)
func TestRepeatedLastErrorUpdatesTime(t *testing.T) {
	usersDir = t.TempDir()
	setLastError("u", "provider returned 400: rejected")
	f := loadUsage("u")
	f.LastErrorAt = "2026-11-27T18:05:00Z"
	saveUsage("u", f)
	setLastError("u", "provider returned 400: rejected")
	if f = loadUsage("u"); f.LastErrorAt == "2026-11-27T18:05:00Z" || f.LastErrorAt == "" {
		t.Fatalf("time not moved on a repeated error: %+v", f)
	}
}
