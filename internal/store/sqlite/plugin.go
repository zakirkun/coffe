package sqlite

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *Store) SavePlugin(ctx context.Context, plugin domain.Plugin) error {
	methods, err := json.Marshal(plugin.Methods())
	if err != nil {
		return fmt.Errorf("encode plugin methods: %w", err)
	}
	const query = `INSERT INTO plugins (
		id, name, executable_path, sha256, methods_json, enabled, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name, executable_path = excluded.executable_path,
		sha256 = excluded.sha256, methods_json = excluded.methods_json, enabled = excluded.enabled`
	if _, err := s.db.ExecContext(ctx, query,
		plugin.ID(), plugin.Name().String(), plugin.ExecutablePath(), plugin.SHA256(),
		string(methods), plugin.Enabled(), formatTime(plugin.CreatedAt()),
	); err != nil {
		return fmt.Errorf("save plugin: %w", err)
	}
	return nil
}

func (s *Store) GetPlugin(ctx context.Context, id domain.PluginID) (domain.Plugin, error) {
	const query = `SELECT id, name, executable_path, sha256, methods_json, enabled, created_at
		FROM plugins WHERE id = ?`
	plugin, err := scanPlugin(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Plugin{}, service.ErrNotFound
	}
	if err != nil {
		return domain.Plugin{}, err
	}
	return plugin, nil
}

func (s *Store) ListPlugins(ctx context.Context) (plugins []domain.Plugin, err error) {
	const query = `SELECT id, name, executable_path, sha256, methods_json, enabled, created_at
		FROM plugins ORDER BY name, id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query plugins: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		plugin, err := scanPlugin(rows)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, plugin)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plugins: %w", err)
	}
	return plugins, nil
}

func scanPlugin(row rowScanner) (domain.Plugin, error) {
	var id, rawName, executablePath, hash, rawMethods, rawCreatedAt string
	var enabled bool
	if err := row.Scan(
		&id, &rawName, &executablePath, &hash, &rawMethods, &enabled, &rawCreatedAt,
	); err != nil {
		return domain.Plugin{}, fmt.Errorf("scan plugin: %w", err)
	}
	name, err := domain.NewName(rawName)
	if err != nil {
		return domain.Plugin{}, err
	}
	var methods []string
	if err := json.Unmarshal([]byte(rawMethods), &methods); err != nil {
		return domain.Plugin{}, fmt.Errorf("decode plugin methods: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.Plugin{}, fmt.Errorf("parse plugin created_at: %w", err)
	}
	plugin, err := domain.NewPlugin(domain.PluginValues{
		ID: domain.PluginID(id), Name: name, ExecutablePath: executablePath,
		SHA256: hash, Methods: methods, Enabled: enabled, CreatedAt: createdAt,
	})
	if err != nil {
		return domain.Plugin{}, fmt.Errorf("restore plugin: %w", err)
	}
	return plugin, nil
}
