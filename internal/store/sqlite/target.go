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

func (s *Store) CreateTarget(ctx context.Context, target domain.Target) error {
	const query = `INSERT INTO targets (
        id, project_id, name, host, port, ssh_user, remote_path, branch, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query,
		target.ID(), target.ProjectID(), target.Name().String(), target.Host().String(), target.Port().Int(),
		target.User().String(), target.RemotePath().String(), target.Branch().String(), formatTime(target.CreatedAt()),
	); err != nil {
		return fmt.Errorf("insert target: %w", err)
	}
	return nil
}

func (s *Store) ListTargets(ctx context.Context, projectID domain.ProjectID) (targets []domain.Target, err error) {
	const query = `SELECT id, project_id, name, host, port, ssh_user, remote_path, branch, created_at
        FROM targets WHERE project_id = ? ORDER BY name, id`
	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("query targets: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		target, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate targets: %w", err)
	}
	return targets, nil
}

func (s *Store) GetTarget(ctx context.Context, id domain.TargetID) (domain.Target, error) {
	const query = `SELECT id, project_id, name, host, port, ssh_user, remote_path, branch, created_at
        FROM targets WHERE id = ?`
	target, err := scanTarget(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Target{}, fmt.Errorf("target %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.Target{}, err
	}
	return target, nil
}

func scanTarget(row rowScanner) (domain.Target, error) {
	var raw storedTarget
	if err := row.Scan(
		&raw.id, &raw.projectID, &raw.name, &raw.host, &raw.port,
		&raw.user, &raw.path, &raw.branch, &raw.createdAt,
	); err != nil {
		return domain.Target{}, fmt.Errorf("scan target: %w", err)
	}
	values, err := restoreTargetValues(raw)
	if err != nil {
		return domain.Target{}, err
	}
	target, err := domain.NewTarget(values)
	if err != nil {
		return domain.Target{}, fmt.Errorf("restore target: %w", err)
	}
	return target, nil
}

type storedTarget struct {
	id        string
	projectID string
	name      string
	host      string
	port      int
	user      string
	path      string
	branch    string
	createdAt string
}

func restoreTargetValues(raw storedTarget) (domain.TargetValues, error) {
	name, err := domain.NewName(raw.name)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("stored target name: %w", err)
	}
	host, err := domain.NewHost(raw.host)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("stored target host: %w", err)
	}
	port, err := domain.NewPort(raw.port)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("stored target port: %w", err)
	}
	user, err := domain.NewSSHUser(raw.user)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("stored target user: %w", err)
	}
	remotePath, err := domain.NewRemotePath(raw.path)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("stored target path: %w", err)
	}
	branch, err := domain.NewBranch(raw.branch)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("stored target branch: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, raw.createdAt)
	if err != nil {
		return domain.TargetValues{}, fmt.Errorf("parse target created_at: %w", err)
	}
	return domain.TargetValues{
		ID: domain.TargetID(raw.id), ProjectID: domain.ProjectID(raw.projectID), Name: name, Host: host,
		Port: port, User: user, RemotePath: remotePath, Branch: branch, CreatedAt: createdAt,
	}, nil
}
