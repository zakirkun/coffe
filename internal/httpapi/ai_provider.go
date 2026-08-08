package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"net/http"
	"time"
)

type createAIProviderRequest struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	BaseURL    string `json:"base_url"`
	Model      string `json:"model"`
	SecretName string `json:"secret_name"`
	Priority   int    `json:"priority"`
	Enabled    bool   `json:"enabled"`
}

type aiProviderDTO struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Kind      domain.AIProviderKind `json:"kind"`
	BaseURL   string                `json:"base_url"`
	Model     string                `json:"model"`
	Priority  int                   `json:"priority"`
	Enabled   bool                  `json:"enabled"`
	HasSecret bool                  `json:"has_secret"`
	CreatedAt time.Time             `json:"created_at"`
}

type aiUsageDTO struct {
	ID               string    `json:"id"`
	ProviderID       string    `json:"provider_id"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	LatencyMillis    int64     `json:"latency_millis"`
	Success          bool      `json:"success"`
	CreatedAt        time.Time `json:"created_at"`
}

func (s *Server) createAIProvider(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[createAIProviderRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	provider, err := s.app.CreateAIProvider(request.Context(), service.CreateAIProviderInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusCreated, body: toAIProviderDTO(provider)})
}

func (s *Server) listAIProviders(writer http.ResponseWriter, request *http.Request) {
	providers, err := s.app.ListAIProviders(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]aiProviderDTO, 0, len(providers))
	for _, provider := range providers {
		result = append(result, toAIProviderDTO(provider))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) listAIModels(writer http.ResponseWriter, request *http.Request) {
	models, err := s.app.ListAIModels(request.Context(), request.PathValue("providerID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: models})
}

func (s *Server) listAIUsage(writer http.ResponseWriter, request *http.Request) {
	usage, err := s.app.ListAIUsage(request.Context(), 100)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]aiUsageDTO, 0, len(usage))
	for _, record := range usage {
		result = append(result, aiUsageDTO{
			ID: string(record.ID()), ProviderID: string(record.ProviderID()), Model: record.Model(),
			PromptTokens: record.PromptTokens(), CompletionTokens: record.CompletionTokens(),
			TotalTokens: record.TotalTokens(), LatencyMillis: record.LatencyMillis(),
			Success: record.Success(), CreatedAt: record.CreatedAt(),
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func toAIProviderDTO(provider domain.AIProvider) aiProviderDTO {
	return aiProviderDTO{
		ID: string(provider.ID()), Name: provider.Name().String(), Kind: provider.Kind(),
		BaseURL: provider.BaseURL(), Model: provider.Model(), Priority: provider.Priority(),
		Enabled: provider.Enabled(), HasSecret: provider.SecretName() != "", CreatedAt: provider.CreatedAt(),
	}
}
