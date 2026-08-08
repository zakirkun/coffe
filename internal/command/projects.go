package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

func newProjectCommand(state *rootState) *cobra.Command {
	project := &cobra.Command{Use: "project", Short: "Manage local projects"}
	var name, rootPath string
	add := &cobra.Command{
		Use: "add", Short: "Register a local project", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				created, err := runtime.App.CreateProject(state.ctx, service.CreateProjectInput{
					Name: name, RootPath: rootPath,
				})
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), created.ID()); err != nil {
					return fmt.Errorf("write project id: %w", err)
				}
				return nil
			})
		},
	}
	add.Flags().StringVar(&name, "name", "", "project name")
	add.Flags().StringVar(&rootPath, "root", "", "absolute local project path")
	project.AddCommand(add)
	return project
}

func newTargetCommand(state *rootState) *cobra.Command {
	target := &cobra.Command{Use: "target", Short: "Manage deployment targets"}
	var input service.CreateTargetInput
	add := &cobra.Command{
		Use: "add", Short: "Register an SSH deployment target", Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				created, err := runtime.App.CreateTarget(state.ctx, input)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(command.OutOrStdout(), created.ID()); err != nil {
					return fmt.Errorf("write target id: %w", err)
				}
				return nil
			})
		},
	}
	flags := add.Flags()
	flags.StringVar(&input.ProjectID, "project", "", "registered project ID")
	flags.StringVar(&input.Name, "name", "", "target name")
	flags.StringVar(&input.Host, "host", "", "SSH host")
	flags.IntVar(&input.Port, "port", 22, "SSH port")
	flags.StringVar(&input.User, "user", "", "SSH user")
	flags.StringVar(&input.RemotePath, "path", "", "absolute project path on the server")
	flags.StringVar(&input.Branch, "branch", "main", "deployment branch")
	target.AddCommand(add)
	return target
}
