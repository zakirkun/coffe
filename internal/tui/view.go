package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	if m.width < minWidth || m.height < minHeight {
		view := tea.NewView(fmt.Sprintf(
			"Coffe needs at least 80x24 cells. Current terminal: %dx%d.", m.width, m.height,
		))
		view.AltScreen = true
		view.WindowTitle = "Coffe"
		return view
	}
	content := m.renderDashboard()
	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = "Coffe"
	if m.focus == focusPrompt {
		if rawCursor := m.prompt.Cursor(); rawCursor != nil {
			cursor := *rawCursor
			offsetX, offsetY := m.promptOffset()
			cursor.X += offsetX
			cursor.Y += offsetY
			view.Cursor = &cursor
		}
	}
	return view
}

func (m Model) renderDashboard() string {
	if m.width < 100 {
		paneHeight := m.height / 4
		lastPaneHeight := m.height - paneHeight*3
		return lipgloss.JoinVertical(lipgloss.Left,
			m.styles.pane("Projects / Targets", m.renderRegistry(), m.width, paneHeight, m.focus < focusPrompt),
			m.styles.pane("Prompt", m.renderPrompt(), m.width, paneHeight, m.focus == focusPrompt),
			m.styles.pane("Jobs", m.renderJobs(), m.width, paneHeight, m.focus == focusJobs),
			m.styles.pane("Live Logs", m.renderEvents(), m.width, lastPaneHeight, false),
		)
	}
	leftWidth := (m.width - 1) * 2 / 5
	rightWidth := m.width - leftWidth - 1
	topHeight := (m.height - 1) / 2
	bottomHeight := m.height - topHeight - 1
	top := lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.pane("Projects / Targets", m.renderRegistry(), leftWidth, topHeight, m.focus < focusPrompt),
		" ",
		m.styles.pane("Prompt", m.renderPrompt(), rightWidth, topHeight, m.focus == focusPrompt),
	)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.pane("Jobs", m.renderJobs(), leftWidth, bottomHeight, m.focus == focusJobs),
		" ",
		m.styles.pane("Live Logs", m.renderEvents(), rightWidth, bottomHeight, false),
	)
	footer := m.styles.mutedRow.Render("Tab focus  Enter run  a approve  x execute  c cancel  y retry  b rollback  s swarm  r refresh  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom, footer)
}

func (m Model) renderRegistry() string {
	var builder strings.Builder
	if len(m.projects) == 0 {
		return m.styles.mutedRow.Render("No projects registered.")
	}
	for index, project := range m.projects {
		row := "  " + project.Name().String()
		if index == m.projectIndex {
			row = m.styles.selected.Render("> " + project.Name().String())
		}
		builder.WriteString(row + "\n")
	}
	builder.WriteString(m.styles.mutedRow.Render("Targets") + "\n")
	if len(m.targets) == 0 {
		builder.WriteString(m.styles.mutedRow.Render("  No targets"))
		return builder.String()
	}
	for index, target := range m.targets {
		row := "  " + target.Name().String() + "  " + target.Host().String()
		if index == m.targetIndex {
			row = m.styles.selected.Render("> " + target.Name().String() + "  " + target.Host().String())
		}
		builder.WriteString(row + "\n")
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func (m Model) renderPrompt() string {
	selection := "Select a project and target"
	if project, ok := m.selectedProject(); ok {
		selection = project.Name().String()
		if target, targetOK := m.selectedTarget(); targetOK {
			selection += " -> " + target.Name().String()
		}
	}
	return m.styles.mutedRow.Render(selection) + "\n" + m.prompt.View() + "\n" + m.status
}

func (m Model) renderEvents() string {
	if len(m.events) == 0 {
		return m.styles.mutedRow.Render("No events for the selected plan.")
	}
	start := max(0, len(m.events)-10)
	var builder strings.Builder
	for _, event := range m.events[start:] {
		builder.WriteString(event.OccurredAt().Format("15:04:05") + "  " + event.Message() + "\n")
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func (m Model) promptOffset() (int, int) {
	if m.width < 100 {
		paneHeight := m.height / 4
		return 2, paneHeight + 3
	}
	leftWidth := (m.width - 1) * 2 / 5
	return leftWidth + 3, 3
}
