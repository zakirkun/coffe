package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"path/filepath"
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
	if len(history) != 1 || history[0].Statement() != "SELECT 1" {
		t.Fatalf("query history = %v", history)
	}
}
