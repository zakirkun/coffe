package tui

func (m *Model) clampSelection() {
	m.projectIndex = clampIndex(m.projectIndex, len(m.projects))
	m.targetIndex = clampIndex(m.targetIndex, len(m.targets))
	m.planIndex = clampIndex(m.planIndex, len(m.plans))
}

func clampIndex(index, length int) int {
	if length == 0 {
		return -1
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

func (m *Model) moveSelection(delta int) {
	switch m.focus {
	case focusProjects:
		m.projectIndex = clampIndex(m.projectIndex+delta, len(m.projects))
		m.targetIndex = 0
	case focusTargets:
		m.targetIndex = clampIndex(m.targetIndex+delta, len(m.targets))
	case focusJobs:
		m.planIndex = clampIndex(m.planIndex+delta, len(m.plans))
	case focusPrompt:
	}
}

func (m *Model) resizePrompt() {
	width := m.width - 6
	if m.width >= 100 {
		leftWidth := (m.width - 1) * 2 / 5
		width = m.width - leftWidth - 7
	}
	m.prompt.SetWidth(max(20, width))
	if m.width < 100 {
		m.prompt.SetHeight(1)
		return
	}
	m.prompt.SetHeight(3)
}
