package httpapi

import (
	"coffe/internal/domain"
	"time"
)

type projectDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RootPath  string    `json:"root_path"`
	CreatedAt time.Time `json:"created_at"`
}

type targetDTO struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Name       string    `json:"name"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	User       string    `json:"user"`
	RemotePath string    `json:"remote_path"`
	Branch     string    `json:"branch"`
	CreatedAt  time.Time `json:"created_at"`
}

type planDTO struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"project_id"`
	TargetID   string            `json:"target_id"`
	SourceID   string            `json:"source_plan_id,omitempty"`
	Kind       domain.PlanKind   `json:"kind"`
	Status     domain.PlanStatus `json:"status"`
	Revision   string            `json:"revision,omitempty"`
	Steps      []domain.StepKind `json:"steps"`
	Commands   []string          `json:"commands,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	ApprovedAt *time.Time        `json:"approved_at,omitempty"`
}

type eventDTO struct {
	ID         string           `json:"id"`
	PlanID     string           `json:"plan_id"`
	Kind       domain.EventKind `json:"kind"`
	Message    string           `json:"message"`
	OccurredAt time.Time        `json:"occurred_at"`
}

func toProjectDTO(project domain.Project) projectDTO {
	return projectDTO{
		ID: string(project.ID()), Name: project.Name().String(),
		RootPath: project.RootPath(), CreatedAt: project.CreatedAt(),
	}
}

func toTargetDTO(target domain.Target) targetDTO {
	return targetDTO{
		ID: string(target.ID()), ProjectID: string(target.ProjectID()), Name: target.Name().String(),
		Host: target.Host().String(), Port: target.Port().Int(), User: target.User().String(),
		RemotePath: target.RemotePath().String(), Branch: target.Branch().String(), CreatedAt: target.CreatedAt(),
	}
}

func toPlanDTO(plan domain.ActionPlan, commands []string) planDTO {
	result := planDTO{
		ID: string(plan.ID()), ProjectID: string(plan.ProjectID()), TargetID: string(plan.TargetID()),
		SourceID: string(plan.SourcePlanID()), Kind: plan.Kind(), Status: plan.Status(),
		Steps: plan.Steps(), Commands: commands, CreatedAt: plan.CreatedAt(),
	}
	if revision, ok := plan.Revision(); ok {
		result.Revision = revision.String()
	}
	if approvedAt, ok := plan.ApprovedAt(); ok {
		result.ApprovedAt = &approvedAt
	}
	return result
}

func toEventDTO(event domain.Event) eventDTO {
	return eventDTO{
		ID: string(event.ID()), PlanID: string(event.PlanID()), Kind: event.Kind(),
		Message: event.Message(), OccurredAt: event.OccurredAt(),
	}
}
