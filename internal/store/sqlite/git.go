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

func (s *Store) SaveGitOperation(ctx context.Context, operation domain.GitOperation) error {
	var approvedAt sql.NullString
	if value, ok := operation.ApprovedAt(); ok {
		approvedAt = sql.NullString{String: formatTime(value), Valid: true}
	}
	const query = `INSERT INTO git_operations (
		id, project_id, kind, remote, branch, status, created_at, approved_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET status = excluded.status, approved_at = excluded.approved_at`
	if _, err := s.db.ExecContext(ctx, query,
		operation.ID(), operation.ProjectID(), operation.Kind(), operation.Remote().String(),
		operation.Branch().String(), operation.Status(), formatTime(operation.CreatedAt()), approvedAt,
	); err != nil {
		return fmt.Errorf("save git operation: %w", err)
	}
	return nil
}

func (s *Store) GetGitOperation(
	ctx context.Context,
	id domain.GitOperationID,
) (domain.GitOperation, error) {
	const query = `SELECT id, project_id, kind, remote, branch, status, created_at, approved_at
		FROM git_operations WHERE id = ?`
	operation, err := scanGitOperation(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GitOperation{}, fmt.Errorf("git operation %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.GitOperation{}, err
	}
	return operation, nil
}

func (s *Store) ListGitOperations(
	ctx context.Context,
	limit int,
) (operations []domain.GitOperation, err error) {
	const query = `SELECT id, project_id, kind, remote, branch, status, created_at, approved_at
		FROM git_operations ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query git operations: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		operation, err := scanGitOperation(rows)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate git operations: %w", err)
	}
	return operations, nil
}

func (s *Store) AppendGitEvent(
	ctx context.Context,
	operationID domain.GitOperationID,
	message string,
	occurredAt time.Time,
) error {
	const query = `INSERT INTO git_events (operation_id, message, occurred_at) VALUES (?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query, operationID, message, formatTime(occurredAt)); err != nil {
		return fmt.Errorf("append git event: %w", err)
	}
	return nil
}

func (s *Store) ListGitEvents(
	ctx context.Context,
	operationID domain.GitOperationID,
) (events []domain.GitEvent, err error) {
	const query = `SELECT message, occurred_at FROM git_events WHERE operation_id = ? ORDER BY sequence`
	rows, err := s.db.QueryContext(ctx, query, operationID)
	if err != nil {
		return nil, fmt.Errorf("query git events: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var message, rawOccurredAt string
		if err := rows.Scan(&message, &rawOccurredAt); err != nil {
			return nil, fmt.Errorf("scan git event: %w", err)
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, rawOccurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse git event time: %w", err)
		}
		event, err := domain.NewGitEvent(operationID, message, occurredAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate git events: %w", err)
	}
	return events, nil
}

func scanGitOperation(row rowScanner) (domain.GitOperation, error) {
	var id, projectID, rawKind, rawRemote, rawBranch, rawStatus, rawCreatedAt string
	var rawApprovedAt sql.NullString
	if err := row.Scan(
		&id, &projectID, &rawKind, &rawRemote, &rawBranch, &rawStatus, &rawCreatedAt, &rawApprovedAt,
	); err != nil {
		return domain.GitOperation{}, fmt.Errorf("scan git operation: %w", err)
	}
	remote, err := domain.NewGitRemote(rawRemote)
	if err != nil {
		return domain.GitOperation{}, err
	}
	branch, err := domain.NewBranch(rawBranch)
	if err != nil {
		return domain.GitOperation{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.GitOperation{}, fmt.Errorf("parse git created_at: %w", err)
	}
	approvedAt, err := parseNullableTime(rawApprovedAt)
	if err != nil {
		return domain.GitOperation{}, fmt.Errorf("parse git approved_at: %w", err)
	}
	operation, err := domain.NewGitOperation(domain.GitOperationValues{
		ID: domain.GitOperationID(id), ProjectID: domain.ProjectID(projectID),
		Kind: domain.GitOperationKind(rawKind), Remote: remote, Branch: branch,
		Status: domain.GitOperationStatus(rawStatus), CreatedAt: createdAt, ApprovedAt: approvedAt,
	})
	if err != nil {
		return domain.GitOperation{}, fmt.Errorf("restore git operation: %w", err)
	}
	return operation, nil
}
