package command

import (
	"coffe/internal/bootstrap"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newAuditCommand(state *rootState) *cobra.Command {
	command := &cobra.Command{Use: "audit", Short: "Inspect or export redacted audit records"}
	command.AddCommand(newAuditListCommand(state), newAuditExportCommand(state))
	return command
}

func newAuditListCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "list", Args: cobra.NoArgs, Short: "List recent audit records",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				entries, err := runtime.App.ListAudit(state.ctx, 100)
				if err != nil {
					return err
				}
				for _, entry := range entries {
					if _, err := fmt.Fprintf(command.OutOrStdout(), "%s %s %s %s %s\n",
						entry.OccurredAt().Format("2006-01-02T15:04:05Z07:00"),
						entry.Category(), entry.Action(), entry.Outcome(), entry.ResourceID(),
					); err != nil {
						return fmt.Errorf("write audit record: %w", err)
					}
				}
				return nil
			})
		},
	}
}

func newAuditExportCommand(state *rootState) *cobra.Command {
	var outputPath string
	command := &cobra.Command{
		Use: "export", Args: cobra.NoArgs, Short: "Export audit records as formula-safe CSV",
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			return withRuntime(state, func(runtime *bootstrap.Runtime) (err error) {
				if !filepath.IsAbs(filepath.Clean(outputPath)) {
					return fmt.Errorf("audit output path must be absolute")
				}
				content, err := runtime.App.ExportAuditCSV(state.ctx, 1000)
				if err != nil {
					return err
				}
				output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) //nolint:gosec // path is absolute, exclusive, and mode 0600.
				if err != nil {
					return fmt.Errorf("create audit export: %w", err)
				}
				defer func() { err = errors.Join(err, output.Close()) }()
				if _, err := output.Write(content); err != nil {
					return fmt.Errorf("write audit export: %w", err)
				}
				return nil
			})
		},
	}
	command.Flags().StringVar(&outputPath, "out", "", "absolute CSV output path")
	return command
}
