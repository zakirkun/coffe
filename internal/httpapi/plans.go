package httpapi

import (
	"coffe/internal/service"
	"net/http"
)

type previewPlanRequest struct {
	ProjectID string `json:"project_id"`
	TargetID  string `json:"target_id"`
	Prompt    string `json:"prompt"`
}

func (s *Server) previewPlan(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[previewPlanRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	plan, err := s.app.PreviewDeploy(request.Context(), service.PreviewInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	preview, err := s.app.PlanPreview(request.Context(), string(plan.ID()))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{
		status: http.StatusCreated, body: toPlanDTO(preview.Plan, preview.Commands),
	})
}

func (s *Server) approvePlan(writer http.ResponseWriter, request *http.Request) {
	plan, err := s.app.ApprovePlan(request.Context(), request.PathValue("planID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: toPlanDTO(plan, nil)})
}

func (s *Server) executePlan(writer http.ResponseWriter, request *http.Request) {
	plan, err := s.app.StartPlan(request.Context(), request.PathValue("planID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusAccepted, body: toPlanDTO(plan, nil)})
}

func (s *Server) cancelPlan(writer http.ResponseWriter, request *http.Request) {
	if err := s.app.CancelPlan(request.Context(), request.PathValue("planID")); err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusAccepted, body: map[string]string{
		"status": "cancellation_requested",
	}})
}

func (s *Server) retryPlan(writer http.ResponseWriter, request *http.Request) {
	plan, err := s.app.RetryPlan(request.Context(), request.PathValue("planID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusCreated, body: toPlanDTO(plan, nil)})
}

func (s *Server) previewRollback(writer http.ResponseWriter, request *http.Request) {
	plan, err := s.app.PreviewRollback(request.Context(), request.PathValue("targetID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	preview, err := s.app.PlanPreview(request.Context(), string(plan.ID()))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{
		status: http.StatusCreated, body: toPlanDTO(preview.Plan, preview.Commands),
	})
}

func (s *Server) getPlan(writer http.ResponseWriter, request *http.Request) {
	preview, err := s.app.PlanPreview(request.Context(), request.PathValue("planID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{
		status: http.StatusOK, body: toPlanDTO(preview.Plan, preview.Commands),
	})
}

func (s *Server) listPlans(writer http.ResponseWriter, request *http.Request) {
	plans, err := s.app.ListPlans(request.Context(), 50)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]planDTO, 0, len(plans))
	for _, plan := range plans {
		result = append(result, toPlanDTO(plan, nil))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) listEvents(writer http.ResponseWriter, request *http.Request) {
	events, err := s.app.ListEvents(request.Context(), request.PathValue("planID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]eventDTO, 0, len(events))
	for _, event := range events {
		result = append(result, toEventDTO(event))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}
