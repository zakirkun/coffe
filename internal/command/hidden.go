package command

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

func readHidden(command *cobra.Command, prompt string) (string, error) {
	if _, err := fmt.Fprint(command.ErrOrStderr(), prompt); err != nil {
		return "", fmt.Errorf("write hidden prompt: %w", err)
	}
	value, err := term.ReadPassword(os.Stdin.Fd())
	if err != nil {
		return "", fmt.Errorf("read hidden value: %w", err)
	}
	if _, err := fmt.Fprintln(command.ErrOrStderr()); err != nil {
		return "", fmt.Errorf("write prompt newline: %w", err)
	}
	return string(value), nil
}
