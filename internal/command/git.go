package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

func newGitCommand(state *rootState) *cobra.Command {
	git := &cobra.Command{Use: "git", Short: "Inspect and approve project Git operations"}
	git.AddCommand(
		newGitStatusCommand(state), newGitDiffCommand(state), newGitPreviewCommand(state),
		newGitApproveCommand(state), newGitExecuteCommand(state),
	)
	return git
}

func newGitStatusCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "status PROJECT_ID", Args: cobra.ExactArgs(1), Short: "Read project Git status",
		RunE: func(command *cobra.Command, args []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				status, err := runtime.App.GitStatus(state.ctx, args[0])
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintf(command.OutOrStdout(),
					"%s ahead=%d behind=%d staged=%d modified=%d untracked=%d\n",
					status.Branch, status.Ahead, status.Behind, status.Staged, status.Modified, status.Untracked,
				); err != nil {
					return fmt.Errorf("write git status: %w", err)
				}
				return nil
			})
		},
	}
}

func newGitDiffCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "diff PROJECT_ID", Args: cobra.ExactArgs(1), Short: "Read capped project diff",
		RunE: func(command *cobra.Command, args []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				diff, err := runtime.App.GitDiff(state.ctx, args[0])
				if err != nil {
					return err
				}
				if _, err := fmt.Fprint(command.OutOrStdout(), diff); err != nil {
					return fmt.Errorf("write git diff: %w", err)
				}
				return nil
			})
		},
	}
}

func newGitPreviewCommand(state *rootState) *cobra.Command {
	var input service.CreateGitOperationInput
	command := &cobra.Command{
		Use: "preview", Args: cobra.NoArgs, Short: "Preview a pull or push operation",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				preview, err := runtime.App.PreviewGit(state.ctx, input)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintf(command.OutOrStdout(), "%s\n%s\n", preview.Operation.ID(), preview.Command); err != nil {
					return fmt.Errorf("write git preview: %w", err)
				}
				return nil
			})
		},
	}
	flags := command.Flags()
	flags.StringVar(&input.ProjectID, "project", "", "registered project ID")
	flags.StringVar(&input.Kind, "kind", "pull", "operation kind: pull or push")
	flags.StringVar(&input.Remote, "remote", "origin", "Git remote name")
	flags.StringVar(&input.Branch, "branch", "main", "Git branch")
	return command
}

func newGitApproveCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "approve OPERATION_ID", Args: cobra.ExactArgs(1), Short: "Approve a Git operation",
		RunE: func(_ *cobra.Command, args []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				_, err := runtime.App.ApproveGit(state.ctx, args[0])
				return err
			})
		},
	}
}

func newGitExecuteCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "execute OPERATION_ID", Args: cobra.ExactArgs(1), Short: "Execute an approved Git operation",
		RunE: func(_ *cobra.Command, args []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				_, err := runtime.App.ExecuteGit(state.ctx, args[0])
				return err
			})
		},
	}
}
