package gittool_test

import (
	"coffe/internal/gittool"
	"testing"
)

func Test_ParseStatus_reads_branch_counts_and_worktree_state(t *testing.T) {
	// Given
	raw := "# branch.head main\n# branch.ab +2 -1\n1 M. N... 100644 100644 100644 abc abc staged.go\n1 .M N... 100644 100644 100644 abc abc changed.go\n? new.go\n"

	// When
	status, err := gittool.ParseStatus(raw)
	// Then
	if err != nil {
		t.Fatalf("ParseStatus() error = %v", err)
	}
	if status.Branch != "main" || status.Ahead != 2 || status.Behind != 1 {
		t.Fatalf("branch/ahead/behind = %q/%d/%d", status.Branch, status.Ahead, status.Behind)
	}
	if status.Staged != 1 || status.Modified != 1 || status.Untracked != 1 {
		t.Fatalf("staged/modified/untracked = %d/%d/%d", status.Staged, status.Modified, status.Untracked)
	}
}
