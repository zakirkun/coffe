//go:build wailsapp

package main

import (
	"fmt"
	"path/filepath"

	"coffe/internal/bootstrap"
	"coffe/internal/pluginexec"
	"coffe/internal/service"
)

type PluginRegisterInput struct {
	Name           string   `json:"name"`
	ExecutablePath string   `json:"executable_path"`
	Methods        []string `json:"methods"`
	Enabled        bool     `json:"enabled"`
}

func (a *App) RegisterPlugin(input PluginRegisterInput) (string, error) {
	path, err := filepath.Abs(input.ExecutablePath)
	if err != nil {
		return "", fmt.Errorf("resolve plugin path: %w", err)
	}
	hash, err := pluginexec.HashFile(path)
	if err != nil {
		return "", err
	}
	var pluginID string
	err = a.withRuntime(func(runtimeState *bootstrap.Runtime) error {
		plugin, err := runtimeState.App.RegisterPlugin(a.ctx, service.RegisterPluginInput{
			Name: input.Name, ExecutablePath: path, SHA256: hash,
			Methods: input.Methods, Enabled: input.Enabled,
		})
		if err != nil {
			return err
		}
		pluginID = string(plugin.ID())
		return nil
	})
	return pluginID, err
}
