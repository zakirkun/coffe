package command

import (
	"coffe/internal/secrets"
	"strings"

	"github.com/spf13/cobra"
)

func newSecretCommand() *cobra.Command {
	secret := &cobra.Command{Use: "secret", Short: "Manage values in the OS keyring"}
	set := &cobra.Command{
		Use: "set NAME", Short: "Store a secret without writing it to SQLite", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			value, err := readHidden(command, "Secret value: ")
			if err != nil {
				return err
			}
			keyring, err := secrets.NewKeyring("coffe")
			if err != nil {
				return err
			}
			return keyring.Set(command.Context(), args[0], strings.TrimSpace(value))
		},
	}
	secret.AddCommand(set)
	return secret
}
