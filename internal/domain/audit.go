package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidAudit = errors.New("domain: invalid audit entry")

type (
	AuditID       string
	AuditCategory string
	AuditOutcome  string
)

const (
	AuditDeploy   AuditCategory = "deploy"
	AuditGit      AuditCategory = "git"
	AuditAgent    AuditCategory = "agent"
	AuditDatabase AuditCategory = "database"
	AuditAI       AuditCategory = "ai"
	AuditPolicy   AuditCategory = "policy"
	AuditBackup   AuditCategory = "backup"
	AuditUpdate   AuditCategory = "update"
	AuditWorker   AuditCategory = "worker"
	AuditPlugin   AuditCategory = "plugin"
)

const (
	AuditSuccess AuditOutcome = "success"
	AuditFailure AuditOutcome = "failure"
	AuditDenied  AuditOutcome = "denied"
)

func (c AuditCategory) Valid() bool {
	switch c {
	case AuditDeploy, AuditGit, AuditAgent, AuditDatabase, AuditAI, AuditPolicy, AuditBackup, AuditUpdate, AuditWorker, AuditPlugin:
		return true
	default:
		return false
	}
}

func (o AuditOutcome) Valid() bool {
	return o == AuditSuccess || o == AuditFailure || o == AuditDenied
}

type AuditEntry struct {
	occurredAt time.Time
	id         AuditID
	category   AuditCategory
	action     string
	resourceID string
	outcome    AuditOutcome
	message    string
}

type AuditValues struct {
	ID         AuditID
	Category   AuditCategory
	Action     string
	ResourceID string
	Outcome    AuditOutcome
	Message    string
	OccurredAt time.Time
}

func NewAuditEntry(values AuditValues) (AuditEntry, error) {
	action := strings.TrimSpace(values.Action)
	resourceID := strings.TrimSpace(values.ResourceID)
	message := strings.TrimSpace(values.Message)
	if values.ID == "" || !values.Category.Valid() || action == "" || resourceID == "" ||
		!values.Outcome.Valid() || message == "" || len(message) > 500 || values.OccurredAt.IsZero() ||
		hasControl(action) || hasControl(resourceID) || hasControl(message) {
		return AuditEntry{}, fmt.Errorf("audit entry: %w", ErrInvalidAudit)
	}
	return AuditEntry{
		id: values.ID, category: values.Category, action: action,
		resourceID: resourceID, outcome: values.Outcome, message: message,
		occurredAt: values.OccurredAt.UTC(),
	}, nil
}

func (e AuditEntry) ID() AuditID             { return e.id }
func (e AuditEntry) Category() AuditCategory { return e.category }
func (e AuditEntry) Action() string          { return e.action }
func (e AuditEntry) ResourceID() string      { return e.resourceID }
func (e AuditEntry) Outcome() AuditOutcome   { return e.outcome }
func (e AuditEntry) Message() string         { return e.message }
func (e AuditEntry) OccurredAt() time.Time   { return e.occurredAt }
