package dbtool

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"fmt"
	"regexp"
)

var tableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (r *Reader) Tables(ctx context.Context, request service.DatabaseAccess) ([]domain.DatabaseTable, error) {
	query := "SELECT name, type FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY name"
	if request.Connection.Driver() == domain.DatabaseMySQL {
		query = "SHOW FULL TABLES"
	}
	result, err := r.Query(ctx, service.DatabaseQuery{
		Connection: request.Connection, Password: request.Password, SQL: query,
	})
	if err != nil {
		return nil, err
	}
	tables := make([]domain.DatabaseTable, 0, len(result.Rows))
	for _, row := range result.Rows {
		if len(row) < 2 {
			return nil, fmt.Errorf("database table row has %d columns", len(row))
		}
		tables = append(tables, domain.DatabaseTable{Name: row[0], Type: row[1]})
	}
	return tables, nil
}

func (r *Reader) Columns(
	ctx context.Context,
	request service.DatabaseAccess,
	tableName string,
) ([]domain.DatabaseColumn, error) {
	if !tableNamePattern.MatchString(tableName) {
		return nil, fmt.Errorf("table name %q: %w", tableName, ErrReadOnlyQuery)
	}
	query := "SELECT name, type, notnull FROM pragma_table_info('" + tableName + "')"
	if request.Connection.Driver() == domain.DatabaseMySQL {
		query = "DESCRIBE `" + tableName + "`"
	}
	result, err := r.Query(ctx, service.DatabaseQuery{
		Connection: request.Connection, Password: request.Password, SQL: query,
	})
	if err != nil {
		return nil, err
	}
	columns := make([]domain.DatabaseColumn, 0, len(result.Rows))
	for _, row := range result.Rows {
		column, err := mapColumn(request.Connection.Driver(), row)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func mapColumn(driver domain.DatabaseDriver, row []string) (domain.DatabaseColumn, error) {
	if len(row) < 3 {
		return domain.DatabaseColumn{}, fmt.Errorf("database column row has %d fields", len(row))
	}
	if driver == domain.DatabaseSQLite {
		return domain.DatabaseColumn{Name: row[0], DataType: row[1], Nullable: row[2] == "0"}, nil
	}
	return domain.DatabaseColumn{Name: row[0], DataType: row[1], Nullable: row[2] == "YES"}, nil
}
