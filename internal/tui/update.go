package tui

import (
	"coffe/internal/domain"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
)

var errSelectionRequired = errors.New("select a project, target, and plan first")

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = typed.Width, typed.Height
		m.resizePrompt()
	case tea.KeyPressMsg:
		return m.updateKey(typed)
	case dashboardMsg:
		m.projects, m.targets, m.plans, m.events = typed.projects, typed.targets, typed.plans, typed.events
		m.agents, m.databases = typed.agents, typed.databases
		m.gitOperations, m.providers, m.aiUsage = typed.gitOperations, typed.providers, typed.aiUsage
		m.policy, m.audit, m.workers, m.plugins = typed.policy, typed.audit, typed.workers, typed.plugins
		m.clampSelection()
		m.loading = false
	case previewMsg:
		m.commands = typed.preview.Commands
		m.plans = append([]domain.ActionPlan{typed.preview.Plan}, m.plans...)
		m.planIndex = 0
		m.focus = focusJobs
		m.prompt.Blur()
		m.status = "Preview ready. Press a to approve."
		m.loading = false
	case planMsg:
		m.replaceSelectedPlan(typed.plan)
		m.status = "Plan is " + string(typed.plan.Status())
		m.loading = false
		return m, m.loadDashboard()
	case agentMsg:
		m.agents = append([]domain.AgentSession{typed.session}, m.agents...)
		m.status = "Agent started: " + string(typed.session.Tool())
		m.loading = false
	case queryMsg:
		m.queryResult = typed.result
		m.status = fmt.Sprintf("Query returned %d row(s)", len(typed.result.Rows))
		m.loading = false
	case swarmMsg:
		m.swarmNodes, m.swarmServices = typed.nodes, typed.services
		m.status = fmt.Sprintf("Swarm: %d node(s), %d service(s)", len(typed.nodes), len(typed.services))
		m.loading = false
	case statusMsg:
		m.status = string(typed)
		m.loading = false
	case gitStatusMsg:
		m.gitStatus = typed.status
		m.status = fmt.Sprintf("Git %s: ahead %d, behind %d", typed.status.Branch, typed.status.Ahead, typed.status.Behind)
		m.loading = false
	case gitPreviewMsg:
		m.gitOperations = append([]domain.GitOperation{typed.preview.Operation}, m.gitOperations...)
		m.status = "Git preview: " + typed.preview.Command
		m.loading = false
	case gitOperationMsg:
		m.status = "Git operation is " + string(typed.operation.Status())
		m.loading = false
		return m, m.loadDashboard()
	case errorMsg:
		m.status = "Error: " + typed.err.Error()
		m.loading = false
	case tickMsg:
		return m, tea.Batch(m.loadDashboard(), tick())
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	name := key.String()
	switch {
	case name == "ctrl+c" || (name == "q" && m.focus != focusPrompt):
		return m, tea.Quit
	case name == "tab":
		m.focus = (m.focus + 1) % 4
		return m, m.syncPromptFocus()
	case m.focus == focusPrompt:
		return m.updatePromptKey(key)
	default:
		return m.updateDashboardKey(name)
	}
}

func (m Model) updatePromptKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.focus = focusJobs
		m.prompt.Blur()
		return m, nil
	case "enter":
		m.loading = true
		m.status = "Creating preview..."
		return m, m.submitPrompt()
	}
	var command tea.Cmd
	m.prompt, command = m.prompt.Update(key)
	return m, command
}

func (m Model) updateDashboardKey(name string) (tea.Model, tea.Cmd) {
	switch name {
	case "up", "k":
		m.moveSelection(-1)
		if m.focus == focusProjects || m.focus == focusJobs {
			return m, m.loadDashboard()
		}
	case "down", "j":
		m.moveSelection(1)
		if m.focus == focusProjects || m.focus == focusJobs {
			return m, m.loadDashboard()
		}
	case "a":
		m.loading = true
		m.status = "Recording approval..."
		return m, m.approvePlan()
	case "x":
		m.loading = true
		m.status = "Executing approved plan..."
		return m, m.executePlan()
	case "c":
		m.loading = true
		return m, m.cancelPlan()
	case "y":
		m.loading = true
		return m, m.retryPlan()
	case "b":
		m.loading = true
		return m, m.rollbackPlan()
	case "s":
		m.loading = true
		return m, m.loadSwarm()
	case "r":
		m.loading = true
		return m, m.loadDashboard()
	}
	return m, nil
}

func (m *Model) syncPromptFocus() tea.Cmd {
	if m.focus == focusPrompt {
		return m.prompt.Focus()
	}
	m.prompt.Blur()
	return nil
}

func (m *Model) replaceSelectedPlan(plan domain.ActionPlan) {
	if m.planIndex >= 0 && m.planIndex < len(m.plans) {
		m.plans[m.planIndex] = plan
	}
}
