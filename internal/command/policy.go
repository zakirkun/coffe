package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/service"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newPolicyCommand(state *rootState) *cobra.Command {
	command := &cobra.Command{Use: "policy", Short: "Read or update local safety policy"}
	command.AddCommand(newPolicyShowCommand(state), newPolicySetCommand(state))
	return command
}

func newPolicyShowCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use: "show", Args: cobra.NoArgs, Short: "Print local safety policy",
		RunE: func(command *cobra.Command, _ []string) error {
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				policy, err := runtime.App.GetPolicy(state.ctx)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintf(command.OutOrStdout(),
					"git_push=%t swarm=%t max_agents=%d audit_days=%d tools=%v\n",
					policy.AllowGitPush(), policy.AllowSwarmDeploy(), policy.MaxConcurrentAgents(),
					policy.AuditRetentionDays(), policy.AgentTools(),
				); err != nil {
					return fmt.Errorf("write policy: %w", err)
				}
				return nil
			})
		},
	}
}

func newPolicySetCommand(state *rootState) *cobra.Command {
	var input service.UpdatePolicyInput
	var tools string
	command := &cobra.Command{
		Use: "set", Args: cobra.NoArgs, Short: "Replace local safety policy",
		RunE: func(_ *cobra.Command, _ []string) error {
			input.AgentTools = strings.Split(tools, ",")
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				_, err := runtime.App.UpdatePolicy(state.ctx, input)
				return err
			})
		},
	}
	flags := command.Flags()
	flags.BoolVar(&input.AllowGitPush, "git-push", true, "allow approved Git push")
	flags.BoolVar(&input.AllowSwarmDeploy, "swarm", true, "allow approved Swarm deployment")
	flags.StringVar(&tools, "agent-tools", "opencode,claude,codex,gemini,aider", "comma-separated agent allowlist")
	flags.IntVar(&input.MaxConcurrentAgents, "max-agents", 4, "maximum concurrent coding agents")
	flags.IntVar(&input.AuditRetentionDays, "audit-days", 365, "audit retention in days")
	return command
}
