package aiassistant

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
)

// message is one entry of the chat history the browser sends back
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type toolSpec struct {
	Name        string
	Description string
	Properties  map[string]any
	Required    []string
}

type toolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

type toolResult struct {
	ID      string
	Content string
	IsError bool
}

// turnMsg is one turn of the conversation sent to the provider, with tool calls and results
type turnMsg struct {
	Role    string
	Text    string
	Calls   []toolCall
	Results []toolResult
}

type turnOut struct {
	Text         string
	Calls        []toolCall
	InputTokens  int64
	OutputTokens int64
}

var errRefused = errors.New("the model declined to answer this")

// chatTurn sends the conversation to the configured provider, streams text through onText and returns any tool calls the model made, tests swap it for a fake model
var chatTurn = func(ctx context.Context, s Settings, system string, conv []turnMsg, tools []toolSpec, onText func(string)) (turnOut, error) {
	if s.Provider == "anthropic" {
		return turnAnthropic(ctx, s, system, conv, tools, onText)
	}
	return turnOpenAI(ctx, s, system, conv, tools, onText)
}

func turnAnthropic(ctx context.Context, s Settings, system string, conv []turnMsg, tools []toolSpec, onText func(string)) (turnOut, error) {
	opts := []option.RequestOption{option.WithAPIKey(s.APIKey)}
	if s.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(s.BaseURL))
	}
	client := anthropic.NewClient(opts...)

	params := anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(s.Model),
		MaxTokens: int64(s.MaxTokens),
		// cached so the follow-up calls after each action don't pay for the docs again
		System: []anthropic.BetaTextBlockParam{{Text: system, CacheControl: anthropic.NewBetaCacheControlEphemeralParam()}},
		// a policy decline gets re-served by the server's default fallback model, models without fallbacks just run alone
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		Fallbacks: anthropic.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()},
	}
	if supportsEffort(s.Model) {
		params.OutputConfig = anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortLow}
	}
	for _, t := range tools {
		params.Tools = append(params.Tools, anthropic.BetaToolUnionParam{OfTool: &anthropic.BetaToolParam{
			Name:        t.Name,
			Description: anthropic.String(t.Description),
			InputSchema: anthropic.BetaToolInputSchemaParam{Properties: t.Properties, Required: t.Required},
		}})
	}
	for _, m := range conv {
		var blocks []anthropic.BetaContentBlockParamUnion
		for _, r := range m.Results {
			blocks = append(blocks, anthropic.NewBetaToolResultBlock(r.ID, r.Content, r.IsError))
		}
		if m.Text != "" {
			blocks = append(blocks, anthropic.NewBetaTextBlock(m.Text))
		}
		for _, c := range m.Calls {
			blocks = append(blocks, anthropic.NewBetaToolUseBlock(c.ID, c.Input, c.Name))
		}
		if len(blocks) == 0 {
			continue
		}
		role := anthropic.BetaMessageParamRoleUser
		if m.Role == "assistant" {
			role = anthropic.BetaMessageParamRoleAssistant
		}
		params.Messages = append(params.Messages, anthropic.BetaMessageParam{Role: role, Content: blocks})
	}

	stream := client.Beta.Messages.NewStreaming(ctx, params)
	defer stream.Close()
	var msg anthropic.BetaMessage
	for stream.Next() {
		ev := stream.Current()
		_ = msg.Accumulate(ev)
		if delta, ok := ev.AsAny().(anthropic.BetaRawContentBlockDeltaEvent); ok {
			if d, ok := delta.Delta.AsAny().(anthropic.BetaTextDelta); ok {
				onText(d.Text)
			}
		}
	}
	if err := stream.Err(); err != nil {
		var apiErr *anthropic.Error
		if errors.As(err, &apiErr) {
			return turnOut{}, fmt.Errorf("provider returned %d", apiErr.StatusCode)
		}
		return turnOut{}, err
	}

	out := turnOut{
		InputTokens:  msg.Usage.InputTokens + msg.Usage.CacheCreationInputTokens + msg.Usage.CacheReadInputTokens,
		OutputTokens: msg.Usage.OutputTokens,
	}
	if msg.StopReason == anthropic.BetaStopReasonRefusal {
		return out, errRefused
	}
	for _, b := range msg.Content {
		switch b.Type {
		case "text":
			out.Text += b.Text
		case "tool_use":
			out.Calls = append(out.Calls, toolCall{ID: b.ID, Name: b.Name, Input: b.Input})
		}
	}
	return out, nil
}

