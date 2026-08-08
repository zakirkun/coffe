package gittool_test

import (
	"coffe/internal/domain"
	"coffe/internal/gittool"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func Test_Tool_reads_real_repository_status_and_diff(t *testing.T) {
	// Given
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	runGit(t, root, "config", "user.email", "coffe@example.invalid")
	runGit(t, root, "config", "user.name", "Coffe Test")
	path := filepath.Join(root, "app.txt")
	if err := os.WriteFile(path, []byte("first\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	runGit(t, root, "add", "app.txt")
	runGit(t, root, "commit", "-m", "initial")
	if err := os.WriteFile(path, []byte("second\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	project := gitProject(t, root)

	// When
	status, err := (gittool.Tool{}).Status(context.Background(), project)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	diff, err := (gittool.Tool{}).Diff(context.Background(), project)
	// Then
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if status.Branch != "main" || status.Modified != 1 || diff == "" {
		t.Fatalf("status/diff = %#v/%q", status, diff)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil { //nolint:gosec // test invokes the fixed git binary with test-owned arguments.
		t.Fatalf("git %v error = %v: %s", args, err, output)
	}
}

func gitProject(t *testing.T, root string) domain.Project {
	t.Helper()
	name, err := domain.NewName("git-test")
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
