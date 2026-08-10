package sqlite

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) SaveAgentSession(ctx context.Context, session domain.AgentSession) error {
	startedAt := nullableTime(session.StartedAt())
	finishedAt := nullableTime(session.FinishedAt())
	const query = `INSERT INTO agent_sessions (
		id, project_id, tool, status, created_at, started_at, finished_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		status = excluded.status, started_at = excluded.started_at, finished_at = excluded.finished_at`
	if _, err := s.db.ExecContext(ctx, query,
		session.ID(), session.ProjectID(), session.Tool(), session.Status(),
		formatTime(session.CreatedAt()), startedAt, finishedAt,
	); err != nil {
		return fmt.Errorf("save agent session: %w", err)
	}
	return nil
}

func (s *Store) GetAgentSession(
	ctx context.Context,
	id domain.AgentSessionID,
) (domain.AgentSession, error) {
	const query = `SELECT id, project_id, tool, status, created_at, started_at, finished_at
		FROM agent_sessions WHERE id = ?`
	session, err := scanAgentSession(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AgentSession{}, fmt.Errorf("agent session %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.AgentSession{}, err
	}
	return session, nil
}

func (s *Store) ListAgentSessions(ctx context.Context, limit int) (sessions []domain.AgentSession, err error) {
	const query = `SELECT id, project_id, tool, status, created_at, started_at, finished_at
		FROM agent_sessions ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query agent sessions: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		session, err := scanAgentSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agent sessions: %w", err)
	}
	return sessions, nil
}

func (s *Store) ListRunningAgentSessions(ctx context.Context) (sessions []domain.AgentSession, err error) {
	const query = `SELECT id, project_id, tool, status, created_at, started_at, finished_at
		FROM agent_sessions WHERE status = ? ORDER BY created_at DESC, id DESC`
	rows, err := s.db.QueryContext(ctx, query, domain.AgentRunning)
	if err != nil {
		return nil, fmt.Errorf("query running agent sessions: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		session, err := scanAgentSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate running agent sessions: %w", err)
	}
	return sessions, nil
}

func (s *Store) AppendAgentEvent(ctx context.Context, event domain.AgentEvent) error {
	const query = `INSERT INTO agent_events (session_id, message, occurred_at) VALUES (?, ?, ?)`
	if _, err := s.db.ExecContext(
		ctx, query, event.SessionID(), event.Message(), formatTime(event.OccurredAt()),
	); err != nil {
		return fmt.Errorf("append agent event: %w", err)
	}
	return nil
}

func (s *Store) ListAgentEvents(
	ctx context.Context,
	sessionID domain.AgentSessionID,
) (events []domain.AgentEvent, err error) {
	const query = `SELECT message, occurred_at FROM agent_events WHERE session_id = ? ORDER BY sequence`
	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query agent events: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var message, rawOccurredAt string
		if err := rows.Scan(&message, &rawOccurredAt); err != nil {
			return nil, fmt.Errorf("scan agent event: %w", err)
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, rawOccurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse agent event time: %w", err)
		}
		event, err := domain.NewAgentEvent(sessionID, message, occurredAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agent events: %w", err)
	}
	return events, nil
}

func scanAgentSession(row rowScanner) (domain.AgentSession, error) {
	var id, projectID, rawTool, rawStatus, rawCreatedAt string
	var rawStartedAt, rawFinishedAt sql.NullString
	if err := row.Scan(
		&id, &projectID, &rawTool, &rawStatus, &rawCreatedAt, &rawStartedAt, &rawFinishedAt,
	); err != nil {
		return domain.AgentSession{}, fmt.Errorf("scan agent session: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.AgentSession{}, fmt.Errorf("parse agent created_at: %w", err)
	}
	startedAt, err := parseNullableTime(rawStartedAt)
	if err != nil {
		return domain.AgentSession{}, fmt.Errorf("parse agent started_at: %w", err)
	}
	finishedAt, err := parseNullableTime(rawFinishedAt)
	if err != nil {
		return domain.AgentSession{}, fmt.Errorf("parse agent finished_at: %w", err)
	}
	session, err := domain.NewAgentSession(domain.AgentSessionValues{
		ID: domain.AgentSessionID(id), ProjectID: domain.ProjectID(projectID),
		Tool: domain.AgentTool(rawTool), Status: domain.AgentStatus(rawStatus),
		CreatedAt: createdAt, StartedAt: startedAt, FinishedAt: finishedAt,
	})
	if err != nil {
		return domain.AgentSession{}, fmt.Errorf("restore agent session: %w", err)
	}
	return session, nil
}

func nullableTime(value time.Time) sql.NullString {
	return sql.NullString{String: formatTime(value), Valid: !value.IsZero()}
}

func parseNullableTime(value sql.NullString) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, value.String)
}
