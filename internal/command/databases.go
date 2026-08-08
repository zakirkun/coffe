package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

func newDatabaseCommand(state *rootState) *cobra.Command {
	database := &cobra.Command{Use: "database", Short: "Manage read-only database connections"}
	database.AddCommand(newSQLiteDatabaseCommand(state), newMySQLDatabaseCommand(state))
	return database
}

func newSQLiteDatabaseCommand(state *rootState) *cobra.Command {
	var name, path string
	command := &cobra.Command{
		Use: "add-sqlite", Short: "Register a local SQLite database", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				connection, err := runtime.App.CreateDatabase(state.ctx, service.CreateDatabaseInput{
					Name: name, Driver: "sqlite", SQLitePath: path,
				})
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), connection.ID()); err != nil {
					return fmt.Errorf("write database id: %w", err)
				}
				return nil
			})
		},
	}
	command.Flags().StringVar(&name, "name", "", "connection name")
	command.Flags().StringVar(&path, "path", "", "absolute SQLite database path")
	return command
}

func newMySQLDatabaseCommand(state *rootState) *cobra.Command {
	var input service.CreateDatabaseInput
	command := &cobra.Command{
		Use: "add-mysql", Short: "Register a MySQL read-only connection", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			input.Driver = "mysql"
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				connection, err := runtime.App.CreateDatabase(state.ctx, input)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), connection.ID()); err != nil {
					return fmt.Errorf("write database id: %w", err)
				}
				return nil
			})
		},
	}
	flags := command.Flags()
	flags.StringVar(&input.Name, "name", "", "connection name")
	flags.StringVar(&input.Host, "host", "", "MySQL host")
	flags.IntVar(&input.Port, "port", 3306, "MySQL port")
	flags.StringVar(&input.User, "user", "", "MySQL user")
	flags.StringVar(&input.Database, "database", "", "MySQL database")
	flags.StringVar(&input.SecretName, "secret", "", "password key in the OS keyring")
	flags.BoolVar(&input.TLS, "tls", true, "require TLS")
	return command
}
