package dbtool

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

var ErrReadOnlyQuery = service.ErrReadOnlyQuery

type Reader struct {
	timeout time.Duration
	maxRows int
}

func NewReader(timeout time.Duration, maxRows int) *Reader {
	return &Reader{timeout: timeout, maxRows: maxRows}
}

func (r *Reader) Query(
	ctx context.Context,
	request service.DatabaseQuery,
) (result domain.QueryResult, err error) {
	query := strings.TrimSpace(request.SQL)
	if !isReadQuery(query) || request.Offset < 0 || request.Limit < 0 {
		return domain.QueryResult{}, ErrReadOnlyQuery
	}
	limit := request.Limit
	if limit == 0 || limit > r.maxRows {
		limit = r.maxRows
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	database, err := r.open(request.Connection, request.Password)
	if err != nil {
		return domain.QueryResult{}, err
	}
	defer func() { err = errors.Join(err, database.Close()) }()
	transaction, err := database.BeginTx(queryCtx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return domain.QueryResult{}, fmt.Errorf("begin read transaction: %w", err)
	}
	defer func() { err = errors.Join(err, transaction.Rollback()) }()
	rows, err := transaction.QueryContext(queryCtx, query)
	if err != nil {
		return domain.QueryResult{}, fmt.Errorf("query database: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	columns, err := rows.Columns()
	if err != nil {
		return domain.QueryResult{}, fmt.Errorf("query columns: %w", err)
	}
	result.Columns = columns
	skipped := 0
	for rows.Next() {
		if skipped < request.Offset {
			skipped++
			continue
		}
		if len(result.Rows) == limit {
			result.Truncated = true
			break
		}
		row, err := scanRow(rows, len(columns))
		if err != nil {
			return domain.QueryResult{}, err
		}
		result.Rows = append(result.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return domain.QueryResult{}, fmt.Errorf("iterate query rows: %w", err)
	}
	return result, nil
}

func (r *Reader) open(connection domain.DatabaseConnection, password string) (*sql.DB, error) {
	var database *sql.DB
	var err error
	switch connection.Driver() {
	case domain.DatabaseSQLite:
		database, err = sql.Open("sqlite", connection.SQLitePath())
	case domain.DatabaseMySQL:
		config := mysql.NewConfig()
		config.User = connection.User()
		config.Passwd = password
		config.Net = "tcp"
		config.Addr = net.JoinHostPort(connection.Host().String(), strconv.Itoa(connection.Port().Int()))
		config.DBName = connection.Database()
		config.ParseTime = true
		config.Timeout = r.timeout
		config.ReadTimeout = r.timeout
		config.WriteTimeout = r.timeout
		if connection.TLS() {
			config.TLSConfig = "true"
		}
		database, err = sql.Open("mysql", config.FormatDSN())
	default:
		return nil, fmt.Errorf("database driver %q: unsupported", connection.Driver())
	}
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if connection.Driver() == domain.DatabaseSQLite {
		database.SetMaxOpenConns(1)
	}
	database.SetConnMaxLifetime(3 * time.Minute)
	database.SetConnMaxIdleTime(time.Minute)
	if connection.Driver() == domain.DatabaseMySQL {
		database.SetMaxOpenConns(4)
		database.SetMaxIdleConns(4)
	}
	return database, nil
}

func isReadQuery(query string) bool {
	if query == "" || strings.ContainsRune(query, ';') || strings.ContainsRune(query, '\x00') ||
		strings.Contains(query, "/*") || strings.Contains(query, "--") || strings.Contains(query, "#") {
		return false
	}
	upper := strings.ToUpper(query)
	for _, forbidden := range []string{
		" INTO OUTFILE ", " INTO DUMPFILE ", " FOR UPDATE", " LOCK IN SHARE MODE", "LOAD_FILE(",
	} {
		if strings.Contains(upper, forbidden) {
			return false
		}
	}
	keyword := strings.ToUpper(strings.Fields(query)[0])
	switch keyword {
	case "SELECT", "SHOW", "DESCRIBE", "DESC", "EXPLAIN":
		return true
	default:
		return false
	}
}

func scanRow(rows *sql.Rows, width int) ([]string, error) {
	values := make([]any, width)
	destinations := make([]any, width)
	for index := range values {
		destinations[index] = &values[index]
	}
	if err := rows.Scan(destinations...); err != nil {
		return nil, fmt.Errorf("scan query row: %w", err)
	}
	result := make([]string, width)
	for index, value := range values {
		switch typed := value.(type) {
		case nil:
			result[index] = "NULL"
		case []byte:
			result[index] = string(typed)
		case time.Time:
			result[index] = typed.UTC().Format(time.RFC3339Nano)
		default:
			result[index] = fmt.Sprint(typed)
		}
	}
	return result, nil
}
