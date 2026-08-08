package service

import (
	"coffe/internal/domain"
	"context"
	"fmt"
	"strings"
)

type CreateDatabaseInput struct {
	Name       string
	Driver     string
	Host       string
	Port       int
	User       string
	Database   string
	SecretName string
	SQLitePath string
	TLS        bool
}

func (a *App) CreateDatabase(
	ctx context.Context,
	input CreateDatabaseInput,
) (domain.DatabaseConnection, error) {
	name, err := domain.NewName(input.Name)
	if err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("database name: %w", err)
	}
	driver := domain.DatabaseDriver(strings.TrimSpace(input.Driver))
	if !driver.Valid() {
		return domain.DatabaseConnection{}, fmt.Errorf("database driver: %w", ErrInvalidInput)
	}
	rawID, err := a.ids.New()
	if err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("new database id: %w", err)
	}
	values := domain.DatabaseValues{
		ID: domain.DatabaseID(rawID), Name: name, Driver: driver,
		User: input.User, Database: input.Database, SecretName: input.SecretName,
		SQLitePath: input.SQLitePath, TLS: input.TLS, CreatedAt: a.clock.Now(),
	}
	if driver == domain.DatabaseMySQL {
		values.Host, err = domain.NewHost(input.Host)
		if err != nil {
			return domain.DatabaseConnection{}, fmt.Errorf("database host: %w", err)
		}
		values.Port, err = domain.NewPort(input.Port)
		if err != nil {
			return domain.DatabaseConnection{}, fmt.Errorf("database port: %w", err)
		}
	}
	connection, err := domain.NewDatabaseConnection(values)
	if err != nil {
		return domain.DatabaseConnection{}, err
	}
	if err := a.repository.CreateDatabaseConnection(ctx, connection); err != nil {
		return domain.DatabaseConnection{}, fmt.Errorf("create database connection: %w", err)
	}
	return connection, nil
}

func (a *App) ListDatabases(ctx context.Context) ([]domain.DatabaseConnection, error) {
	connections, err := a.repository.ListDatabaseConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("list database connections: %w", err)
	}
	return connections, nil
}

func (a *App) QueryDatabase(
	ctx context.Context,
	rawID string,
	query string,
) (domain.QueryResult, error) {
	return a.QueryDatabasePage(ctx, rawID, QueryPageInput{SQL: query})
}

func (a *App) ListDatabaseTables(ctx context.Context, rawID string) ([]domain.DatabaseTable, error) {
	connection, password, err := a.databaseAccess(ctx, rawID)
	if err != nil {
		return nil, err
	}
	tables, err := a.databases.Tables(ctx, DatabaseAccess{Connection: connection, Password: password})
	if err != nil {
		return nil, fmt.Errorf("list database tables: %w", err)
	}
	return tables, nil
}

func (a *App) ListDatabaseColumns(
	ctx context.Context,
	rawID string,
	tableName string,
) ([]domain.DatabaseColumn, error) {
	connection, password, err := a.databaseAccess(ctx, rawID)
	if err != nil {
		return nil, err
	}
	columns, err := a.databases.Columns(
		ctx, DatabaseAccess{Connection: connection, Password: password}, tableName,
	)
	if err != nil {
		return nil, fmt.Errorf("list database columns: %w", err)
	}
	return columns, nil
}

func (a *App) databaseAccess(
	ctx context.Context,
	rawID string,
) (domain.DatabaseConnection, string, error) {
	id := domain.DatabaseID(strings.TrimSpace(rawID))
	if id == "" || string(id) != rawID {
		return domain.DatabaseConnection{}, "", fmt.Errorf("database id: %w", ErrInvalidInput)
	}
	connection, err := a.repository.GetDatabaseConnection(ctx, id)
	if err != nil {
		return domain.DatabaseConnection{}, "", fmt.Errorf("get database connection: %w", err)
	}
	if connection.SecretName() == "" {
		return connection, "", nil
	}
	password, err := a.secrets.Get(ctx, connection.SecretName())
	if err != nil {
		return domain.DatabaseConnection{}, "", fmt.Errorf("get database credential: %w", err)
	}
	return connection, password, nil
}
