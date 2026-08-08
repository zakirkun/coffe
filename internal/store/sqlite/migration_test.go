package sqlite_test

import (
	"coffe/internal/domain"
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	coffesqlite "coffe/internal/store/sqlite"

	_ "modernc.org/sqlite"
)

func Test_Open_migrates_legacy_v1_plan_columns(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "legacy.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	legacy := `
		CREATE TABLE projects (id TEXT PRIMARY KEY, name TEXT NOT NULL, root_path TEXT NOT NULL, created_at TEXT NOT NULL);
		CREATE TABLE targets (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, name TEXT NOT NULL, host TEXT NOT NULL,
			port INTEGER NOT NULL, ssh_user TEXT NOT NULL, remote_path TEXT NOT NULL, branch TEXT NOT NULL, created_at TEXT NOT NULL);
		CREATE TABLE plans (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, target_id TEXT NOT NULL, status TEXT NOT NULL,
			steps_json TEXT NOT NULL, created_at TEXT NOT NULL, approved_at TEXT);
		INSERT INTO projects VALUES ('project-1', 'legacy', 'C:\legacy', '2026-08-07T10:00:00Z');
		INSERT INTO targets VALUES ('target-1', 'project-1', 'office', 'server.example.com', 22, 'deploy', '/srv/app', 'main', '2026-08-07T10:00:00Z');
		INSERT INTO plans VALUES ('plan-1', 'project-1', 'target-1', 'draft', '["inspect_repository"]', '2026-08-07T10:00:00Z', NULL);`
	if _, err := database.Exec(legacy); err != nil {
		t.Fatalf("Exec(legacy) error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close(legacy) error = %v", err)
	}

	// When
	store, err := coffesqlite.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})
	plan, err := store.GetPlan(context.Background(), domain.PlanID("plan-1"))
	// Then
	if err != nil {
		t.Fatalf("GetPlan() error = %v", err)
	}
	if plan.Kind() != domain.PlanComposeDeploy {
		t.Fatalf("migrated plan kind = %q", plan.Kind())
	}
}