// supportsEffort is true for the claude models that take output_config.effort, older ones 400 on it
func supportsEffort(model string) bool {
	for _, p := range []string{"claude-fable-", "claude-mythos-", "claude-opus-5", "claude-sonnet-5", "claude-haiku-5", "claude-opus-4-6", "claude-opus-4-7", "claude-opus-4-8", "claude-sonnet-4-6"} {
		if strings.HasPrefix(model, p) {
			return true
		}
	}
	return false
}

// turnOpenAI talks to openrouter, openai or anything else that speaks /chat/completions
func turnOpenAI(ctx context.Context, s Settings, system string, conv []turnMsg, tools []toolSpec, onText func(string)) (turnOut, error) {
	msgs := []map[string]any{{"role": "system", "content": system}}
	for _, m := range conv {
		for _, r := range m.Results {
			msgs = append(msgs, map[string]any{"role": "tool", "tool_call_id": r.ID, "content": r.Content})
		}
		if m.Role == "assistant" {
			am := map[string]any{"role": "assistant", "content": m.Text}
			if len(m.Calls) > 0 {
				var calls []map[string]any
				for _, c := range m.Calls {
					calls = append(calls, map[string]any{"id": c.ID, "type": "function", "function": map[string]any{"name": c.Name, "arguments": string(c.Input)}})
				}
				am["tool_calls"] = calls
			}
			msgs = append(msgs, am)
		} else if m.Text != "" {
			msgs = append(msgs, map[string]any{"role": "user", "content": m.Text})
		}
	}
	reqBody := map[string]any{
		"model":          s.Model,
		"messages":       msgs,
		"max_tokens":     s.MaxTokens,
		"stream":         true,
		"stream_options": map[string]any{"include_usage": true},
	}
	if len(tools) > 0 {
		var ts []map[string]any
		for _, t := range tools {
			ts = append(ts, map[string]any{"type": "function", "function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  map[string]any{"type": "object", "properties": t.Properties, "required": t.Required},
			}})
		}
		reqBody["tools"] = ts
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return turnOut{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if s.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.APIKey)
	}
	if s.Provider == "openrouter" {
		req.Header.Set("X-Title", s.BrandName)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return turnOut{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return turnOut{}, fmt.Errorf("provider returned %d: %s", resp.StatusCode, providerError(snippet))
	}

	var out turnOut
	var text strings.Builder
	// tool calls arrive in pieces keyed by index
	calls := map[int]*toolCall{}
	args := map[int]*strings.Builder{}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "[DONE]" {
			break
		}
		var ev struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int64 `json:"prompt_tokens"`
				CompletionTokens int64 `json:"completion_tokens"`
			} `json:"usage"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		if ev.Error != nil {
			return out, errors.New(ev.Error.Message)
		}
		if ev.Usage != nil {
			out.InputTokens, out.OutputTokens = ev.Usage.PromptTokens, ev.Usage.CompletionTokens
		}
		for _, c := range ev.Choices {
			if c.Delta.Content != "" {
				text.WriteString(c.Delta.Content)
				onText(c.Delta.Content)
			}
			for _, tc := range c.Delta.ToolCalls {
				call, ok := calls[tc.Index]
				if !ok {
					call = &toolCall{}
					calls[tc.Index] = call
					args[tc.Index] = &strings.Builder{}
				}
				if tc.ID != "" {
					call.ID = tc.ID
				}
				if tc.Function.Name != "" {
					call.Name = tc.Function.Name
				}
				args[tc.Index].WriteString(tc.Function.Arguments)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return out, err
	}
	out.Text = text.String()

	idx := make([]int, 0, len(calls))
	for i := range calls {
		idx = append(idx, i)
	}
	sort.Ints(idx)
	for _, i := range idx {
		c := calls[i]
		c.Input = json.RawMessage(args[i].String())
		if !json.Valid(c.Input) {
			c.Input = json.RawMessage("{}")
		}
		if c.ID == "" {
			c.ID = fmt.Sprintf("call_%d", i)
		}
		out.Calls = append(out.Calls, *c)
	}
	return out, nil
}

// providerError pulls the message out of an openai-style error body so the admin sees why in the logs
func providerError(body []byte) string {
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error.Message != "" {
		return e.Error.Message
	}
	return strings.TrimSpace(string(body))
}
