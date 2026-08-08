package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func Run(ctx context.Context, app Application) error {
	program := tea.NewProgram(NewModel(ctx, app))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	return nil
}
