//go:build wailsapp

package main

import (
	"errors"
	"fmt"

	"coffe/internal/backup"
	"coffe/internal/bootstrap"
	"coffe/internal/domain"
	"coffe/internal/service"
)

type BackupExportInput struct {
	OutputPath         string `json:"output_path"`
	Passphrase         string `json:"passphrase"`
	RepeatedPassphrase string `json:"repeated_passphrase"`
}

type BackupImportInput struct {
	InputPath  string `json:"input_path"`
	Passphrase string `json:"passphrase"`
	Confirmed  bool   `json:"confirmed"`
}

func (a *App) ExportBackup(input BackupExportInput) error {
	if input.Passphrase != input.RepeatedPassphrase {
		return fmt.Errorf("backup passphrases do not match")
	}
	return a.withRuntime(func(runtimeState *bootstrap.Runtime) error {
		if err := backup.Export(a.ctx, backup.ExportOptions{
			DatabasePath: a.config.DatabasePath,
			OutputPath:   input.OutputPath,
			Passphrase:   input.Passphrase,
		}); err != nil {
			return err
		}
		return runtimeState.App.RecordAudit(a.ctx, service.AuditRecordInput{
			Category: domain.AuditBackup, Action: "export", ResourceID: "workspace",
			Outcome: domain.AuditSuccess, Message: "encrypted workspace backup exported",
		})
	})
}

func (a *App) ImportBackup(input BackupImportInput) error {
	if !input.Confirmed {
		return fmt.Errorf("backup import requires confirmation")
	}
	a.lifecycle.Lock()
	defer a.lifecycle.Unlock()
	if a.runtime == nil {
		return ErrRuntimeUnavailable
	}
	closeErr := a.runtime.Close(a.ctx)
	a.runtime = nil
	a.bridge = nil
	if closeErr != nil {
		runtimeState, openErr := bootstrap.Open(a.ctx, a.config, a.logger)
		if openErr == nil {
			a.installRuntime(runtimeState)
		}
		return errors.Join(closeErr, openErr)
	}
	importErr := backup.Import(a.ctx, backup.ImportOptions{
		DatabasePath: a.config.DatabasePath,
		InputPath:    input.InputPath,
		Passphrase:   input.Passphrase,
	})
	runtimeState, openErr := bootstrap.Open(a.ctx, a.config, a.logger)
	if openErr != nil {
		return errors.Join(importErr, openErr)
	}
	a.installRuntime(runtimeState)
	if importErr != nil {
		return importErr
	}
	return runtimeState.App.RecordAudit(a.ctx, service.AuditRecordInput{
		Category: domain.AuditBackup, Action: "import", ResourceID: "workspace",
		Outcome: domain.AuditSuccess, Message: "encrypted workspace backup imported",
	})
}
