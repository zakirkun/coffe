package backup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"filippo.io/age"
)

type ImportOptions struct {
	DatabasePath string
	InputPath    string
	Passphrase   string
}

func Import(ctx context.Context, options ImportOptions) (err error) {
	if err := validatePassphrase(options.Passphrase); err != nil {
		return err
	}
	databasePath, err := cleanPath(options.DatabasePath)
	if err != nil {
		return err
	}
	inputPath, err := cleanPath(options.InputPath)
	if err != nil {
		return err
	}
	if databasePath == inputPath {
		return fmt.Errorf("backup input equals database path: %w", ErrInvalidBackup)
	}
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o700); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	temporary, err := decryptToTemp(inputPath, databasePath, options.Passphrase)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, removeIfExists(temporary)) }()
	if err := verify(ctx, temporary); err != nil {
		return err
	}
	return replaceDatabase(ctx, databasePath, temporary)
}

func decryptToTemp(inputPath, databasePath, passphrase string) (path string, err error) {
	source, err := os.Open(inputPath) //nolint:gosec // inputPath is validated as an absolute user-selected backup file.
	if err != nil {
		return "", fmt.Errorf("open encrypted backup: %w", err)
	}
	defer func() { err = errors.Join(err, source.Close()) }()
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return "", fmt.Errorf("create backup identity: %w", err)
	}
	decrypted, err := age.Decrypt(source, identity)
	if err != nil {
		return "", fmt.Errorf("decrypt backup: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(databasePath), ".coffe-import-*.db")
	if err != nil {
		return "", fmt.Errorf("create import file: %w", err)
	}
	path = temporary.Name()
	succeeded := false
	defer func() {
		err = errors.Join(err, temporary.Close())
		if !succeeded {
			err = errors.Join(err, removeIfExists(path))
		}
	}()
	written, err := io.Copy(temporary, io.LimitReader(decrypted, maxBackupBytes+1))
	if err != nil {
		return "", fmt.Errorf("write decrypted backup: %w", err)
	}
	if written > maxBackupBytes {
		return "", fmt.Errorf("backup exceeds size limit: %w", ErrInvalidBackup)
	}
	if err := temporary.Chmod(0o600); err != nil {
		return "", fmt.Errorf("protect import file: %w", err)
	}
	succeeded = true
	return path, nil
}

func replaceDatabase(ctx context.Context, databasePath, temporary string) error {
	backupPath := databasePath + ".before-import-" + strconv.FormatInt(time.Now().Unix(), 10)
	exists := true
	if _, err := os.Stat(databasePath); errors.Is(err, os.ErrNotExist) {
		exists = false
	} else if err != nil {
		return fmt.Errorf("inspect current database: %w", err)
	}
	if exists {
		if err := checkpoint(ctx, databasePath); err != nil {
			return err
		}
		if err := os.Rename(databasePath, backupPath); err != nil {
			return fmt.Errorf("preserve current database: %w", err)
		}
	}
	if err := os.Rename(temporary, databasePath); err != nil {
		if exists {
			return errors.Join(err, os.Rename(backupPath, databasePath))
		}
		return fmt.Errorf("install imported database: %w", err)
	}
	return errors.Join(removeIfExists(databasePath+"-wal"), removeIfExists(databasePath+"-shm"))
}
