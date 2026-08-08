package sqlite

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *Store) AppendEvent(ctx context.Context, event domain.Event) error {
	const query = `INSERT INTO events (id, plan_id, kind, message, occurred_at) VALUES (?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(
		ctx, query, event.ID(), event.PlanID(), event.Kind(), event.Message(), formatTime(event.OccurredAt()),
	); err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (s *Store) ListEvents(ctx context.Context, planID domain.PlanID) (events []domain.Event, err error) {
	const query = `SELECT id, plan_id, kind, message, occurred_at
        FROM events WHERE plan_id = ? ORDER BY occurred_at, id`
	rows, err := s.db.QueryContext(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var id, rawPlanID, rawKind, message, rawOccurredAt string
		if err := rows.Scan(&id, &rawPlanID, &rawKind, &message, &rawOccurredAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, rawOccurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse event occurred_at: %w", err)
		}
		event, err := domain.NewEvent(domain.EventValues{
			ID: domain.EventID(id), PlanID: domain.PlanID(rawPlanID), Kind: domain.EventKind(rawKind),
			Message: message, OccurredAt: occurredAt,
		})
		if err != nil {
			return nil, fmt.Errorf("restore event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return events, nil
}
