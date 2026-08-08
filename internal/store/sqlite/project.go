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

func (s *Store) CreateProject(ctx context.Context, project domain.Project) error {
	const query = `INSERT INTO projects (id, name, root_path, created_at) VALUES (?, ?, ?, ?)`
	if _, err := s.db.ExecContext(
		ctx, query, project.ID(), project.Name().String(), project.RootPath(), formatTime(project.CreatedAt()),
	); err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (s *Store) ListProjects(ctx context.Context) (projects []domain.Project, err error) {
	const query = `SELECT id, name, root_path, created_at FROM projects ORDER BY name, id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (s *Store) GetProject(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
	const query = `SELECT id, name, root_path, created_at FROM projects WHERE id = ?`
	project, err := scanProject(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Project{}, fmt.Errorf("project %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

type rowScanner interface{ Scan(...any) error }

func scanProject(row rowScanner) (domain.Project, error) {
	var id, rawName, rootPath, rawCreatedAt string
	if err := row.Scan(&id, &rawName, &rootPath, &rawCreatedAt); err != nil {
		return domain.Project{}, fmt.Errorf("scan project: %w", err)
	}
	name, err := domain.NewName(rawName)
	if err != nil {
		return domain.Project{}, fmt.Errorf("stored project name: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("parse project created_at: %w", err)
	}
	project, err := domain.NewProject(domain.ProjectValues{
		ID: domain.ProjectID(id), Name: name, RootPath: rootPath, CreatedAt: createdAt,
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("restore project: %w", err)
	}
	return project, nil
}

func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
