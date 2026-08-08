package ai

import (
	"bytes"
	"coffe/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidResponse = errors.New("ai: invalid response")

const maxResponseBytes = 4 * 1024 * 1024

type OpenAIConfig struct {
	BaseURL string
	Model   string
	APIKey  string
	Timeout time.Duration
}

type OpenAIPlanner struct {
	baseURL  string
	endpoint string
	model    string
	apiKey   string
	client   *http.Client
}

func NewOpenAIPlanner(config OpenAIConfig) (*OpenAIPlanner, error) {
	baseURL, err := url.Parse(strings.TrimRight(config.BaseURL, "/"))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("openai base URL %q: invalid", config.BaseURL)
	}
	if strings.TrimSpace(config.Model) == "" || config.Timeout <= 0 {
		return nil, fmt.Errorf("openai model or timeout: invalid")
	}
	return &OpenAIPlanner{
		baseURL:  baseURL.String(),
		endpoint: baseURL.String() + "/chat/completions",
		model:    config.Model, apiKey: config.APIKey,
		client: &http.Client{Timeout: config.Timeout},
	}, nil
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature int           `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Usage CompletionUsage `json:"usage"`
}

type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (p *OpenAIPlanner) Classify(ctx context.Context, prompt string) (domain.Intent, error) {
	intent, _, err := p.ClassifyDetailed(ctx, prompt)
	return intent, err
}

func (p *OpenAIPlanner) ClassifyDetailed(
	ctx context.Context,
	prompt string,
) (intent domain.Intent, usage CompletionUsage, err error) {
	payload, err := json.Marshal(chatRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: "Return only JSON with action deploy. Reject every other operation."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0,
	})
	if err != nil {
		return "", CompletionUsage{}, fmt.Errorf("encode chat request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", CompletionUsage{}, fmt.Errorf("create chat request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	response, err := p.client.Do(request)
	if err != nil {
		return "", CompletionUsage{}, fmt.Errorf("send chat request: %w", err)
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", CompletionUsage{}, fmt.Errorf("chat response status %d", response.StatusCode)
	}
	var chat chatResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes))
	if err := decoder.Decode(&chat); err != nil {
		return "", CompletionUsage{}, fmt.Errorf("decode chat response: %w", err)
	}
	if len(chat.Choices) != 1 {
		return "", CompletionUsage{}, fmt.Errorf("choice count %d: %w", len(chat.Choices), ErrInvalidResponse)
	}
	var classified struct {
		Action domain.Intent `json:"action"`
	}
	contentDecoder := json.NewDecoder(strings.NewReader(chat.Choices[0].Message.Content))
	contentDecoder.DisallowUnknownFields()
	if err := contentDecoder.Decode(&classified); err != nil {
		return "", CompletionUsage{}, fmt.Errorf("decode classified intent: %w", err)
	}
	if !classified.Action.Valid() {
		return "", CompletionUsage{}, fmt.Errorf("action %q: %w", classified.Action, ErrInvalidResponse)
	}
	return classified.Action, chat.Usage, nil
}
