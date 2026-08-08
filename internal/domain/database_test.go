package domain_test

import (
	"coffe/internal/domain"
	"errors"
	"testing"
	"time"
)

func Test_NewDatabaseConnection_accepts_local_sqlite(t *testing.T) {
	// Given
	name, err := domain.NewName("analytics")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}

	// When
	connection, err := domain.NewDatabaseConnection(domain.DatabaseValues{
		ID: "db-1", Name: name, Driver: domain.DatabaseSQLite,
		SQLitePath: t.TempDir() + "/analytics.db",
		CreatedAt:  time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	// Then
	if err != nil {
		t.Fatalf("NewDatabaseConnection() error = %v", err)
	}
	if connection.Driver() != domain.DatabaseSQLite {
		t.Fatalf("Driver() = %q", connection.Driver())
	}
}

func Test_NewDatabaseConnection_rejects_mysql_without_secret_reference(t *testing.T) {
	// Given
	name, nameErr := domain.NewName("production")
	host, hostErr := domain.NewHost("db.example.com")
	port, portErr := domain.NewPort(3306)
	if err := errors.Join(nameErr, hostErr, portErr); err != nil {
		t.Fatalf("database value error = %v", err)
	}

	// When
	_, err := domain.NewDatabaseConnection(domain.DatabaseValues{
		ID: "db-1", Name: name, Driver: domain.DatabaseMySQL,
		Host: host, Port: port, User: "reader", Database: "app",
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})

	// Then
	if !errors.Is(err, domain.ErrInvalidDatabase) {
		t.Fatalf("NewDatabaseConnection() error = %v", err)
	}
}
