package service_test

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
)

func (r *memoryRepository) SavePlugin(_ context.Context, plugin domain.Plugin) error {
	r.plugins[plugin.ID()] = plugin
	return nil
}

func (r *memoryRepository) GetPlugin(_ context.Context, id domain.PluginID) (domain.Plugin, error) {
	plugin, ok := r.plugins[id]
	if !ok {
		return domain.Plugin{}, service.ErrNotFound
	}
	return plugin, nil
}

func (r *memoryRepository) ListPlugins(_ context.Context) ([]domain.Plugin, error) {
	plugins := make([]domain.Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}
