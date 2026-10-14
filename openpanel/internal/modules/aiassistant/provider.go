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
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var errRefused = errors.New("the model declined to answer this")

// streamChat sends the conversation to the configured provider and calls onText with each piece of the reply as it arrives
func streamChat(ctx context.Context, s Settings, system string, msgs []message, onText func(string)) error {
	if s.Provider == "anthropic" {
		return streamAnthropic(ctx, s, system, msgs, onText)
	}
	return streamOpenAI(ctx, s, system, msgs, onText)
}

func streamAnthropic(ctx context.Context, s Settings, system string, msgs []message, onText func(string)) error {
	opts := []option.RequestOption{option.WithAPIKey(s.APIKey)}
	if s.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(s.BaseURL))
	}
	client := anthropic.NewClient(opts...)

	params := anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(s.Model),
		MaxTokens: int64(s.MaxTokens),
		System:    []anthropic.BetaTextBlockParam{{Text: system}},
		// a policy decline gets re-served by the server's default fallback model, models without fallbacks just run alone
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		Fallbacks: anthropic.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()},
	}
	if supportsEffort(s.Model) {
		params.OutputConfig = anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortLow}
	}
	for _, m := range msgs {
		block := anthropic.NewBetaTextBlock(m.Content)
		if m.Role == "assistant" {
			params.Messages = append(params.Messages, anthropic.BetaMessageParam{Role: anthropic.BetaMessageParamRoleAssistant, Content: []anthropic.BetaContentBlockParamUnion{block}})
		} else {
			params.Messages = append(params.Messages, anthropic.NewBetaUserMessage(block))
		}
	}

	stream := client.Beta.Messages.NewStreaming(ctx, params)
	defer stream.Close()
	refused := false
	for stream.Next() {
		switch ev := stream.Current().AsAny().(type) {
		case anthropic.BetaRawContentBlockDeltaEvent:
			if d, ok := ev.Delta.AsAny().(anthropic.BetaTextDelta); ok {
				onText(d.Text)
			}
		case anthropic.BetaRawMessageDeltaEvent:
			if ev.Delta.StopReason == anthropic.BetaStopReasonRefusal {
				refused = true
			}
		}
	}
	if err := stream.Err(); err != nil {
		var apiErr *anthropic.Error
		if errors.As(err, &apiErr) {
			return fmt.Errorf("provider returned %d", apiErr.StatusCode)
		}
		return err
	}
	if refused {
		return errRefused
	}
	return nil
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

// streamOpenAI talks to openrouter, openai or anything else that speaks /chat/completions
func streamOpenAI(ctx context.Context, s Settings, system string, msgs []message, onText func(string)) error {
	all := append([]message{{Role: "system", Content: system}}, msgs...)
	body, _ := json.Marshal(map[string]any{
		"model":      s.Model,
		"messages":   all,
		"max_tokens": s.MaxTokens,
		"stream":     true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("provider returned %d: %s", resp.StatusCode, providerError(snippet))
	}

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
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		if ev.Error != nil {
			return errors.New(ev.Error.Message)
		}
		for _, c := range ev.Choices {
			if c.Delta.Content != "" {
				onText(c.Delta.Content)
			}
		}
	}
	return sc.Err()
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
