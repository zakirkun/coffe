package sqlite

import (
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *Store) SaveQueryHistory(ctx context.Context, history domain.QueryHistory) error {
	const query = `INSERT INTO query_history (
		id, database_id, statement, row_count, duration_millis, truncated, success, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query,
		history.ID(), history.DatabaseID(), history.Statement(), history.RowCount(),
		history.DurationMillis(), history.Truncated(), history.Success(), formatTime(history.CreatedAt()),
	); err != nil {
		return fmt.Errorf("save query history: %w", err)
	}
	return nil
}

func (s *Store) ListQueryHistory(
	ctx context.Context,
	databaseID domain.DatabaseID,
	limit int,
) (history []domain.QueryHistory, err error) {
	const query = `SELECT id, database_id, statement, row_count, duration_millis, truncated, success, created_at
		FROM query_history WHERE database_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, databaseID, limit)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var id, rawDatabaseID, statement, rawCreatedAt string
		var rowCount int
		var durationMillis int64
		var truncated, success bool
		if err := rows.Scan(
			&id, &rawDatabaseID, &statement, &rowCount, &durationMillis,
			&truncated, &success, &rawCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan query history: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse query history created_at: %w", err)
		}
		record, err := domain.NewQueryHistory(domain.QueryHistoryValues{
			ID: domain.QueryHistoryID(id), DatabaseID: domain.DatabaseID(rawDatabaseID), Statement: statement,
			RowCount: rowCount, DurationMillis: durationMillis,
			Truncated: truncated, Success: success, CreatedAt: createdAt,
		})
		if err != nil {
			return nil, err
		}
		history = append(history, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate query history: %w", err)
	}
	return history, nil
}
