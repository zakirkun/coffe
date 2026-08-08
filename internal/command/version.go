package command

import (
	"coffe/internal/version"
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use: "version", Args: cobra.NoArgs, Short: "Print build version information",
		RunE: func(command *cobra.Command, _ []string) error {
			if _, err := fmt.Fprintf(command.OutOrStdout(),
				"coffe %s commit=%s built=%s\n", version.Version, version.Commit, version.Date,
			); err != nil {
				return fmt.Errorf("write version: %w", err)
			}
			return nil
		},
	}
}
