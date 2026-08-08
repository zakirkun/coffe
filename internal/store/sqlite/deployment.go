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

func (s *Store) SaveDeployment(ctx context.Context, record domain.DeploymentRecord) error {
	const query = `INSERT INTO deployments (
		plan_id, target_id, kind, revision_before, revision_after, health, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query,
		record.PlanID(), record.TargetID(), record.Kind(),
		record.RevisionBefore().String(), record.RevisionAfter().String(),
		record.Health(), formatTime(record.CreatedAt()),
	); err != nil {
		return fmt.Errorf("save deployment: %w", err)
	}
	return nil
}

func (s *Store) LatestDeployment(
	ctx context.Context,
	targetID domain.TargetID,
) (domain.DeploymentRecord, error) {
	const query = `SELECT plan_id, target_id, kind, revision_before, revision_after, health, created_at
		FROM deployments WHERE target_id = ? ORDER BY created_at DESC, plan_id DESC LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, targetID)
	var planID, rawTargetID, rawKind, rawBefore, rawAfter, rawHealth, rawCreatedAt string
	if err := row.Scan(
		&planID, &rawTargetID, &rawKind, &rawBefore, &rawAfter, &rawHealth, &rawCreatedAt,
	); errors.Is(err, sql.ErrNoRows) {
		return domain.DeploymentRecord{}, fmt.Errorf("deployment for target %s: %w", targetID, service.ErrNotFound)
	} else if err != nil {
		return domain.DeploymentRecord{}, fmt.Errorf("scan deployment: %w", err)
	}
	before, err := domain.NewGitRevision(rawBefore)
	if err != nil {
		return domain.DeploymentRecord{}, fmt.Errorf("restore before revision: %w", err)
	}
	after, err := domain.NewGitRevision(rawAfter)
	if err != nil {
		return domain.DeploymentRecord{}, fmt.Errorf("restore after revision: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.DeploymentRecord{}, fmt.Errorf("parse deployment created_at: %w", err)
	}
	record, err := domain.NewDeploymentRecord(domain.DeploymentValues{
		PlanID: domain.PlanID(planID), TargetID: domain.TargetID(rawTargetID), Kind: domain.PlanKind(rawKind),
		RevisionBefore: before, RevisionAfter: after,
		Health: domain.HealthStatus(rawHealth), CreatedAt: createdAt,
	})
	if err != nil {
		return domain.DeploymentRecord{}, fmt.Errorf("restore deployment: %w", err)
	}
	return record, nil
}
