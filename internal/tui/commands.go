package tui

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"time"

	tea "charm.land/bubbletea/v2"
)

type dashboardMsg struct {
	projects      []domain.Project
	targets       []domain.Target
	plans         []domain.ActionPlan
	events        []domain.Event
	agents        []domain.AgentSession
	databases     []domain.DatabaseConnection
	gitOperations []domain.GitOperation
	providers     []domain.AIProvider
	aiUsage       []domain.AIUsage
	policy        domain.Policy
	audit         []domain.AuditEntry
	workers       []domain.Worker
	plugins       []domain.Plugin
}

type (
	previewMsg struct{ preview service.Preview }
	planMsg    struct{ plan domain.ActionPlan }
	errorMsg   struct{ err error }
	tickMsg    time.Time
)

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(now time.Time) tea.Msg { return tickMsg(now) })
}

func (m Model) loadDashboard() tea.Cmd {
	return func() tea.Msg { return m.fetchDashboard() }
}

func (m Model) previewDeploy() tea.Cmd {
	project, projectOK := m.selectedProject()
	target, targetOK := m.selectedTarget()
	prompt := m.prompt.Value()
	return func() tea.Msg {
		if !projectOK || !targetOK {
			return errorMsg{err: errSelectionRequired}
		}
		plan, err := m.app.PreviewDeploy(m.ctx, service.PreviewInput{
			ProjectID: string(project.ID()), TargetID: string(target.ID()), Prompt: prompt,
		})
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

func (m Model) approvePlan() tea.Cmd {
	plan, ok := m.selectedPlan()
	return func() tea.Msg {
		if !ok {
			return errorMsg{err: errSelectionRequired}
		}
		approved, err := m.app.ApprovePlan(m.ctx, string(plan.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		return planMsg{plan: approved}
	}
}

func (m Model) executePlan() tea.Cmd {
	plan, ok := m.selectedPlan()
	return func() tea.Msg {
		if !ok {
			return errorMsg{err: errSelectionRequired}
		}
		finished, err := m.app.StartPlan(m.ctx, string(plan.ID()))
		if err != nil {
			return errorMsg{err: err}
		}
		return planMsg{plan: finished}
	}
}
