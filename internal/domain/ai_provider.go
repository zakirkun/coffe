package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidAIProvider = errors.New("domain: invalid AI provider")

type (
	AIProviderID   string
	AIProviderKind string
)

const (
	AIProviderOpenAI   AIProviderKind = "openai"
	AIProviderOllama   AIProviderKind = "ollama"
	AIProviderVLLM     AIProviderKind = "vllm"
	AIProviderLMStudio AIProviderKind = "lmstudio"
)

func (k AIProviderKind) Valid() bool {
	switch k {
	case AIProviderOpenAI, AIProviderOllama, AIProviderVLLM, AIProviderLMStudio:
		return true
	default:
		return false
	}
}

type AIProvider struct {
	createdAt  time.Time
	id         AIProviderID
	name       Name
	kind       AIProviderKind
	baseURL    string
	model      string
	secretName string
	priority   int
	enabled    bool
}

type AIProviderValues struct {
	ID         AIProviderID
	Name       Name
	Kind       AIProviderKind
	BaseURL    string
	Model      string
	SecretName string
	Priority   int
	Enabled    bool
	CreatedAt  time.Time
}

func NewAIProvider(values AIProviderValues) (AIProvider, error) {
	baseURL, err := parseProviderURL(values.BaseURL)
	if values.ID == "" || values.Name.String() == "" || !values.Kind.Valid() || err != nil ||
		strings.TrimSpace(values.Model) == "" || values.Priority < 0 || values.CreatedAt.IsZero() ||
		hasControl(values.Model) || hasControl(values.SecretName) {
		return AIProvider{}, fmt.Errorf("AI provider: %w", ErrInvalidAIProvider)
	}
	return AIProvider{
		id: values.ID, name: values.Name, kind: values.Kind, baseURL: baseURL,
		model: strings.TrimSpace(values.Model), secretName: strings.TrimSpace(values.SecretName),
		priority: values.Priority, enabled: values.Enabled, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func parseProviderURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", ErrInvalidAIProvider
	}
	return parsed.String(), nil
}

func (p AIProvider) ID() AIProviderID     { return p.id }
func (p AIProvider) Name() Name           { return p.name }
func (p AIProvider) Kind() AIProviderKind { return p.kind }
func (p AIProvider) BaseURL() string      { return p.baseURL }
func (p AIProvider) Model() string        { return p.model }
func (p AIProvider) SecretName() string   { return p.secretName }
func (p AIProvider) Priority() int        { return p.priority }
func (p AIProvider) Enabled() bool        { return p.enabled }
func (p AIProvider) CreatedAt() time.Time { return p.createdAt }

type AIUsageID string

type AIUsage struct {
	createdAt        time.Time
	id               AIUsageID
	providerID       AIProviderID
	model            string
	promptTokens     int
	completionTokens int
	totalTokens      int
	latencyMillis    int64
	success          bool
}

type AIUsageValues struct {
	ID               AIUsageID
	ProviderID       AIProviderID
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	LatencyMillis    int64
	Success          bool
	CreatedAt        time.Time
}

func NewAIUsage(values AIUsageValues) (AIUsage, error) {
	if values.ID == "" || values.ProviderID == "" || strings.TrimSpace(values.Model) == "" ||
		values.PromptTokens < 0 || values.CompletionTokens < 0 || values.TotalTokens < 0 ||
		values.LatencyMillis < 0 || values.CreatedAt.IsZero() {
		return AIUsage{}, fmt.Errorf("AI usage: %w", ErrInvalidAIProvider)
	}
	return AIUsage{
		id: values.ID, providerID: values.ProviderID, model: values.Model,
		promptTokens: values.PromptTokens, completionTokens: values.CompletionTokens,
		totalTokens: values.TotalTokens, latencyMillis: values.LatencyMillis,
		success: values.Success, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func (u AIUsage) ID() AIUsageID            { return u.id }
func (u AIUsage) ProviderID() AIProviderID { return u.providerID }
func (u AIUsage) Model() string            { return u.model }
func (u AIUsage) PromptTokens() int        { return u.promptTokens }
func (u AIUsage) CompletionTokens() int    { return u.completionTokens }
func (u AIUsage) TotalTokens() int         { return u.totalTokens }
func (u AIUsage) LatencyMillis() int64     { return u.latencyMillis }
func (u AIUsage) Success() bool            { return u.success }
func (u AIUsage) CreatedAt() time.Time     { return u.createdAt }
