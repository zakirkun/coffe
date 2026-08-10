//go:build wailsapp

package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"coffe/internal/bootstrap"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("desktop failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	config, err := desktopOptions()
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runtimeState, err := bootstrap.Open(ctx, config, logger)
	if err != nil {
		return err
	}
	app := newApp(runtimeState, config, logger)
	defer func() { err = errors.Join(err, app.close(ctx)) }()
	if err := wails.Run(&options.App{
		Title:                    "Coffe",
		Width:                    1440,
		Height:                   900,
		MinWidth:                 900,
		MinHeight:                640,
		BackgroundColour:         &options.RGBA{R: 14, G: 16, B: 17, A: 255},
		EnableDefaultContextMenu: false,
		AssetServer:              &assetserver.Options{Assets: assets},
		OnStartup:                app.startup,
		Bind:                     []interface{}{app},
		Windows: &windows.Options{
			Theme:                windows.SystemDefault,
			IsZoomControlEnabled: true,
		},
	}); err != nil {
		return fmt.Errorf("run Wails desktop: %w", err)
	}
	return nil
}

func desktopOptions() (bootstrap.Options, error) {
	config, err := bootstrap.DefaultOptions()
	if err != nil {
		return bootstrap.Options{}, err
	}
	databasePath := os.Getenv("COFFE_DB")
	if databasePath == "" {
		return config, nil
	}
	clean := filepath.Clean(databasePath)
	if !filepath.IsAbs(clean) {
		return bootstrap.Options{}, fmt.Errorf("COFFE_DB must be an absolute path")
	}
	config.DatabasePath = clean
	return config, nil
}
