package httpapi

import (
	"coffe/internal/service"
	"net/http"
)

type createProjectRequest struct {
	Name     string `json:"name"`
	RootPath string `json:"root_path"`
}

func (s *Server) createProject(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[createProjectRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	project, err := s.app.CreateProject(request.Context(), service.CreateProjectInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusCreated, body: toProjectDTO(project)})
}

func (s *Server) listProjects(writer http.ResponseWriter, request *http.Request) {
	projects, err := s.app.ListProjects(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]projectDTO, 0, len(projects))
	for _, project := range projects {
		result = append(result, toProjectDTO(project))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

type createTargetRequest struct {
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	RemotePath string `json:"remote_path"`
	Branch     string `json:"branch"`
}

func (s *Server) createTarget(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[createTargetRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	target, err := s.app.CreateTarget(request.Context(), service.CreateTargetInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusCreated, body: toTargetDTO(target)})
}

func (s *Server) listTargets(writer http.ResponseWriter, request *http.Request) {
	targets, err := s.app.ListTargets(request.Context(), request.PathValue("projectID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]targetDTO, 0, len(targets))
	for _, target := range targets {
		result = append(result, toTargetDTO(target))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}
