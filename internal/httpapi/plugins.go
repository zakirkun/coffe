package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"encoding/json"
	"net/http"
	"time"
)

type registerPluginRequest struct {
	Name           string   `json:"name"`
	ExecutablePath string   `json:"executable_path"`
	SHA256         string   `json:"sha256"`
	Methods        []string `json:"methods"`
	Enabled        bool     `json:"enabled"`
}

type invokePluginRequest struct {
	Method string          `json:"method"`
	Input  json.RawMessage `json:"input"`
}

type pluginDTO struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ExecutablePath string    `json:"executable_path"`
	SHA256         string    `json:"sha256"`
	Methods        []string  `json:"methods"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

func (s *Server) registerPlugin(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[registerPluginRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	plugin, err := s.app.RegisterPlugin(request.Context(), service.RegisterPluginInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusCreated, body: toPluginDTO(plugin)})
}

func (s *Server) listPlugins(writer http.ResponseWriter, request *http.Request) {
	plugins, err := s.app.ListPlugins(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]pluginDTO, 0, len(plugins))
	for _, plugin := range plugins {
		result = append(result, toPluginDTO(plugin))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) invokePlugin(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[invokePluginRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result, err := s.app.InvokePlugin(request.Context(), service.PluginInvocationInput{
		PluginID: request.PathValue("pluginID"), Method: input.Method, Input: input.Input,
	})
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func toPluginDTO(plugin domain.Plugin) pluginDTO {
	return pluginDTO{
		ID: string(plugin.ID()), Name: plugin.Name().String(), ExecutablePath: plugin.ExecutablePath(),
		SHA256: plugin.SHA256(), Methods: plugin.Methods(), Enabled: plugin.Enabled(),
		CreatedAt: plugin.CreatedAt(),
	}
}
