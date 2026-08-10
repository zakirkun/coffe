//go:build wailsapp

package main

import (
	"fmt"

	"coffe/internal/bootstrap"
	"coffe/internal/domain"
	"coffe/internal/service"
	coffeupdate "coffe/internal/update"
	"coffe/internal/version"
)

type UpdateCheckInput struct {
	ManifestURL    string `json:"manifest_url"`
	CurrentVersion string `json:"current_version"`
}

type UpdateDownloadInput struct {
	ManifestURL    string `json:"manifest_url"`
	CurrentVersion string `json:"current_version"`
	OutputPath     string `json:"output_path"`
}

type UpdateRelease struct {
	Version   string `json:"version"`
	Available bool   `json:"available"`
	AssetURL  string `json:"asset_url"`
}

type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func (a *App) CheckUpdate(input UpdateCheckInput) (UpdateRelease, error) {
	release, err := checkDesktopUpdate(a, input.ManifestURL, input.CurrentVersion)
	if err != nil {
		return UpdateRelease{}, err
	}
	return UpdateRelease{
		Version:   release.Version,
		Available: release.Available,
		AssetURL:  release.Asset.URL,
	}, nil
}

func (a *App) DownloadUpdate(input UpdateDownloadInput) error {
	release, err := checkDesktopUpdate(a, input.ManifestURL, input.CurrentVersion)
	if err != nil {
		return err
	}
	client, err := coffeupdate.NewClient(coffeupdate.ClientConfig{ManifestURL: input.ManifestURL})
	if err != nil {
		return err
	}
	if err := client.Download(a.ctx, release, input.OutputPath); err != nil {
		return err
	}
	return a.withRuntime(func(runtimeState *bootstrap.Runtime) error {
		return runtimeState.App.RecordAudit(a.ctx, service.AuditRecordInput{
			Category: domain.AuditUpdate, Action: "download", ResourceID: release.Version,
			Outcome: domain.AuditSuccess, Message: "verified update staged",
		})
	})
}

func (a *App) VersionInfo() VersionInfo {
	return VersionInfo{Version: version.Version, Commit: version.Commit, Date: version.Date}
}

func checkDesktopUpdate(a *App, manifestURL, currentVersion string) (coffeupdate.Release, error) {
	if currentVersion == "" {
		return coffeupdate.Release{}, fmt.Errorf("current version is unavailable")
	}
	client, err := coffeupdate.NewClient(coffeupdate.ClientConfig{ManifestURL: manifestURL})
	if err != nil {
		return coffeupdate.Release{}, err
	}
	return client.Check(a.ctx, currentVersion)
}
