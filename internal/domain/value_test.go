package domain_test

import (
	"coffe/internal/domain"
	"errors"
	"testing"
)

func Test_NewName_trims_valid_input(t *testing.T) {
	// Given
	raw := "  payments-api  "

	// When
	name, err := domain.NewName(raw)
	// Then
	if err != nil {
		t.Fatalf("NewName() error = %v", err)
	}
	if name.String() != "payments-api" {
		t.Fatalf("Name.String() = %q, want %q", name.String(), "payments-api")
	}
}

func Test_NewName_rejects_empty_input(t *testing.T) {
	// Given
	raw := "   "

	// When
	_, err := domain.NewName(raw)

	// Then
	if !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("NewName() error = %v, want ErrInvalidName", err)
	}
}

func Test_NewRemotePath_requires_absolute_posix_path(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "absolute path", raw: "/srv/payments", wantErr: false},
		{name: "relative path", raw: "srv/payments", wantErr: true},
		{name: "root path", raw: "/", wantErr: true},
		{name: "newline", raw: "/srv/app\nwhoami", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			_, err := domain.NewRemotePath(tt.raw)

			// Then
			if tt.wantErr && !errors.Is(err, domain.ErrInvalidRemotePath) {
				t.Fatalf("NewRemotePath() error = %v, want ErrInvalidRemotePath", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("NewRemotePath() error = %v", err)
			}
		})
	}
}

func Test_NewBranch_rejects_shell_metacharacters(t *testing.T) {
	// Given
	raw := "main; rm -rf /"

	// When
	_, err := domain.NewBranch(raw)

	// Then
	if !errors.Is(err, domain.ErrInvalidBranch) {
		t.Fatalf("NewBranch() error = %v, want ErrInvalidBranch", err)
	}
}

func Test_NewComposeFile_rejects_path_traversal(t *testing.T) {
	// Given
	raw := "../docker-compose.yml"

	// When
	_, err := domain.NewComposeFile(raw)

	// Then
	if !errors.Is(err, domain.ErrInvalidComposeFile) {
		t.Fatalf("NewComposeFile() error = %v, want ErrInvalidComposeFile", err)
	}
}
