package tui

import (
	"strconv"
	"strings"
)

func (m Model) renderJobs() string {
	sections := []string{
		m.renderDeploymentJobs(), m.renderAgentAndDatabaseJobs(),
		m.renderGitJobs(), m.renderRuntimeSummary(), m.renderQuerySummary(),
	}
	var visible []string
	for _, section := range sections {
		if section != "" {
			visible = append(visible, section)
		}
	}
	return strings.Join(visible, "\n")
}

func (m Model) renderDeploymentJobs() string {
	var builder strings.Builder
	if len(m.plans) == 0 {
		builder.WriteString(m.styles.mutedRow.Render("No deployment plans yet.") + "\n")
	}
	for index, plan := range m.plans {
		row := "[" + string(plan.Status()) + "] " + string(plan.ID())
		if index == m.planIndex {
			row = m.styles.selected.Render("> " + row)
		}
		builder.WriteString(row + "\n")
	}
	if len(m.commands) > 0 {
		builder.WriteString(m.styles.mutedRow.Render("Preview") + "\n")
		for _, command := range m.commands {
			builder.WriteString("  " + command + "\n")
		}
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func (m Model) renderAgentAndDatabaseJobs() string {
	var builder strings.Builder
	if len(m.agents) > 0 {
		builder.WriteString(m.styles.mutedRow.Render("Agents") + "\n")
		for _, session := range m.agents {
			builder.WriteString("  [" + string(session.Status()) + "] " +
				string(session.Tool()) + " " + string(session.ID()) + "\n")
		}
	}
	if len(m.databases) > 0 {
		builder.WriteString(m.styles.mutedRow.Render("Databases") + "\n")
		for _, database := range m.databases {
			builder.WriteString("  " + database.Name().String() + " [" + string(database.Driver()) + "]\n")
		}
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func (m Model) renderGitJobs() string {
	var builder strings.Builder
	if m.gitStatus.Branch != "" {
		builder.WriteString("Git " + m.gitStatus.Branch + "  ahead " + strconv.Itoa(m.gitStatus.Ahead) +
			"  behind " + strconv.Itoa(m.gitStatus.Behind) + "\n")
	}
	if len(m.gitOperations) > 0 {
		builder.WriteString(m.styles.mutedRow.Render("Git operations") + "\n")
		for _, operation := range m.gitOperations {
			builder.WriteString("  [" + string(operation.Status()) + "] " +
				string(operation.Kind()) + " " + string(operation.ID()) + "\n")
		}
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func (m Model) renderRuntimeSummary() string {
	var lines []string
	if len(m.providers) > 0 {
		lines = append(lines, "AI providers  "+strconv.Itoa(len(m.providers))+
			"  usage records "+strconv.Itoa(len(m.aiUsage)))
	}
	if !m.policy.UpdatedAt().IsZero() {
		lines = append(lines, "Policy  agents "+strconv.Itoa(m.policy.MaxConcurrentAgents())+
			"  audit records "+strconv.Itoa(len(m.audit)))
	}
	if len(m.workers) > 0 {
		lines = append(lines, "Workers  "+strconv.Itoa(len(m.workers)))
	}
	if len(m.plugins) > 0 {
		lines = append(lines, "Plugins  "+strconv.Itoa(len(m.plugins)))
	}
	if len(m.swarmNodes) > 0 || len(m.swarmServices) > 0 {
		lines = append(lines, "Swarm  "+strconv.Itoa(len(m.swarmNodes))+
			" node(s)  "+strconv.Itoa(len(m.swarmServices))+" service(s)")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderQuerySummary() string {
	if len(m.queryResult.Columns) == 0 {
		return ""
	}
	result := m.styles.mutedRow.Render("Query") + "\n  " + strings.Join(m.queryResult.Columns, " | ")
	if len(m.queryResult.Rows) > 0 {
		result += "\n  " + strings.Join(m.queryResult.Rows[0], " | ")
	}
	return result
}
