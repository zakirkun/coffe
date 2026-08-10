package service

import "testing"

func Test_redactSQLLiterals_removes_values_and_preserves_structure(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		want      string
	}{
		{name: "quoted strings", statement: "SELECT 'secret', 'it''s'", want: "SELECT ?, ?"},
		{name: "double quoted strings", statement: `SELECT "secret" FROM users`, want: "SELECT ? FROM users"},
		{name: "backslash escape", statement: `SELECT 'slash\'secret' FROM users`, want: "SELECT ? FROM users"},
		{name: "numbers", statement: "SELECT 12, -3.5e+2, 0xCAFE", want: "SELECT ?, -?, ?"},
		{name: "MySQL binary numbers", statement: "SELECT 0b1010, 0B0101, col0b1 FROM bits_0b10", want: "SELECT ?, ?, col0b1 FROM bits_0b10"},
		{name: "quoted identifiers", statement: "SELECT `table2`, [field3] FROM logs_2026 WHERE id = 7", want: "SELECT `table2`, [field3] FROM logs_2026 WHERE id = ?"},
		{name: "unterminated string", statement: "SELECT 'secret", want: "SELECT ?"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			got := redactSQLLiterals(test.statement)

			// Then
			if got != test.want {
				t.Fatalf("redactSQLLiterals() = %q, want %q", got, test.want)
			}
		})
	}
}
