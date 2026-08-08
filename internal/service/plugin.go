package service

import (
	"coffe/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type RegisterPluginInput struct {
	Name           string
	ExecutablePath string
	SHA256         string
	Methods        []string
	Enabled        bool
}

func (a *App) RegisterPlugin(ctx context.Context, input RegisterPluginInput) (domain.Plugin, error) {
	name, err := domain.NewName(input.Name)
	if err != nil {
		return domain.Plugin{}, err
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.Plugin{}, fmt.Errorf("new plugin id: %w", err)
	}
	plugin, err := domain.NewPlugin(domain.PluginValues{
		ID: domain.PluginID(rawID), Name: name, ExecutablePath: input.ExecutablePath,
		SHA256: input.SHA256, Methods: input.Methods, Enabled: input.Enabled, CreatedAt: a.clock.Now(),
	})
	if err != nil {
		return domain.Plugin{}, err
	}
	if err := a.plugins.Verify(plugin); err != nil {
		return domain.Plugin{}, fmt.Errorf("verify plugin: %w", err)
	}
	if err := a.repository.SavePlugin(ctx, plugin); err != nil {
		return domain.Plugin{}, fmt.Errorf("save plugin: %w", err)
	}
	if err := a.appendAudit(ctx, auditInput{
		category: domain.AuditPlugin, action: "register", resourceID: string(plugin.ID()),
		outcome: domain.AuditSuccess, message: "hash-pinned plugin registered",
	}); err != nil {
		return domain.Plugin{}, err
	}
	return plugin, nil
}

func (a *App) ListPlugins(ctx context.Context) ([]domain.Plugin, error) {
	plugins, err := a.repository.ListPlugins(ctx)
	if err != nil {
		return nil, fmt.Errorf("list plugins: %w", err)
	}
	return plugins, nil
}

type PluginInvocationInput struct {
	PluginID string
	Method   string
	Input    json.RawMessage
}

func (a *App) InvokePlugin(ctx context.Context, input PluginInvocationInput) (json.RawMessage, error) {
	id := domain.PluginID(strings.TrimSpace(input.PluginID))
	if id == "" || string(id) != input.PluginID {
		return nil, fmt.Errorf("plugin id: %w", ErrInvalidInput)
	}
	plugin, err := a.repository.GetPlugin(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get plugin: %w", err)
	}
	result, invokeErr := a.plugins.Invoke(ctx, plugin, input.Method, input.Input)
	outcome := domain.AuditSuccess
	message := "plugin invocation succeeded"
	if invokeErr != nil {
		outcome = domain.AuditFailure
		message = "plugin invocation failed"
	}
	auditErr := a.appendAudit(ctx, auditInput{
		category: domain.AuditPlugin, action: "invoke", resourceID: string(plugin.ID()),
		outcome: outcome, message: message,
	})
	if invokeErr != nil || auditErr != nil {
		return nil, errors.Join(invokeErr, auditErr)
	}
	return result, nil
}
