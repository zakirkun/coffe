package dbtool_test

import (
	"coffe/internal/dbtool"
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func Test_Reader_queries_SQLite_and_rejects_mutation(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "analytics.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := database.Exec(`CREATE TABLE metrics (name TEXT, value INTEGER);
		INSERT INTO metrics VALUES ('jobs', 7), ('agents', 8), ('queries', 9)`); err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	connection := sqliteConnection(t, path)
	reader := dbtool.NewReader(2*time.Second, 100)

	// When
	result, err := reader.Query(context.Background(), service.DatabaseQuery{
		Connection: connection, SQL: "SELECT name, value FROM metrics ORDER BY value",
		Offset: 1, Limit: 1,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	_, mutationErr := reader.Query(context.Background(), service.DatabaseQuery{
		Connection: connection, SQL: "DELETE FROM metrics",
	})
	_, outfileErr := reader.Query(context.Background(), service.DatabaseQuery{
		Connection: connection, SQL: "SELECT name INTO OUTFILE '/tmp/leak' FROM metrics",
	})

	// Then
	if len(result.Rows) != 1 || result.Rows[0][0] != "agents" || !result.Truncated {
		t.Fatalf("Query() rows = %v", result.Rows)
	}
	if !errors.Is(mutationErr, dbtool.ErrReadOnlyQuery) {
		t.Fatalf("Query(DELETE) error = %v", mutationErr)
	}
	if !errors.Is(outfileErr, dbtool.ErrReadOnlyQuery) {
		t.Fatalf("Query(INTO OUTFILE) error = %v", outfileErr)
	}
}

func sqliteConnection(t *testing.T, path string) domain.DatabaseConnection {
	t.Helper()
	name, err := domain.NewName("analytics")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	connection, err := domain.NewDatabaseConnection(domain.DatabaseValues{
		ID: "db-1", Name: name, Driver: domain.DatabaseSQLite, SQLitePath: path,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewDatabaseConnection() error = %v", err)
	}
	return connection
}
