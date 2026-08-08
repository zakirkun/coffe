package domain_test

import (
	"coffe/internal/domain"
	"errors"
	"testing"
)

func Test_NewGitRevision_accepts_SHA1_and_rejects_shell_input(t *testing.T) {
	// Given
	valid := "0123456789abcdef0123456789abcdef01234567"
	invalid := valid + ";whoami"

	// When
	revision, validErr := domain.NewGitRevision(valid)
	_, invalidErr := domain.NewGitRevision(invalid)

	// Then
	if validErr != nil || revision.String() != valid {
		t.Fatalf("NewGitRevision(valid) = %q, %v", revision.String(), validErr)
	}
	if !errors.Is(invalidErr, domain.ErrInvalidGitRevision) {
		t.Fatalf("NewGitRevision(invalid) error = %v", invalidErr)
	}
}

func Test_NewStackName_rejects_invalid_docker_stack_name(t *testing.T) {
	// Given
	raw := "prod stack;whoami"

	// When
	_, err := domain.NewStackName(raw)

	// Then
	if !errors.Is(err, domain.ErrInvalidStackName) {
		t.Fatalf("NewStackName() error = %v", err)
	}
}
