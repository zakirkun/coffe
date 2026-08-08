package sqlite

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) CreateDatabaseConnection(ctx context.Context, connection domain.DatabaseConnection) error {
	const query = `INSERT INTO database_connections (
		id, name, driver, host, port, user, database_name, secret_name, sqlite_path, tls, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, query,
		connection.ID(), connection.Name().String(), connection.Driver(),
		connection.Host().String(), connection.Port().Int(), connection.User(), connection.Database(),
		connection.SecretName(), connection.SQLitePath(), connection.TLS(), formatTime(connection.CreatedAt()),
	); err != nil {
		return fmt.Errorf("create database connection: %w", err)
	}
	return nil
}

func (s *Store) GetDatabaseConnection(
	ctx context.Context,
	id domain.DatabaseID,
) (domain.DatabaseConnection, error) {
	const query = `SELECT id, name, driver, host, port, user, database_name, secret_name, sqlite_path, tls, created_at
		FROM database_connections WHERE id = ?`
	connection, err := scanDatabaseConnection(s.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DatabaseConnection{}, fmt.Errorf("database connection %s: %w", id, service.ErrNotFound)
	}
	if err != nil {
		return domain.DatabaseConnection{}, err
	}
	return connection, nil
}

func (s *Store) ListDatabaseConnections(
	ctx context.Context,
) (connections []domain.DatabaseConnection, err error) {
	const query = `SELECT id, name, driver, host, port, user, database_name, secret_name, sqlite_path, tls, created_at
		FROM database_connections ORDER BY name, id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query database connections: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		connection, err := scanDatabaseConnection(rows)
		if err != nil {
			return nil, err
		}
		connections = append(connections, connection)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate database connections: %w", err)
	}
	return connections, nil
}

func scanDatabaseConnection(row rowScanner) (domain.DatabaseConnection, error) {
	var id, rawName, rawDriver, rawHost, rawUser, database, secretName, sqlitePath, rawCreatedAt string
	var rawPort int
	var tls bool
	if err := row.Scan(
		&id, &rawName, &rawDriver, &rawHost, &rawPort, &rawUser,
		&database, &secretName, &sqlitePath, &tls, &rawCreatedAt,
	); err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("scan database connection: %w", err)
	}
	name, err := domain.NewName(rawName)
	if err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("restore database name: %w", err)
	}
	var host domain.Host
	var port domain.Port
	if domain.DatabaseDriver(rawDriver) == domain.DatabaseMySQL {
		host, err = domain.NewHost(rawHost)
		if err != nil {
			return domain.DatabaseConnection{}, fmt.Errorf("restore database host: %w", err)
		}
		port, err = domain.NewPort(rawPort)
		if err != nil {
			return domain.DatabaseConnection{}, fmt.Errorf("restore database port: %w", err)
		}
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("parse database created_at: %w", err)
	}
	connection, err := domain.NewDatabaseConnection(domain.DatabaseValues{
		ID: domain.DatabaseID(id), Name: name, Driver: domain.DatabaseDriver(rawDriver),
		Host: host, Port: port, User: rawUser, Database: database, SecretName: secretName,
		SQLitePath: sqlitePath, TLS: tls, CreatedAt: createdAt,
	})
	if err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("restore database connection: %w", err)
	}
	return connection, nil
}
