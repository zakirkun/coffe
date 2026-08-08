package domain_test

import (
	"coffe/internal/domain"
	"testing"
	"time"
)

func Test_GitOperation_requires_approval_before_start(t *testing.T) {
	// Given
	createdAt := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	remote, err := domain.NewGitRemote("origin")
	if err != nil {
		t.Fatalf("NewGitRemote() error = %v", err)
	}
	branch, err := domain.NewBranch("main")
	if err != nil {
		t.Fatalf("NewBranch() error = %v", err)
	}
	operation, err := domain.NewGitOperation(domain.GitOperationValues{
		ID: "git-1", ProjectID: "project-1", Kind: domain.GitPull,
		Remote: remote, Branch: branch, Status: domain.GitDraft, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("NewGitOperation() error = %v", err)
	}

	// When
	approved, err := operation.Approve(createdAt.Add(time.Second))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	running, err := approved.Start()
	// Then
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if running.Status() != domain.GitRunning {
		t.Fatalf("Start() status = %q, want running", running.Status())
	}
}
