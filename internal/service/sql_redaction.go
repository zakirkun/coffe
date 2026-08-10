package service

import "strings"

func redactSQLLiterals(statement string) string {
	var redacted strings.Builder
	redacted.Grow(len(statement))
	for index := 0; index < len(statement); {
		switch {
		case statement[index] == '\'' || statement[index] == '"':
			redacted.WriteByte('?')
			index = skipSQLQuoted(statement, index, statement[index])
		case statement[index] == '`':
			next := skipSQLQuoted(statement, index, '`')
			redacted.WriteString(statement[index:next])
			index = next
		case statement[index] == '[':
			next := strings.IndexByte(statement[index+1:], ']')
			if next < 0 {
				redacted.WriteString(statement[index:])
				return redacted.String()
			}
			next += index + 2
			redacted.WriteString(statement[index:next])
			index = next
		case startsSQLNumber(statement, index):
			redacted.WriteByte('?')
			index = skipSQLNumber(statement, index)
		default:
			redacted.WriteByte(statement[index])
			index++
		}
	}
	return redacted.String()
}

func skipSQLQuoted(statement string, start int, quote byte) int {
	for index := start + 1; index < len(statement); index++ {
		if statement[index] == '\\' {
			index++
			continue
		}
		if statement[index] != quote {
			continue
		}
		if index+1 < len(statement) && statement[index+1] == quote {
			index++
			continue
		}
		return index + 1
	}
	return len(statement)
}

func startsSQLNumber(statement string, index int) bool {
	startsWithDigit := isSQLDigit(statement[index])
	startsWithDot := statement[index] == '.' && index+1 < len(statement) && isSQLDigit(statement[index+1])
	if !startsWithDigit && !startsWithDot {
		return false
	}
	if index == 0 {
		return true
	}
	previous := statement[index-1]
	return !isSQLIdentifierByte(previous) && previous != '?' && previous != '$'
}

func skipSQLNumber(statement string, start int) int {
	index := start
	if index+2 < len(statement) && statement[index] == '0' &&
		(statement[index+1] == 'b' || statement[index+1] == 'B') && isSQLBinary(statement[index+2]) {
		index += 2
		for index < len(statement) && isSQLBinary(statement[index]) {
			index++
		}
		return index
	}
	if index+1 < len(statement) && statement[index] == '0' &&
		(statement[index+1] == 'x' || statement[index+1] == 'X') {
		index += 2
		for index < len(statement) && (isSQLHex(statement[index]) || statement[index] == '_') {
			index++
		}
		return index
	}
	for index < len(statement) && (isSQLDigit(statement[index]) || statement[index] == '_') {
		index++
	}
	if index < len(statement) && statement[index] == '.' {
		index++
		for index < len(statement) && (isSQLDigit(statement[index]) || statement[index] == '_') {
			index++
		}
	}
	if index < len(statement) && (statement[index] == 'e' || statement[index] == 'E') {
		index++
		if index < len(statement) && (statement[index] == '+' || statement[index] == '-') {
			index++
		}
		for index < len(statement) && (isSQLDigit(statement[index]) || statement[index] == '_') {
			index++
		}
	}
	return index
}

func isSQLIdentifierByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' ||
		isSQLDigit(value) || value == '_' || value == '$'
}

func isSQLDigit(value byte) bool { return value >= '0' && value <= '9' }

func isSQLBinary(value byte) bool { return value == '0' || value == '1' }

func isSQLHex(value byte) bool {
	return isSQLDigit(value) || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}
