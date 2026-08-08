package command

import (
	"coffe/internal/bootstrap"
	"coffe/internal/domain"
	"coffe/internal/service"
	"coffe/internal/version"
	"fmt"
	"os"

	coffeupdate "coffe/internal/update"

	"github.com/spf13/cobra"
)

func newUpdateCommand(state *rootState) *cobra.Command {
	command := &cobra.Command{Use: "update", Short: "Check and stage verified binary updates"}
	command.AddCommand(newUpdateCheckCommand(state), newUpdateDownloadCommand(state))
	return command
}

func newUpdateCheckCommand(state *rootState) *cobra.Command {
	var manifestURL, current string
	command := &cobra.Command{
		Use: "check", Args: cobra.NoArgs, Short: "Check an HTTPS release manifest",
		RunE: func(command *cobra.Command, _ []string) error {
			release, err := checkUpdate(state, manifestURL, current)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(command.OutOrStdout(), "%s available=%t\n", release.Version, release.Available); err != nil {
				return fmt.Errorf("write update status: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringVar(&manifestURL, "manifest", "", "HTTPS release manifest URL")
	command.Flags().StringVar(&current, "current", version.Semver(), "current semantic version")
	return command
}

func newUpdateDownloadCommand(state *rootState) *cobra.Command {
	var manifestURL, current, outputPath string
	command := &cobra.Command{
		Use: "download", Args: cobra.NoArgs, Short: "Verify and stage a newer platform asset",
		RunE: func(_ *cobra.Command, _ []string) error {
			release, err := checkUpdate(state, manifestURL, current)
			if err != nil {
				return err
			}
			if outputPath == "" {
				executable, err := os.Executable()
				if err != nil {
					return fmt.Errorf("locate executable: %w", err)
				}
				outputPath = executable + ".new"
			}
			client, err := coffeupdate.NewClient(coffeupdate.ClientConfig{ManifestURL: manifestURL})
			if err != nil {
				return err
			}
			if err := client.Download(state.ctx, release, outputPath); err != nil {
				return err
			}
			return withRuntime(state, func(runtime *bootstrap.Runtime) error {
				return runtime.App.RecordAudit(state.ctx, service.AuditRecordInput{
					Category: domain.AuditUpdate, Action: "download", ResourceID: release.Version,
					Outcome: domain.AuditSuccess, Message: "verified update staged",
				})
			})
		},
	}
	command.Flags().StringVar(&manifestURL, "manifest", "", "HTTPS release manifest URL")
	command.Flags().StringVar(&current, "current", version.Semver(), "current semantic version")
	command.Flags().StringVar(&outputPath, "out", "", "absolute staged binary path")
	return command
}

func checkUpdate(state *rootState, manifestURL, current string) (coffeupdate.Release, error) {
	if current == "" {
		return coffeupdate.Release{}, fmt.Errorf("current version is unavailable; pass --current vX.Y.Z")
	}
	client, err := coffeupdate.NewClient(coffeupdate.ClientConfig{ManifestURL: manifestURL})
	if err != nil {
		return coffeupdate.Release{}, err
	}
	return client.Check(state.ctx, current)
}
