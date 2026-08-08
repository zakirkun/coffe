package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidGitOperation = errors.New("domain: invalid git operation")
	ErrGitTransition       = errors.New("domain: invalid git transition")
	ErrInvalidGitRemote    = errors.New("domain: invalid git remote")
)

var gitRemotePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type (
	GitOperationID     string
	GitOperationKind   string
	GitOperationStatus string
)

const (
	GitPull GitOperationKind = "pull"
	GitPush GitOperationKind = "push"
)

const (
	GitDraft     GitOperationStatus = "draft"
	GitApproved  GitOperationStatus = "approved"
	GitRunning   GitOperationStatus = "running"
	GitSucceeded GitOperationStatus = "succeeded"
	GitFailed    GitOperationStatus = "failed"
)

func (k GitOperationKind) Valid() bool { return k == GitPull || k == GitPush }

func (s GitOperationStatus) Valid() bool {
	switch s {
	case GitDraft, GitApproved, GitRunning, GitSucceeded, GitFailed:
		return true
	default:
		return false
	}
}

type GitRemote struct{ raw string }

func NewGitRemote(raw string) (GitRemote, error) {
	clean := strings.TrimSpace(raw)
	if !gitRemotePattern.MatchString(clean) {
		return GitRemote{}, fmt.Errorf("git remote %q: %w", raw, ErrInvalidGitRemote)
	}
	return GitRemote{raw: clean}, nil
}

func (r GitRemote) String() string { return r.raw }

type GitOperation struct {
	createdAt  time.Time
	approvedAt time.Time
	id         GitOperationID
	projectID  ProjectID
	kind       GitOperationKind
	remote     GitRemote
	branch     Branch
	status     GitOperationStatus
}

type GitOperationValues struct {
	ID         GitOperationID
	ProjectID  ProjectID
	Kind       GitOperationKind
	Remote     GitRemote
	Branch     Branch
	Status     GitOperationStatus
	CreatedAt  time.Time
	ApprovedAt time.Time
}

func NewGitOperation(values GitOperationValues) (GitOperation, error) {
	if values.ID == "" || values.ProjectID == "" || !values.Kind.Valid() ||
		values.Remote.String() == "" || values.Branch.String() == "" ||
		!values.Status.Valid() || values.CreatedAt.IsZero() {
		return GitOperation{}, fmt.Errorf("git operation: %w", ErrInvalidGitOperation)
	}
	if values.Status != GitDraft && values.ApprovedAt.IsZero() {
		return GitOperation{}, fmt.Errorf("git approval time: %w", ErrInvalidGitOperation)
	}
	return GitOperation{
		id: values.ID, projectID: values.ProjectID, kind: values.Kind,
		remote: values.Remote, branch: values.Branch, status: values.Status,
		createdAt: values.CreatedAt.UTC(), approvedAt: values.ApprovedAt.UTC(),
	}, nil
}

func (o GitOperation) Approve(at time.Time) (GitOperation, error) {
	if o.status != GitDraft || at.IsZero() {
		return GitOperation{}, fmt.Errorf("approve git %s: %w", o.status, ErrGitTransition)
	}
	o.status = GitApproved
	o.approvedAt = at.UTC()
	return o, nil
}

func (o GitOperation) Start() (GitOperation, error) {
	if o.status != GitApproved {
		return GitOperation{}, fmt.Errorf("start git %s: %w", o.status, ErrGitTransition)
	}
	o.status = GitRunning
	return o, nil
}

func (o GitOperation) Finish(succeeded bool) (GitOperation, error) {
	if o.status != GitRunning {
		return GitOperation{}, fmt.Errorf("finish git %s: %w", o.status, ErrGitTransition)
	}
	if succeeded {
		o.status = GitSucceeded
	} else {
		o.status = GitFailed
	}
	return o, nil
}

func (o GitOperation) ID() GitOperationID         { return o.id }
func (o GitOperation) ProjectID() ProjectID       { return o.projectID }
func (o GitOperation) Kind() GitOperationKind     { return o.kind }
func (o GitOperation) Remote() GitRemote          { return o.remote }
func (o GitOperation) Branch() Branch             { return o.branch }
func (o GitOperation) Status() GitOperationStatus { return o.status }
func (o GitOperation) CreatedAt() time.Time       { return o.createdAt }
func (o GitOperation) ApprovedAt() (time.Time, bool) {
	return o.approvedAt, !o.approvedAt.IsZero()
}

type GitStatus struct {
	Branch    string `json:"branch"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Staged    int    `json:"staged"`
	Modified  int    `json:"modified"`
	Untracked int    `json:"untracked"`
}

func (s GitStatus) Clean() bool { return s.Staged+s.Modified+s.Untracked == 0 }

type GitEvent struct {
	occurredAt  time.Time
	operationID GitOperationID
	message     string
}

func NewGitEvent(operationID GitOperationID, message string, occurredAt time.Time) (GitEvent, error) {
	if operationID == "" || message == "" || occurredAt.IsZero() || hasControl(message) {
		return GitEvent{}, fmt.Errorf("git event: %w", ErrInvalidGitOperation)
	}
	return GitEvent{operationID: operationID, message: message, occurredAt: occurredAt.UTC()}, nil
}

func (e GitEvent) OperationID() GitOperationID { return e.operationID }
func (e GitEvent) Message() string             { return e.message }
func (e GitEvent) OccurredAt() time.Time       { return e.occurredAt }
