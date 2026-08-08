package sqlite_test

import (
	"coffe/internal/domain"
	"context"
	"path/filepath"
	"testing"
	"time"

	coffesqlite "coffe/internal/store/sqlite"
)

func Test_Store_roundtrips_agent_and_database_metadata(t *testing.T) {
	// Given
	ctx := context.Background()
	store, err := coffesqlite.Open(ctx, filepath.Join(t.TempDir(), "coffe.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	project := newProject(t, now)
	if err := store.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	session, err := domain.NewAgentSession(domain.AgentSessionValues{
		ID: "agent-1", ProjectID: project.ID(), Tool: domain.AgentOpenCode,
		Status: domain.AgentRunning, CreatedAt: now, StartedAt: now,
	})
	if err != nil {
		t.Fatalf("NewAgentSession() error = %v", err)
	}
	name, err := domain.NewName("analytics")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	connection, err := domain.NewDatabaseConnection(domain.DatabaseValues{
		ID: "db-1", Name: name, Driver: domain.DatabaseSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "analytics.db"), CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("NewDatabaseConnection() error = %v", err)
	}

	// When
	if err := store.SaveAgentSession(ctx, session); err != nil {
		t.Fatalf("SaveAgentSession() error = %v", err)
	}
	if err := store.CreateDatabaseConnection(ctx, connection); err != nil {
		t.Fatalf("CreateDatabaseConnection() error = %v", err)
	}
	storedSession, err := store.GetAgentSession(ctx, session.ID())
	if err != nil {
		t.Fatalf("GetAgentSession() error = %v", err)
	}
	connections, err := store.ListDatabaseConnections(ctx)
	// Then
	if err != nil {
		t.Fatalf("ListDatabaseConnections() error = %v", err)
	}
	if storedSession.Status() != domain.AgentRunning || len(connections) != 1 {
		t.Fatalf("stored agent/databases = %q/%d", storedSession.Status(), len(connections))
	}
}
