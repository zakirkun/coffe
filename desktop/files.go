//go:build wailsapp

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"coffe/internal/bootstrap"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type QueryExportInput struct {
	DatabaseID string `json:"database_id"`
	SQL        string `json:"sql"`
	OutputPath string `json:"output_path"`
}

type dialogSpec struct {
	Title       string
	Filename    string
	DisplayName string
	Pattern     string
}

func (a *App) ChoosePath(kind string) (string, error) {
	switch kind {
	case "project_directory":
		return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
			Title:                "Choose project directory",
			CanCreateDirectories: true,
		})
	case "sqlite_database":
		return wailsruntime.OpenFileDialog(a.ctx, openDialog(dialogSpec{
			Title: "Choose SQLite database", DisplayName: "SQLite databases (*.db, *.sqlite, *.sqlite3)",
			Pattern: "*.db;*.sqlite;*.sqlite3",
		}))
	case "plugin_executable":
		return wailsruntime.OpenFileDialog(a.ctx, openDialog(dialogSpec{
			Title: "Choose plugin executable", DisplayName: "All files", Pattern: "*",
		}))
	case "backup_import":
		return wailsruntime.OpenFileDialog(a.ctx, openDialog(dialogSpec{
			Title: "Choose encrypted backup", DisplayName: "Age backups (*.age)", Pattern: "*.age",
		}))
	case "backup_export":
		return wailsruntime.SaveFileDialog(a.ctx, saveDialog(dialogSpec{
			Title: "Export encrypted backup", Filename: "coffe-backup.age",
			DisplayName: "Age backups (*.age)", Pattern: "*.age",
		}))
	case "audit_export":
		return wailsruntime.SaveFileDialog(a.ctx, saveDialog(dialogSpec{
			Title: "Export audit ledger", Filename: "coffe-audit.csv",
			DisplayName: "CSV files (*.csv)", Pattern: "*.csv",
		}))
	case "query_export":
		return wailsruntime.SaveFileDialog(a.ctx, saveDialog(dialogSpec{
			Title: "Export query result", Filename: "coffe-query.csv",
			DisplayName: "CSV files (*.csv)", Pattern: "*.csv",
		}))
	case "update_output":
		return wailsruntime.SaveFileDialog(a.ctx, saveDialog(dialogSpec{
			Title: "Stage verified update", Filename: "coffe.new", DisplayName: "All files", Pattern: "*",
		}))
	default:
		return "", fmt.Errorf("path kind %q: invalid", kind)
	}
}

func (a *App) ExportAudit(outputPath string) error {
	return a.withRuntime(func(runtimeState *bootstrap.Runtime) error {
		content, err := runtimeState.App.ExportAuditCSV(a.ctx, 1000)
		if err != nil {
			return err
		}
		return writeExclusive(outputPath, content)
	})
}

func (a *App) ExportQuery(input QueryExportInput) error {
	return a.withRuntime(func(runtimeState *bootstrap.Runtime) error {
		content, err := runtimeState.App.ExportDatabaseCSV(a.ctx, input.DatabaseID, input.SQL)
		if err != nil {
			return err
		}
		return writeExclusive(input.OutputPath, content)
	})
}

func openDialog(spec dialogSpec) wailsruntime.OpenDialogOptions {
	return wailsruntime.OpenDialogOptions{
		Title:   spec.Title,
		Filters: []wailsruntime.FileFilter{{DisplayName: spec.DisplayName, Pattern: spec.Pattern}},
	}
}

func saveDialog(spec dialogSpec) wailsruntime.SaveDialogOptions {
	return wailsruntime.SaveDialogOptions{
		Title:                spec.Title,
		DefaultFilename:      spec.Filename,
		CanCreateDirectories: true,
		Filters:              []wailsruntime.FileFilter{{DisplayName: spec.DisplayName, Pattern: spec.Pattern}},
	}
}

func writeExclusive(path string, content []byte) (err error) {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return fmt.Errorf("export path must be absolute")
	}
	output, err := os.OpenFile(clean, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create export: %w", err)
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	if _, err := output.Write(content); err != nil {
		return fmt.Errorf("write export: %w", err)
	}
	return nil
}
