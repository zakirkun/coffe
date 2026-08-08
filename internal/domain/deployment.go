package domain

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidDeployment = errors.New("domain: invalid deployment record")

type HealthStatus string

const (
	HealthHealthy HealthStatus = "healthy"
	HealthUnknown HealthStatus = "unknown"
)

func (s HealthStatus) Valid() bool { return s == HealthHealthy || s == HealthUnknown }

type DeploymentRecord struct {
	createdAt      time.Time
	planID         PlanID
	targetID       TargetID
	kind           PlanKind
	revisionBefore GitRevision
	revisionAfter  GitRevision
	health         HealthStatus
}

type DeploymentValues struct {
	PlanID         PlanID
	TargetID       TargetID
	Kind           PlanKind
	RevisionBefore GitRevision
	RevisionAfter  GitRevision
	Health         HealthStatus
	CreatedAt      time.Time
}

func NewDeploymentRecord(values DeploymentValues) (DeploymentRecord, error) {
	if values.PlanID == "" || values.TargetID == "" || !values.Kind.Valid() ||
		values.RevisionBefore.IsZero() || values.RevisionAfter.IsZero() ||
		!values.Health.Valid() || values.CreatedAt.IsZero() {
		return DeploymentRecord{}, fmt.Errorf("deployment record: %w", ErrInvalidDeployment)
	}
	return DeploymentRecord{
		planID: values.PlanID, targetID: values.TargetID, kind: values.Kind,
		revisionBefore: values.RevisionBefore, revisionAfter: values.RevisionAfter,
		health: values.Health, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func (r DeploymentRecord) PlanID() PlanID              { return r.planID }
func (r DeploymentRecord) TargetID() TargetID          { return r.targetID }
func (r DeploymentRecord) Kind() PlanKind              { return r.kind }
func (r DeploymentRecord) RevisionBefore() GitRevision { return r.revisionBefore }
func (r DeploymentRecord) RevisionAfter() GitRevision  { return r.revisionAfter }
func (r DeploymentRecord) Health() HealthStatus        { return r.health }
func (r DeploymentRecord) CreatedAt() time.Time        { return r.createdAt }
