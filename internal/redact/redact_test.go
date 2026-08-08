package redact_test

import (
	"coffe/internal/redact"
	"strings"
	"testing"
)

func Test_Line_removes_credentials_and_terminal_controls(t *testing.T) {
	// Given
	raw := "\x1b[31mconnecting mysql://admin:hunter2@db token=abc123\x1b[0m"

	// When
	clean := redact.Line(raw)

	// Then
	for _, secret := range []string{"hunter2", "abc123", "\x1b"} {
		if strings.Contains(clean, secret) {
			t.Fatalf("Line() retained sensitive content %q in %q", secret, clean)
		}
	}
}
