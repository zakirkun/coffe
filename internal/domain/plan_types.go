package domain

type PlanStatus string

const (
	PlanDraft     PlanStatus = "draft"
	PlanApproved  PlanStatus = "approved"
	PlanRunning   PlanStatus = "running"
	PlanSucceeded PlanStatus = "succeeded"
	PlanFailed    PlanStatus = "failed"
	PlanCanceled  PlanStatus = "canceled"
)

type PlanKind string

const (
	PlanComposeDeploy   PlanKind = "compose_deploy"
	PlanComposeRollback PlanKind = "compose_rollback"
	PlanSwarmDeploy     PlanKind = "swarm_deploy"
)

func (k PlanKind) Valid() bool {
	switch k {
	case PlanComposeDeploy, PlanComposeRollback, PlanSwarmDeploy:
		return true
	default:
		return false
	}
}

type StepKind string

const (
	StepInspectRepository StepKind = "inspect_repository"
	StepGitPull           StepKind = "git_pull"
	StepCaptureRevision   StepKind = "capture_revision"
	StepComposeConfig     StepKind = "compose_config"
	StepComposeBuild      StepKind = "compose_build"
	StepComposeUp         StepKind = "compose_up"
	StepComposeHealth     StepKind = "compose_health"
	StepGitReset          StepKind = "git_reset"
	StepSwarmConfig       StepKind = "swarm_config"
	StepSwarmDeploy       StepKind = "swarm_deploy"
	StepSwarmHealth       StepKind = "swarm_health"
)
