package worker

import "coffe/internal/domain"

type TaskRequest struct {
	PlanID      string                `json:"plan_id"`
	ProjectID   string                `json:"project_id"`
	TargetID    string                `json:"target_id"`
	PlanKind    domain.PlanKind       `json:"plan_kind"`
	Step        domain.StepKind       `json:"step"`
	Root        string                `json:"root"`
	Branch      string                `json:"branch"`
	ComposeFile string                `json:"compose_file"`
	Mode        domain.DeploymentMode `json:"mode"`
	StackName   string                `json:"stack_name,omitempty"`
	Revision    string                `json:"revision,omitempty"`
	Approved    bool                  `json:"approved"`
}

type TaskResponse struct {
	Logs     []string `json:"logs"`
	Revision string   `json:"revision,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
