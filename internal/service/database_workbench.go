package service

import (
	"bytes"
	"coffe/internal/domain"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type QueryPageInput struct {
	SQL    string
	Offset int
	Limit  int
}

func (a *App) QueryDatabasePage(
	ctx context.Context,
	rawID string,
	input QueryPageInput,
) (domain.QueryResult, error) {
	connection, password, err := a.databaseAccess(ctx, rawID)
	if err != nil {
		return domain.QueryResult{}, err
	}
	startedAt := a.clock.Now()
	result, queryErr := a.databases.Query(ctx, DatabaseQuery{
		Connection: connection, Password: password, SQL: input.SQL,
		Offset: input.Offset, Limit: input.Limit,
	})
	historyErr := a.recordQueryHistory(ctx, queryRecordInput{
		databaseID: connection.ID(), query: input.SQL, result: result,
		success: queryErr == nil, startedAt: startedAt,
	})
	outcome := domain.AuditSuccess
	message := "read-only database query succeeded"
	if queryErr != nil {
		outcome = domain.AuditFailure
		message = "read-only database query failed"
	}
	auditErr := a.appendAudit(ctx, auditInput{
		category: domain.AuditDatabase, action: "query", resourceID: string(connection.ID()),
		outcome: outcome, message: message,
	})
	if queryErr != nil || historyErr != nil || auditErr != nil {
		return domain.QueryResult{}, errors.Join(queryErr, historyErr, auditErr)
	}
	return result, nil
}

func (a *App) ExplainDatabase(
	ctx context.Context,
	rawID string,
	query string,
) (domain.QueryResult, error) {
	return a.QueryDatabasePage(ctx, rawID, QueryPageInput{SQL: "EXPLAIN " + strings.TrimSpace(query)})
}

func (a *App) ExportDatabaseCSV(
	ctx context.Context,
	rawID string,
	query string,
) ([]byte, error) {
	result, err := a.QueryDatabasePage(ctx, rawID, QueryPageInput{SQL: query, Limit: 200})
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write(result.Columns); err != nil {
		return nil, fmt.Errorf("write CSV header: %w", err)
	}
	for _, row := range result.Rows {
		clean := make([]string, len(row))
		for index, value := range row {
			clean[index] = safeCSVCell(value)
		}
		if err := writer.Write(clean); err != nil {
			return nil, fmt.Errorf("write CSV row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush CSV: %w", err)
	}
	return output.Bytes(), nil
}

func (a *App) ListQueryHistory(
	ctx context.Context,
	rawID string,
	limit int,
) ([]domain.QueryHistory, error) {
	id := domain.DatabaseID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID || limit < 1 || limit > 200 {
		return nil, fmt.Errorf("query history input: %w", ErrInvalidInput)
	}
	history, err := a.repository.ListQueryHistory(ctx, id, limit)
	if err != nil {
		return nil, fmt.Errorf("list query history: %w", err)
	}
	return history, nil
}

type queryRecordInput struct {
	databaseID domain.DatabaseID
	query      string
	result     domain.QueryResult
	success    bool
	startedAt  time.Time
}

func (a *App) recordQueryHistory(ctx context.Context, input queryRecordInput) error {
	rawID, err := a.ids.New()
	if err != nil {
		return fmt.Errorf("new query history id: %w", err)
	}
	statement := strings.Join(strings.Fields(redactSQLLiterals(input.query)), " ")
	if utf8.RuneCountInString(statement) > 4096 {
		statement = string([]rune(statement)[:4096])
	}
	duration := a.clock.Now().Sub(input.startedAt)
	if duration < 0 {
		duration = 0
	}
	history, err := domain.NewQueryHistory(domain.QueryHistoryValues{
		ID: domain.QueryHistoryID(rawID), DatabaseID: input.databaseID, Statement: statement,
		RowCount: len(input.result.Rows), DurationMillis: duration.Milliseconds(),
		Truncated: input.result.Truncated, Success: input.success, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return err
	}
	if err := a.repository.SaveQueryHistory(ctx, history); err != nil {
		return fmt.Errorf("save query history: %w", err)
	}
	return nil
}

func safeCSVCell(value string) string {
	if value == "" {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
}
