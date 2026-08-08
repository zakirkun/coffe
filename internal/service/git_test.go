package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"testing"
)

func Test_App_GitOperation_requires_preview_and_approval(t *testing.T) {
	// Given
	fixture := newFixture(t)
	preview, err := fixture.app.PreviewGit(context.Background(), service.CreateGitOperationInput{
		ProjectID: string(fixture.project.ID()), Kind: string(domain.GitPull),
		Remote: "origin", Branch: "main",
	})
	if err != nil {
		t.Fatalf("PreviewGit() error = %v", err)
	}
	approved, err := fixture.app.ApproveGit(context.Background(), string(preview.Operation.ID()))
	if err != nil {
		t.Fatalf("ApproveGit() error = %v", err)
	}

	// When
	finished, err := fixture.app.ExecuteGit(context.Background(), string(approved.ID()))
	// Then
	if err != nil {
		t.Fatalf("ExecuteGit() error = %v", err)
	}
	if finished.Status() != domain.GitSucceeded || preview.Command != "git pull" {
		t.Fatalf("git status/command = %q/%q", finished.Status(), preview.Command)
	}
}
