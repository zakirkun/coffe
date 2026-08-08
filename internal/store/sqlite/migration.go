package sqlite

import (
	"context"
	"errors"
	"fmt"
)

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if err := s.ensurePlanColumns(ctx); err != nil {
		return err
	}
	return nil
}

type planColumn struct {
	name string
	ddl  string
}

func (s *Store) ensurePlanColumns(ctx context.Context) (err error) {
	rows, err := s.db.QueryContext(ctx, "PRAGMA table_info(plans)")
	if err != nil {
		return fmt.Errorf("inspect plan columns: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	existing := make(map[string]bool)
	for rows.Next() {
		var sequence, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&sequence, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("scan plan column: %w", err)
		}
		existing[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate plan columns: %w", err)
	}
	for _, column := range []planColumn{
		{name: "source_plan_id", ddl: "ALTER TABLE plans ADD COLUMN source_plan_id TEXT"},
		{name: "kind", ddl: "ALTER TABLE plans ADD COLUMN kind TEXT NOT NULL DEFAULT 'compose_deploy'"},
		{name: "revision", ddl: "ALTER TABLE plans ADD COLUMN revision TEXT"},
	} {
		if existing[column.name] {
			continue
		}
		if _, err := s.db.ExecContext(ctx, column.ddl); err != nil {
			return fmt.Errorf("add plan column %s: %w", column.name, err)
		}
	}
	return nil
}
