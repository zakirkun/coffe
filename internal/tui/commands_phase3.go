package tui

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type (
	gitStatusMsg    struct{ status domain.GitStatus }
	gitPreviewMsg   struct{ preview service.GitPreview }
	gitOperationMsg struct{ operation domain.GitOperation }
)

func (m Model) gitPrompt(prompt string) tea.Cmd {
	parts := strings.Fields(prompt)
	return func() tea.Msg {
		if len(parts) < 2 {
			return errorMsg{err: fmt.Errorf("use git status, pull, push, approve ID, or execute ID")}
		}
		switch strings.ToLower(parts[1]) {
		case "status":
			return m.gitStatusMessage()
		case "pull", "push":
			return m.gitPreviewMessage(strings.ToLower(parts[1]))
		case "approve":
			if len(parts) != 3 {
				return errorMsg{err: fmt.Errorf("use git approve OPERATION_ID")}
			}
			operation, err := m.app.ApproveGit(m.ctx, parts[2])
			if err != nil {
				return errorMsg{err: err}
			}
			return gitOperationMsg{operation: operation}
		case "execute":
			if len(parts) != 3 {
				return errorMsg{err: fmt.Errorf("use git execute OPERATION_ID")}
			}
			operation, err := m.app.ExecuteGit(m.ctx, parts[2])
			if err != nil {
				return errorMsg{err: err}
			}
			return gitOperationMsg{operation: operation}
		default:
			return errorMsg{err: fmt.Errorf("unknown git prompt operation")}
		}
	}
}

func (m Model) gitStatusMessage() tea.Msg {
	project, ok := m.selectedProject()
	if !ok {
		return errorMsg{err: errSelectionRequired}
	}
	status, err := m.app.GitStatus(m.ctx, string(project.ID()))
	if err != nil {
		return errorMsg{err: err}
	}
	return gitStatusMsg{status: status}
}

func (m Model) gitPreviewMessage(kind string) tea.Msg {
	project, ok := m.selectedProject()
	if !ok {
		return errorMsg{err: errSelectionRequired}
	}
	status, err := m.app.GitStatus(m.ctx, string(project.ID()))
	if err != nil {
		return errorMsg{err: err}
	}
	preview, err := m.app.PreviewGit(m.ctx, service.CreateGitOperationInput{
		ProjectID: string(project.ID()), Kind: kind, Remote: "origin", Branch: status.Branch,
	})
	if err != nil {
		return errorMsg{err: err}
	}
	return gitPreviewMsg{preview: preview}
}
