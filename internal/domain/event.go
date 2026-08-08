package domain

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidEvent = errors.New("domain: invalid event")

type EventID string

type EventKind string

const (
	EventPlanCreated   EventKind = "plan_created"
	EventPlanApproved  EventKind = "plan_approved"
	EventPlanStarted   EventKind = "plan_started"
	EventStepStarted   EventKind = "step_started"
	EventStepOutput    EventKind = "step_output"
	EventStepSucceeded EventKind = "step_succeeded"
	EventPlanSucceeded EventKind = "plan_succeeded"
	EventPlanFailed    EventKind = "plan_failed"
	EventPlanCanceled  EventKind = "plan_canceled"
	EventPlanRetried   EventKind = "plan_retried"
	EventPlanRecovered EventKind = "plan_recovered"
)

type Event struct {
	occurredAt time.Time
	id         EventID
	planID     PlanID
	kind       EventKind
	message    string
}

type EventValues struct {
	ID         EventID
	PlanID     PlanID
	Kind       EventKind
	Message    string
	OccurredAt time.Time
}

func NewEvent(values EventValues) (Event, error) {
	if values.ID == "" || values.PlanID == "" || !values.Kind.Valid() || values.Message == "" ||
		values.OccurredAt.IsZero() || hasControl(values.Message) {
		return Event{}, fmt.Errorf("event: %w", ErrInvalidEvent)
	}
	return Event{
		id: values.ID, planID: values.PlanID, kind: values.Kind,
		message: values.Message, occurredAt: values.OccurredAt.UTC(),
	}, nil
}

func (e Event) ID() EventID           { return e.id }
func (e Event) PlanID() PlanID        { return e.planID }
func (e Event) Kind() EventKind       { return e.kind }
func (e Event) Message() string       { return e.message }
func (e Event) OccurredAt() time.Time { return e.occurredAt }

func (k EventKind) Valid() bool {
	switch k {
	case EventPlanCreated, EventPlanApproved, EventPlanStarted, EventStepStarted, EventStepOutput,
		EventStepSucceeded, EventPlanSucceeded, EventPlanFailed, EventPlanCanceled,
		EventPlanRetried, EventPlanRecovered:
		return true
	default:
		return false
	}
}
