package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

const maxBackupBytes int64 = 8 * 1024 * 1024 * 1024

var ErrInvalidBackup = errors.New("backup: invalid input")

func validatePassphrase(passphrase string) error {
	if utf8.RuneCountInString(passphrase) < 12 {
		return fmt.Errorf("passphrase must contain at least 12 characters: %w", ErrInvalidBackup)
	}
	return nil
}

func snapshot(ctx context.Context, databasePath string) (path string, err error) {
	if info, err := os.Stat(databasePath); err != nil {
		return "", fmt.Errorf("inspect source database: %w", err)
	} else if !info.Mode().IsRegular() {
		return "", fmt.Errorf("source database is not a regular file: %w", ErrInvalidBackup)
	}
	temporary, err := os.CreateTemp("", "coffe-snapshot-*.db")
	if err != nil {
		return "", fmt.Errorf("create snapshot path: %w", err)
	}
	path = temporary.Name()
	if err := temporary.Close(); err != nil {
		return "", errors.Join(fmt.Errorf("close snapshot path: %w", err), os.Remove(path))
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("prepare snapshot path: %w", err)
	}
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return "", fmt.Errorf("open source database: %w", err)
	}
	defer func() { err = errors.Join(err, database.Close()) }()
	database.SetMaxOpenConns(1)
	if _, err := database.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return "", fmt.Errorf("checkpoint database: %w", err)
	}
	if _, err := database.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return "", fmt.Errorf("snapshot database: %w", err)
	}
	if err := verify(ctx, path); err != nil {
		return "", err
	}
	return path, nil
}

func verify(ctx context.Context, path string) (err error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open backup database: %w", err)
	}
	defer func() { err = errors.Join(err, database.Close()) }()
	var result string
	if err := database.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("check backup integrity: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("integrity result %q: %w", result, ErrInvalidBackup)
	}
	return nil
}

func checkpoint(ctx context.Context, path string) (err error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open current database: %w", err)
	}
	defer func() { err = errors.Join(err, database.Close()) }()
	if _, err := database.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("checkpoint current database: %w", err)
	}
	return nil
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func cleanPath(path string) (string, error) {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return "", fmt.Errorf("path %q must be absolute: %w", path, ErrInvalidBackup)
	}
	return clean, nil
}
