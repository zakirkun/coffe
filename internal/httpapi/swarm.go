package httpapi

import "net/http"

type swarmNodeDTO struct {
	ID           string `json:"id"`
	Hostname     string `json:"hostname"`
	Status       string `json:"status"`
	Availability string `json:"availability"`
	Manager      bool   `json:"manager"`
}

type swarmServiceDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	Replicas string `json:"replicas"`
	Image    string `json:"image"`
}

func (s *Server) listSwarmNodes(writer http.ResponseWriter, request *http.Request) {
	nodes, err := s.app.ListSwarmNodes(request.Context(), request.PathValue("targetID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]swarmNodeDTO, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, swarmNodeDTO{
			ID: node.ID(), Hostname: node.Hostname(), Status: node.Status(),
			Availability: node.Availability(), Manager: node.Manager(),
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) listSwarmServices(writer http.ResponseWriter, request *http.Request) {
	services, err := s.app.ListSwarmServices(request.Context(), request.PathValue("targetID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]swarmServiceDTO, 0, len(services))
	for _, service := range services {
		result = append(result, swarmServiceDTO{
			ID: service.ID(), Name: service.Name(), Mode: service.Mode(),
			Replicas: service.Replicas(), Image: service.Image(),
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}
