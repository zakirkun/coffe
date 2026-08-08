package tui

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type styles struct {
	text     color.Color
	muted    color.Color
	border   color.Color
	focus    color.Color
	success  color.Color
	error    color.Color
	title    lipgloss.Style
	mutedRow lipgloss.Style
	selected lipgloss.Style
	statusOK lipgloss.Style
	statusNG lipgloss.Style
}

func newStyles() styles {
	text := lipgloss.Color("#ECE7DE")
	muted := lipgloss.Color("#8E9399")
	border := lipgloss.Color("#30343A")
	focus := lipgloss.Color("#D99A45")
	success := lipgloss.Color("#55B983")
	errorColor := lipgloss.Color("#E06C65")
	return styles{
		text: text, muted: muted, border: border, focus: focus, success: success, error: errorColor,
		title:    lipgloss.NewStyle().Bold(true).Foreground(text),
		mutedRow: lipgloss.NewStyle().Foreground(muted),
		selected: lipgloss.NewStyle().Foreground(focus).Bold(true),
		statusOK: lipgloss.NewStyle().Foreground(success),
		statusNG: lipgloss.NewStyle().Foreground(errorColor),
	}
}

func promptStyles() textarea.Styles {
	styles := textarea.DefaultDarkStyles()
	text := lipgloss.Color("#ECE7DE")
	muted := lipgloss.Color("#8E9399")
	focus := lipgloss.Color("#D99A45")
	styles.Focused.Text = styles.Focused.Text.Foreground(text)
	styles.Focused.Placeholder = styles.Focused.Placeholder.Foreground(muted)
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(focus)
	styles.Blurred.Text = styles.Blurred.Text.Foreground(text)
	styles.Blurred.Placeholder = styles.Blurred.Placeholder.Foreground(muted)
	styles.Cursor.Color = focus
	styles.Cursor.Shape = tea.CursorBar
	styles.Cursor.Blink = true
	return styles
}

func (s styles) pane(title, body string, width, height int, focused bool) string {
	borderColor := s.border
	if focused {
		borderColor = s.focus
	}
	heading := s.title.Render(title)
	body = fitPaneBody(body, max(1, width-4), max(1, height-3))
	return lipgloss.NewStyle().
		Width(max(1, width)).
		Height(max(1, height)).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Foreground(s.text).
		Render(heading + "\n" + body)
}

func fitPaneBody(body string, width, height int) string {
	lines := strings.Split(body, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for index, line := range lines {
		lines[index] = ansi.Truncate(line, width, "...")
	}
	return strings.Join(lines, "\n")
}
