package redact

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

const maxLineRunes = 500

var (
	assignmentPattern = regexp.MustCompile(`(?i)\b(password|passwd|token|api[_-]?key|secret)\s*[:=]\s*[^\s]+`)
	urlUserPattern    = regexp.MustCompile(`([A-Za-z][A-Za-z0-9+.-]*://)[^/@\s]+@`)
)

func Line(raw string) string {
	clean := ansi.Strip(raw)
	clean = urlUserPattern.ReplaceAllString(clean, `${1}[REDACTED]@`)
	clean = assignmentPattern.ReplaceAllStringFunc(clean, func(match string) string {
		separator := strings.IndexAny(match, ":=")
		if separator < 0 {
			return "[REDACTED]"
		}
		return strings.TrimSpace(match[:separator]) + "=[REDACTED]"
	})
	clean = strings.Map(func(value rune) rune {
		if unicode.IsControl(value) && value != '\t' {
			return ' '
		}
		return value
	}, clean)
	clean = strings.TrimSpace(clean)
	runes := []rune(clean)
	if len(runes) > maxLineRunes {
		clean = string(runes[:maxLineRunes])
	}
	return clean
}
