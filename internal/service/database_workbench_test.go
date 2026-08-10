package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func Test_App_QueryDatabasePage_records_history(t *testing.T) {
	// Given
	fixture := newFixture(t)
	connection, err := fixture.app.CreateDatabase(context.Background(), service.CreateDatabaseInput{
		Name: "analytics", Driver: string(domain.DatabaseSQLite),
		SQLitePath: filepath.Join(t.TempDir(), "analytics.db"),
	})
	if err != nil {
		t.Fatalf("CreateDatabase() error = %v", err)
	}

	// When
	_, err = fixture.app.QueryDatabasePage(context.Background(), string(connection.ID()), service.QueryPageInput{
		SQL: "SELECT 1", Offset: 0, Limit: 25,
	})
	// Then
	if err != nil {
		t.Fatalf("QueryDatabasePage() error = %v", err)
	}
	history, err := fixture.repo.ListQueryHistory(context.Background(), connection.ID(), 10)
	if err != nil {
		t.Fatalf("ListQueryHistory() error = %v", err)
	}
	if len(history) != 1 || history[0].Statement() != "SELECT ?" {
		t.Fatalf("query history = %v", history)
	}
}

func Test_App_QueryDatabasePage_redacts_literals_from_history(t *testing.T) {
	// Given
	fixture := newFixture(t)
	connection, err := fixture.app.CreateDatabase(context.Background(), service.CreateDatabaseInput{
		Name: "analytics", Driver: string(domain.DatabaseSQLite),
		SQLitePath: filepath.Join(t.TempDir(), "analytics.db"),
	})
	if err != nil {
		t.Fatalf("CreateDatabase() error = %v", err)
	}
	query := "SELECT email FROM users WHERE password = 's3cr''et' AND api_key = 48291 AND flags = 0b1010 AND mask = 0B0101"

	// When
	_, err = fixture.app.QueryDatabasePage(context.Background(), string(connection.ID()), service.QueryPageInput{
		SQL: query, Limit: 25,
	})

	// Then
	if err != nil {
		t.Fatalf("QueryDatabasePage() error = %v", err)
	}
	history, err := fixture.repo.ListQueryHistory(context.Background(), connection.ID(), 10)
	if err != nil {
		t.Fatalf("ListQueryHistory() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("query history count = %d, want 1", len(history))
	}
	statement := history[0].Statement()
	if strings.Contains(statement, "s3cr") || strings.Contains(statement, "48291") ||
		strings.Contains(statement, "0b1010") || strings.Contains(statement, "0B0101") {
		t.Fatalf("query history retained literal: %q", statement)
	}
	if statement != "SELECT email FROM users WHERE password = ? AND api_key = ? AND flags = ? AND mask = ?" {
		t.Fatalf("query history statement = %q", statement)
	}
}
