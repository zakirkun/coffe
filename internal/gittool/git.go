package gittool

import (
	"bytes"
	"coffe/internal/domain"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var ErrInvalidStatus = errors.New("gittool: invalid status output")

const maxOutputBytes = 256 * 1024

type Tool struct{}

func (Tool) Status(ctx context.Context, project domain.Project) (domain.GitStatus, error) {
	output, err := run(ctx, project.RootPath(), "status", "--porcelain=v2", "--branch")
	if err != nil {
		return domain.GitStatus{}, fmt.Errorf("git status: %w", err)
	}
	return ParseStatus(string(output))
}

func (Tool) Branches(ctx context.Context, project domain.Project) ([]string, error) {
	output, err := run(ctx, project.RootPath(), "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("git branches: %w", err)
	}
	var branches []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		branch := strings.TrimSpace(line)
		if branch != "" {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

func (Tool) Diff(ctx context.Context, project domain.Project) (string, error) {
	output, err := run(ctx, project.RootPath(), "diff", "--no-ext-diff", "--unified=3")
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return string(output), nil
}

func (Tool) Render(operation domain.GitOperation) string {
	if operation.Kind() == domain.GitPull {
		return "git pull --ff-only " + operation.Remote().String() + " " + operation.Branch().String()
	}
	return "git push " + operation.Remote().String() + " " + operation.Branch().String()
}

func (Tool) Execute(ctx context.Context, project domain.Project, operation domain.GitOperation) (string, error) {
	var args []string
	if operation.Kind() == domain.GitPull {
		args = []string{"pull", "--ff-only", operation.Remote().String(), operation.Branch().String()}
	} else {
		args = []string{"push", operation.Remote().String(), operation.Branch().String()}
	}
	output, err := run(ctx, project.RootPath(), args...)
	if err != nil {
		return string(output), fmt.Errorf("git %s: %w", operation.Kind(), err)
	}
	return string(output), nil
}

func run(ctx context.Context, root string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-C", root}, args...)
	output := &cappedWriter{limit: maxOutputBytes}
	command := exec.CommandContext(ctx, "git", commandArgs...) //nolint:gosec // git is fixed and mutation arguments are parsed domain values.
	command.Stdout = output
	command.Stderr = output
	if err := command.Run(); err != nil {
		return output.Bytes(), fmt.Errorf("git process: %w", err)
	}
	return output.Bytes(), nil
}

type cappedWriter struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	limit  int
}

func (w *cappedWriter) Write(content []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	length := len(content)
	remaining := w.limit - w.buffer.Len()
	if remaining <= 0 {
		return length, nil
	}
	if len(content) > remaining {
		content = content[:remaining]
	}
	if _, err := w.buffer.Write(content); err != nil {
		return 0, err
	}
	return length, nil
}

func (w *cappedWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return bytes.Clone(w.buffer.Bytes())
}

func ParseStatus(raw string) (domain.GitStatus, error) {
	var status domain.GitStatus
	for line := range strings.SplitSeq(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			fields := strings.Fields(line)
			if len(fields) != 4 {
				return domain.GitStatus{}, ErrInvalidStatus
			}
			ahead, err := strconv.Atoi(strings.TrimPrefix(fields[2], "+"))
			if err != nil {
				return domain.GitStatus{}, fmt.Errorf("parse ahead: %w", err)
			}
			behind, err := strconv.Atoi(strings.TrimPrefix(fields[3], "-"))
			if err != nil {
				return domain.GitStatus{}, fmt.Errorf("parse behind: %w", err)
			}
			status.Ahead, status.Behind = ahead, behind
		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			fields := strings.Fields(line)
			if len(fields) < 2 || len(fields[1]) != 2 {
				return domain.GitStatus{}, ErrInvalidStatus
			}
			if fields[1][0] != '.' {
				status.Staged++
			}
			if fields[1][1] != '.' {
				status.Modified++
			}
		case strings.HasPrefix(line, "? "):
			status.Untracked++
		}
	}
	if status.Branch == "" {
		return domain.GitStatus{}, ErrInvalidStatus
	}
	return status, nil
}
