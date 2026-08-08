package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var ErrInvalidDatabase = errors.New("domain: invalid database connection")

type (
	DatabaseID     string
	DatabaseDriver string
)

const (
	DatabaseSQLite DatabaseDriver = "sqlite"
	DatabaseMySQL  DatabaseDriver = "mysql"
)

func (d DatabaseDriver) Valid() bool { return d == DatabaseSQLite || d == DatabaseMySQL }

type DatabaseConnection struct {
	createdAt  time.Time
	id         DatabaseID
	name       Name
	driver     DatabaseDriver
	host       Host
	port       Port
	user       string
	database   string
	secretName string
	sqlitePath string
	tls        bool
}

type DatabaseValues struct {
	ID         DatabaseID
	Name       Name
	Driver     DatabaseDriver
	Host       Host
	Port       Port
	User       string
	Database   string
	SecretName string
	SQLitePath string
	TLS        bool
	CreatedAt  time.Time
}

func NewDatabaseConnection(values DatabaseValues) (DatabaseConnection, error) {
	if values.ID == "" || values.Name.String() == "" || !values.Driver.Valid() || values.CreatedAt.IsZero() {
		return DatabaseConnection{}, fmt.Errorf("database identity: %w", ErrInvalidDatabase)
	}
	if values.Driver == DatabaseSQLite {
		path := filepath.Clean(strings.TrimSpace(values.SQLitePath))
		if !filepath.IsAbs(path) {
			return DatabaseConnection{}, fmt.Errorf("sqlite path: %w", ErrInvalidDatabase)
		}
		return DatabaseConnection{
			id: values.ID, name: values.Name, driver: values.Driver,
			sqlitePath: path, createdAt: values.CreatedAt.UTC(),
		}, nil
	}
	user := strings.TrimSpace(values.User)
	database := strings.TrimSpace(values.Database)
	secretName := strings.TrimSpace(values.SecretName)
	if values.Host.String() == "" || values.Port == 0 || user == "" || database == "" || secretName == "" ||
		hasControl(user) || hasControl(database) || hasControl(secretName) {
		return DatabaseConnection{}, fmt.Errorf("mysql connection: %w", ErrInvalidDatabase)
	}
	return DatabaseConnection{
		id: values.ID, name: values.Name, driver: values.Driver,
		host: values.Host, port: values.Port, user: user, database: database,
		secretName: secretName, tls: values.TLS, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func (c DatabaseConnection) ID() DatabaseID         { return c.id }
func (c DatabaseConnection) Name() Name             { return c.name }
func (c DatabaseConnection) Driver() DatabaseDriver { return c.driver }
func (c DatabaseConnection) Host() Host             { return c.host }
func (c DatabaseConnection) Port() Port             { return c.port }
func (c DatabaseConnection) User() string           { return c.user }
func (c DatabaseConnection) Database() string       { return c.database }
func (c DatabaseConnection) SecretName() string     { return c.secretName }
func (c DatabaseConnection) SQLitePath() string     { return c.sqlitePath }
func (c DatabaseConnection) TLS() bool              { return c.tls }
func (c DatabaseConnection) CreatedAt() time.Time   { return c.createdAt }

type QueryResult struct {
	Columns   []string   `json:"columns"`
	Rows      [][]string `json:"rows"`
	Truncated bool       `json:"truncated"`
}

type DatabaseTable struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type DatabaseColumn struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Nullable bool   `json:"nullable"`
}
