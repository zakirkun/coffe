package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

var (
	ErrInvalidPlan    = errors.New("domain: invalid action plan")
	ErrPlanTransition = errors.New("domain: invalid plan transition")
)

type PlanID string

type ActionPlan struct {
	createdAt  time.Time
	approvedAt time.Time
	id         PlanID
	projectID  ProjectID
	targetID   TargetID
	sourceID   PlanID
	kind       PlanKind
	status     PlanStatus
	revision   GitRevision
	steps      []StepKind
}

type PlanValues struct {
	ID         PlanID
	ProjectID  ProjectID
	TargetID   TargetID
	SourceID   PlanID
	Kind       PlanKind
	Status     PlanStatus
	Revision   GitRevision
	Steps      []StepKind
	CreatedAt  time.Time
	ApprovedAt time.Time
}

type PlanDraftValues struct {
	ID        PlanID
	ProjectID ProjectID
	TargetID  TargetID
	SourceID  PlanID
	Kind      PlanKind
	Revision  GitRevision
	Steps     []StepKind
	CreatedAt time.Time
}

func NewActionPlan(values PlanDraftValues) (ActionPlan, error) {
	return RestoreActionPlan(PlanValues{
		ID: values.ID, ProjectID: values.ProjectID, TargetID: values.TargetID,
		SourceID: values.SourceID, Kind: values.Kind, Status: PlanDraft, Revision: values.Revision,
		Steps: values.Steps, CreatedAt: values.CreatedAt,
	})
}

func RestoreActionPlan(values PlanValues) (ActionPlan, error) {
	if err := validatePlanValues(values); err != nil {
		return ActionPlan{}, err
	}
	return ActionPlan{
		id: values.ID, projectID: values.ProjectID, targetID: values.TargetID,
		sourceID: values.SourceID, kind: values.Kind, status: values.Status,
		revision: values.Revision, steps: slices.Clone(values.Steps),
		createdAt: values.CreatedAt.UTC(), approvedAt: values.ApprovedAt.UTC(),
	}, nil
}

func validatePlanValues(values PlanValues) error {
	if values.ID == "" || values.ProjectID == "" || values.TargetID == "" || values.CreatedAt.IsZero() {
		return fmt.Errorf("plan identity: %w", ErrInvalidPlan)
	}
	if !values.Kind.Valid() || !validPlanStatus(values.Status) || len(values.Steps) == 0 {
		return fmt.Errorf("plan state: %w", ErrInvalidPlan)
	}
	if err := validatePlanRevision(values.Kind, values.Revision); err != nil {
		return err
	}
	for _, step := range values.Steps {
		if !step.Valid() {
			return fmt.Errorf("step %q: %w", step, ErrInvalidPlan)
		}
	}
	if values.Status != PlanDraft && values.ApprovedAt.IsZero() {
		return fmt.Errorf("approved timestamp: %w", ErrInvalidPlan)
	}
	return nil
}

func validatePlanRevision(kind PlanKind, revision GitRevision) error {
	if kind == PlanComposeRollback && revision.IsZero() {
		return fmt.Errorf("rollback revision: %w", ErrInvalidPlan)
	}
	if kind != PlanComposeRollback && !revision.IsZero() {
		return fmt.Errorf("unexpected revision: %w", ErrInvalidPlan)
	}
	return nil
}

func (p ActionPlan) Approve(at time.Time) (ActionPlan, error) {
	if p.status != PlanDraft || at.IsZero() {
		return ActionPlan{}, fmt.Errorf("approve %s: %w", p.status, ErrPlanTransition)
	}
	p.status = PlanApproved
	p.approvedAt = at.UTC()
	return p, nil
}

func (p ActionPlan) Start() (ActionPlan, error) {
	if p.status != PlanApproved {
		return ActionPlan{}, fmt.Errorf("start %s: %w", p.status, ErrPlanTransition)
	}
	p.status = PlanRunning
	return p, nil
}

func (p ActionPlan) Finish(succeeded bool) (ActionPlan, error) {
	if p.status != PlanRunning {
		return ActionPlan{}, fmt.Errorf("finish %s: %w", p.status, ErrPlanTransition)
	}
	if succeeded {
		p.status = PlanSucceeded
	} else {
		p.status = PlanFailed
	}
	return p, nil
}

func (p ActionPlan) Cancel() (ActionPlan, error) {
	if p.status != PlanRunning {
		return ActionPlan{}, fmt.Errorf("cancel %s: %w", p.status, ErrPlanTransition)
	}
	p.status = PlanCanceled
	return p, nil
}

func (p ActionPlan) Retry(id PlanID, createdAt time.Time) (ActionPlan, error) {
	if p.status != PlanFailed && p.status != PlanCanceled {
		return ActionPlan{}, fmt.Errorf("retry %s: %w", p.status, ErrPlanTransition)
	}
	return NewActionPlan(PlanDraftValues{
		ID: id, ProjectID: p.projectID, TargetID: p.targetID, SourceID: p.id,
		Kind: p.kind, Revision: p.revision, Steps: p.steps, CreatedAt: createdAt,
	})
}

func (p ActionPlan) ID() PlanID           { return p.id }
func (p ActionPlan) ProjectID() ProjectID { return p.projectID }
func (p ActionPlan) TargetID() TargetID   { return p.targetID }
func (p ActionPlan) SourcePlanID() PlanID { return p.sourceID }
func (p ActionPlan) Kind() PlanKind       { return p.kind }
func (p ActionPlan) Status() PlanStatus   { return p.status }
func (p ActionPlan) Steps() []StepKind    { return slices.Clone(p.steps) }
func (p ActionPlan) CreatedAt() time.Time { return p.createdAt }

func (p ActionPlan) ApprovedAt() (time.Time, bool) {
	return p.approvedAt, !p.approvedAt.IsZero()
}

func (p ActionPlan) Revision() (GitRevision, bool) {
	return p.revision, !p.revision.IsZero()
}

func (s StepKind) Valid() bool {
	switch s {
	case StepInspectRepository, StepGitPull, StepCaptureRevision,
		StepComposeConfig, StepComposeBuild, StepComposeUp, StepComposeHealth,
		StepGitReset, StepSwarmConfig, StepSwarmDeploy, StepSwarmHealth:
		return true
	default:
		return false
	}
}

func validPlanStatus(status PlanStatus) bool {
	switch status {
	case PlanDraft, PlanApproved, PlanRunning, PlanSucceeded, PlanFailed, PlanCanceled:
		return true
	default:
		return false
	}
}
