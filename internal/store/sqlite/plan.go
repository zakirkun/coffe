package sqlite

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *Store) SavePlan(ctx context.Context, plan domain.ActionPlan) error {
	stepsJSON, err := json.Marshal(plan.Steps())
	if err != nil {
		return fmt.Errorf("encode plan steps: %w", err)
	}
	var approvedAt sql.NullString
	if value, ok := plan.ApprovedAt(); ok {
		approvedAt = sql.NullString{String: formatTime(value), Valid: true}
	}
	var revision sql.NullString
	if value, ok := plan.Revision(); ok {
		revision = sql.NullString{String: value.String(), Valid: true}
	}
	var sourceID sql.NullString
	if plan.SourcePlanID() != "" {
		sourceID = sql.NullString{String: string(plan.SourcePlanID()), Valid: true}
	}
	const query = `INSERT INTO plans (
		id, project_id, target_id, source_plan_id, kind, status, revision, steps_json, created_at, approved_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET status = excluded.status, approved_at = excluded.approved_at`
	if _, err := s.db.ExecContext(ctx, query,
		plan.ID(), plan.ProjectID(), plan.TargetID(), sourceID, plan.Kind(), plan.Status(), revision, string(stepsJSON),
		formatTime(plan.CreatedAt()), approvedAt,
	); err != nil {
		return fmt.Errorf("save plan: %w", err)
	}
	return nil
}

func (s *Store) GetPlan(ctx context.Context, id domain.PlanID) (domain.ActionPlan, error) {
	const query = `SELECT id, project_id, target_id, source_plan_id, kind, status, revision, steps_json, created_at, approved_at
		FROM plans WHERE id = ?`
	plan, err := scanPlan(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ActionPlan{}, fmt.Errorf("plan %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.ActionPlan{}, err
	}
	return plan, nil
}

func (s *Store) ListPlans(ctx context.Context, limit int) (plans []domain.ActionPlan, err error) {
	const query = `SELECT id, project_id, target_id, source_plan_id, kind, status, revision, steps_json, created_at, approved_at
		FROM plans ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query plans: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	plans = make([]domain.ActionPlan, 0, limit)
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plans: %w", err)
	}
	return plans, nil
}

func scanPlan(row rowScanner) (domain.ActionPlan, error) {
	var id, projectID, targetID, rawKind, rawStatus, rawSteps, rawCreatedAt string
	var rawSourceID, rawRevision, rawApprovedAt sql.NullString
	if err := row.Scan(
		&id, &projectID, &targetID, &rawSourceID, &rawKind, &rawStatus,
		&rawRevision, &rawSteps, &rawCreatedAt, &rawApprovedAt,
	); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("scan plan: %w", err)
	}
	var steps []domain.StepKind
	if err := json.Unmarshal([]byte(rawSteps), &steps); err != nil {
		return domain.ActionPlan{}, fmt.Errorf("decode plan steps: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("parse plan created_at: %w", err)
	}
	var approvedAt time.Time
	if rawApprovedAt.Valid {
		approvedAt, err = time.Parse(time.RFC3339Nano, rawApprovedAt.String)
		if err != nil {
			return domain.ActionPlan{}, fmt.Errorf("parse plan approved_at: %w", err)
		}
	}
	var revision domain.GitRevision
	if rawRevision.Valid {
		revision, err = domain.NewGitRevision(rawRevision.String)
		if err != nil {
			return domain.ActionPlan{}, fmt.Errorf("parse plan revision: %w", err)
		}
	}
	plan, err := domain.RestoreActionPlan(domain.PlanValues{
		ID: domain.PlanID(id), ProjectID: domain.ProjectID(projectID), TargetID: domain.TargetID(targetID),
		SourceID: domain.PlanID(rawSourceID.String), Kind: domain.PlanKind(rawKind),
		Status: domain.PlanStatus(rawStatus), Revision: revision,
		Steps: steps, CreatedAt: createdAt, ApprovedAt: approvedAt,
	})
	if err != nil {
		return domain.ActionPlan{}, fmt.Errorf("restore plan: %w", err)
	}
	return plan, nil
}
