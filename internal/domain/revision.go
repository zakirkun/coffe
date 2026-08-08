package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidGitRevision = errors.New("domain: invalid git revision")
	ErrInvalidStackName   = errors.New("domain: invalid stack name")
)

var (
	revisionPattern = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
	stackPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,62}$`)
)

type GitRevision struct{ raw string }

func NewGitRevision(raw string) (GitRevision, error) {
	clean := strings.ToLower(strings.TrimSpace(raw))
	if !revisionPattern.MatchString(clean) {
		return GitRevision{}, fmt.Errorf("revision %q: %w", raw, ErrInvalidGitRevision)
	}
	return GitRevision{raw: clean}, nil
}

func (r GitRevision) String() string { return r.raw }
func (r GitRevision) IsZero() bool   { return r.raw == "" }

type StackName struct{ raw string }

func NewStackName(raw string) (StackName, error) {
	clean := strings.TrimSpace(raw)
	if !stackPattern.MatchString(clean) {
		return StackName{}, fmt.Errorf("stack name %q: %w", raw, ErrInvalidStackName)
	}
	return StackName{raw: clean}, nil
}

func (n StackName) String() string { return n.raw }
