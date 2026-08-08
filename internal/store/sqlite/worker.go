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

func (s *Store) SaveWorker(ctx context.Context, worker domain.Worker) error {
	lastSeen := nullableTime(worker.LastSeen())
	const query = `INSERT INTO workers (
		id, target_id, name, base_url, secret_name, version, enabled, created_at, last_seen
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		target_id = excluded.target_id, name = excluded.name, base_url = excluded.base_url,
		secret_name = excluded.secret_name, version = excluded.version,
		enabled = excluded.enabled, last_seen = excluded.last_seen`
	if _, err := s.db.ExecContext(ctx, query,
		worker.ID(), worker.TargetID(), worker.Name().String(), worker.BaseURL(), worker.SecretName(),
		worker.Version(), worker.Enabled(), formatTime(worker.CreatedAt()), lastSeen,
	); err != nil {
		return fmt.Errorf("save worker: %w", err)
	}
	return nil
}

func (s *Store) GetWorker(ctx context.Context, id domain.WorkerID) (domain.Worker, error) {
	const query = `SELECT id, target_id, name, base_url, secret_name, version, enabled, created_at, last_seen
		FROM workers WHERE id = ?`
	worker, err := scanWorker(s.db.QueryRowContext(ctx, query, id))
	return workerOrNotFound(worker, err)
}

func (s *Store) GetWorkerByTarget(ctx context.Context, id domain.TargetID) (domain.Worker, error) {
	const query = `SELECT id, target_id, name, base_url, secret_name, version, enabled, created_at, last_seen
		FROM workers WHERE target_id = ?`
	worker, err := scanWorker(s.db.QueryRowContext(ctx, query, id))
	return workerOrNotFound(worker, err)
}

func (s *Store) ListWorkers(ctx context.Context) (workers []domain.Worker, err error) {
	const query = `SELECT id, target_id, name, base_url, secret_name, version, enabled, created_at, last_seen
		FROM workers ORDER BY name, id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query workers: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		worker, err := scanWorker(rows)
		if err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workers: %w", err)
	}
	return workers, nil
}

func workerOrNotFound(worker domain.Worker, err error) (domain.Worker, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Worker{}, service.ErrNotFound
	}
	if err != nil {
		return domain.Worker{}, err
	}
	return worker, nil
}

func scanWorker(row rowScanner) (domain.Worker, error) {
	var id, targetID, rawName, baseURL, secretName, version, rawCreatedAt string
	var enabled bool
	var rawLastSeen sql.NullString
	if err := row.Scan(
		&id, &targetID, &rawName, &baseURL, &secretName,
		&version, &enabled, &rawCreatedAt, &rawLastSeen,
	); err != nil {
		return domain.Worker{}, fmt.Errorf("scan worker: %w", err)
	}
	name, err := domain.NewName(rawName)
	if err != nil {
		return domain.Worker{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.Worker{}, fmt.Errorf("parse worker created_at: %w", err)
	}
	lastSeen, err := parseNullableTime(rawLastSeen)
	if err != nil {
		return domain.Worker{}, fmt.Errorf("parse worker last_seen: %w", err)
	}
	worker, err := domain.NewWorker(domain.WorkerValues{
		ID: domain.WorkerID(id), TargetID: domain.TargetID(targetID), Name: name,
		BaseURL: baseURL, SecretName: secretName, Version: version,
		Enabled: enabled, CreatedAt: createdAt, LastSeen: lastSeen,
	})
	if err != nil {
		return domain.Worker{}, fmt.Errorf("restore worker: %w", err)
	}
	return worker, nil
}
