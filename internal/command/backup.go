package command

import (
	"coffe/internal/backup"
	"coffe/internal/bootstrap"
	"coffe/internal/domain"
	"coffe/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

func newBackupCommand(state *rootState) *cobra.Command {
	command := &cobra.Command{Use: "backup", Short: "Export or import an encrypted workspace snapshot"}
	command.AddCommand(newBackupExportCommand(state), newBackupImportCommand(state))
	return command
}

func newBackupExportCommand(state *rootState) *cobra.Command {
	var outputPath string
	command := &cobra.Command{
		Use: "export", Args: cobra.NoArgs, Short: "Export an age-encrypted SQLite snapshot",
		RunE: func(command *cobra.Command, _ []string) error {
			passphrase, err := confirmedPassphrase(command)
			if err != nil {
				return err
			}
			if err := backup.Export(state.ctx, backup.ExportOptions{
				DatabasePath: state.options.DatabasePath, OutputPath: outputPath, Passphrase: passphrase,
			}); err != nil {
				return err
			}
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				return runtime.App.RecordAudit(state.ctx, service.AuditRecordInput{
					Category: domain.AuditBackup, Action: "export", ResourceID: "workspace",
					Outcome: domain.AuditSuccess, Message: "encrypted workspace backup exported",
				})
			})
		},
	}
	command.Flags().StringVar(&outputPath, "out", "", "absolute output .age file")
	return command
}

func newBackupImportCommand(state *rootState) *cobra.Command {
	var inputPath string
	var confirmed bool
	command := &cobra.Command{
		Use: "import", Args: cobra.NoArgs, Short: "Replace local state from an encrypted snapshot",
		RunE: func(command *cobra.Command, _ []string) error {
			if !confirmed {
				return fmt.Errorf("backup import requires --confirm")
			}
			passphrase, err := readHidden(command, "Backup passphrase: ")
			if err != nil {
				return err
			}
			if err := backup.Import(state.ctx, backup.ImportOptions{
				DatabasePath: state.options.DatabasePath, InputPath: inputPath, Passphrase: passphrase,
			}); err != nil {
				return err
			}
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				return runtime.App.RecordAudit(state.ctx, service.AuditRecordInput{
					Category: domain.AuditBackup, Action: "import", ResourceID: "workspace",
					Outcome: domain.AuditSuccess, Message: "encrypted workspace backup imported",
				})
			})
		},
	}
	command.Flags().StringVar(&inputPath, "in", "", "absolute input .age file")
	command.Flags().BoolVar(&confirmed, "confirm", false, "confirm replacement of the current SQLite state")
	return command
}

func confirmedPassphrase(command *cobra.Command) (string, error) {
	first, err := readHidden(command, "Backup passphrase: ")
	if err != nil {
		return "", err
	}
	second, err := readHidden(command, "Repeat passphrase: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", fmt.Errorf("backup passphrases do not match")
	}
	return first, nil
}
