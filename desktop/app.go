//go:build wailsapp

package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"coffe/internal/bootstrap"
	"coffe/internal/desktop"
	"coffe/internal/httpapi"
)

var ErrRuntimeUnavailable = errors.New("desktop runtime unavailable")

type App struct {
	ctx       context.Context
	config    bootstrap.Options
	logger    *slog.Logger
	lifecycle sync.RWMutex
	runtime   *bootstrap.Runtime
	bridge    *desktop.Bridge
}

func newApp(runtimeState *bootstrap.Runtime, config bootstrap.Options, logger *slog.Logger) *App {
	return &App{
		config:  config,
		logger:  logger,
		runtime: runtimeState,
		bridge:  desktop.NewBridge(httpapi.New(runtimeState.App, logger)),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Request(input desktop.Request) (desktop.Response, error) {
	a.lifecycle.RLock()
	defer a.lifecycle.RUnlock()
	if a.bridge == nil {
		return desktop.Response{}, ErrRuntimeUnavailable
	}
	return a.bridge.Request(a.ctx, input)
}

func (a *App) close(ctx context.Context) error {
	a.lifecycle.Lock()
	defer a.lifecycle.Unlock()
	if a.runtime == nil {
		return nil
	}
	err := a.runtime.Close(ctx)
	a.runtime = nil
	a.bridge = nil
	return err
}

func (a *App) installRuntime(runtimeState *bootstrap.Runtime) {
	a.runtime = runtimeState
	a.bridge = desktop.NewBridge(httpapi.New(runtimeState.App, a.logger))
}

func (a *App) withRuntime(action func(*bootstrap.Runtime) error) error {
	a.lifecycle.RLock()
	defer a.lifecycle.RUnlock()
	if a.runtime == nil {
		return ErrRuntimeUnavailable
	}
	return action(a.runtime)
}
