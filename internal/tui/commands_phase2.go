package tui

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type (
	agentMsg struct{ session domain.AgentSession }
	queryMsg struct{ result domain.QueryResult }
	swarmMsg struct {
		nodes    []domain.SwarmNode
		services []domain.SwarmService
	}
)
type statusMsg string

func (m Model) submitPrompt() tea.Cmd {
	prompt := strings.TrimSpace(m.prompt.Value())
	lower := strings.ToLower(prompt)
	if strings.HasPrefix(lower, "git ") {
		return m.gitPrompt(prompt)
	}
	if strings.HasPrefix(lower, "agent ") {
		return m.startAgentPrompt(prompt)
	}
	if strings.HasPrefix(lower, "sql ") {
		return m.queryDatabasePrompt(strings.TrimSpace(prompt[4:]))
	}
	return m.previewDeploy()
}

func (m Model) startAgentPrompt(prompt string) tea.Cmd {
	project, ok := m.selectedProject()
	parts := strings.SplitN(prompt, " ", 3)
	return func() tea.Msg {
		if !ok || len(parts) != 3 {
			return errorMsg{err: fmt.Errorf("use agent TOOL PROMPT")}
		}
		session, err := m.app.StartAgent(m.ctx, service.StartAgentInput{
			ProjectID: string(project.ID()), Tool: strings.ToLower(parts[1]), Prompt: parts[2],
		})
		if err != nil {
			return errorMsg{err: err}
		}
		return agentMsg{session: session}
	}
}

func (m Model) queryDatabasePrompt(query string) tea.Cmd {
	return func() tea.Msg {
		if len(m.databases) == 0 || query == "" {
			return errorMsg{err: fmt.Errorf("register a database and use sql QUERY")}
		}
		result, err := m.app.QueryDatabase(m.ctx, string(m.databases[0].ID()), query)
		if err != nil {
			return errorMsg{err: err}
		}
		return queryMsg{result: result}
	}
}

func (m Model) cancelPlan() tea.Cmd {
	plan, ok := m.selectedPlan()
	return func() tea.Msg {
		if !ok {
			return errorMsg{err: errSelectionRequired}
		}
		if err := m.app.CancelPlan(m.ctx, string(plan.ID())); err != nil {
			return errorMsg{err: err}
		}
		return statusMsg("Cancellation requested")
	}
}

func (m Model) retryPlan() tea.Cmd {
	plan, ok := m.selectedPlan()
	return func() tea.Msg {
		if !ok {
			return errorMsg{err: errSelectionRequired}
		}
		retry, err := m.app.RetryPlan(m.ctx, string(plan.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		preview, err := m.app.PlanPreview(m.ctx, string(retry.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		return previewMsg{preview: preview}
	}
}

func (m Model) rollbackPlan() tea.Cmd {
	target, ok := m.selectedTarget()
	return func() tea.Msg {
		if !ok {
			return errorMsg{err: errSelectionRequired}
		}
		plan, err := m.app.PreviewRollback(m.ctx, string(target.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		preview, err := m.app.PlanPreview(m.ctx, string(plan.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		return previewMsg{preview: preview}
	}
}

func (m Model) loadSwarm() tea.Cmd {
	target, ok := m.selectedTarget()
	return func() tea.Msg {
		if !ok {
			return errorMsg{err: errSelectionRequired}
		}
		nodes, err := m.app.ListSwarmNodes(m.ctx, string(target.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		services, err := m.app.ListSwarmServices(m.ctx, string(target.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		return swarmMsg{nodes: nodes, services: services}
	}
}
