package config_test

import (
	"coffe/internal/config"
	"coffe/internal/domain"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_WorkflowLoader_Load_parses_repository_config(t *testing.T) {
	// Given
	root := t.TempDir()
	contents := []byte("version: 1\ndeploy:\n  compose_file: deploy/compose.yml\n  build: true\n")
	if err := os.WriteFile(filepath.Join(root, "coffe.yaml"), contents, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	project := newProject(t, root)

	// When
	workflow, err := (config.WorkflowLoader{}).Load(context.Background(), project)
	// Then
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if workflow.ComposeFile().String() != "deploy/compose.yml" {
		t.Fatalf("ComposeFile() = %q, want deploy/compose.yml", workflow.ComposeFile().String())
	}
	if !workflow.Build() {
		t.Fatal("Build() = false, want true")
	}
}

func Test_WorkflowLoader_Load_rejects_unknown_fields(t *testing.T) {
	// Given
	root := t.TempDir()
	contents := []byte("version: 1\ndeploy:\n  compose_file: compose.yml\n  build: false\n  command: whoami\n")
	if err := os.WriteFile(filepath.Join(root, "coffe.yaml"), contents, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// When
	_, err := (config.WorkflowLoader{}).Load(context.Background(), newProject(t, root))

	// Then
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
}

func Test_WorkflowLoader_Load_parses_swarm_stack(t *testing.T) {
	// Given
	root := t.TempDir()
	contents := []byte("version: 1\ndeploy:\n  mode: swarm\n  compose_file: stack.yml\n  stack_name: payments\n")
	if err := os.WriteFile(filepath.Join(root, "coffe.yaml"), contents, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// When
	workflow, err := (config.WorkflowLoader{}).Load(context.Background(), newProject(t, root))
	// Then
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if workflow.Mode() != domain.DeploymentSwarm || workflow.StackName().String() != "payments" {
		t.Fatalf("workflow mode/stack = %q/%q", workflow.Mode(), workflow.StackName().String())
	}
}

func newProject(t *testing.T, root string) domain.Project {
	t.Helper()
	name, err := domain.NewName("payments")
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	project, err := domain.NewProject(domain.ProjectValues{
		ID: "project-1", Name: name, RootPath: root,
		CreatedAt: time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewProject() error = %v", err)
	}
	return project
}
