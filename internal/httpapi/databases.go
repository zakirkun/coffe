package httpapi

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"log/slog"
	"net/http"
	"time"
)

type createDatabaseRequest struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Database   string `json:"database"`
	SecretName string `json:"secret_name"`
	SQLitePath string `json:"sqlite_path"`
	TLS        bool   `json:"tls"`
}

type databaseDTO struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Driver     domain.DatabaseDriver `json:"driver"`
	Host       string                `json:"host,omitempty"`
	Port       int                   `json:"port,omitempty"`
	User       string                `json:"user,omitempty"`
	Database   string                `json:"database,omitempty"`
	SQLitePath string                `json:"sqlite_path,omitempty"`
	TLS        bool                  `json:"tls"`
	HasSecret  bool                  `json:"has_secret"`
	CreatedAt  time.Time             `json:"created_at"`
}

type queryDatabaseRequest struct {
	SQL    string `json:"sql"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type queryHistoryDTO struct {
	ID             string    `json:"id"`
	DatabaseID     string    `json:"database_id"`
	Statement      string    `json:"statement"`
	RowCount       int       `json:"row_count"`
	DurationMillis int64     `json:"duration_millis"`
	Truncated      bool      `json:"truncated"`
	Success        bool      `json:"success"`
	CreatedAt      time.Time `json:"created_at"`
}

func (s *Server) createDatabase(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[createDatabaseRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	connection, err := s.app.CreateDatabase(request.Context(), service.CreateDatabaseInput(input))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{
		status: http.StatusCreated, body: toDatabaseDTO(connection),
	})
}

func (s *Server) listDatabases(writer http.ResponseWriter, request *http.Request) {
	connections, err := s.app.ListDatabases(request.Context())
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]databaseDTO, 0, len(connections))
	for _, connection := range connections {
		result = append(result, toDatabaseDTO(connection))
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) queryDatabase(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[queryDatabaseRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result, err := s.app.QueryDatabasePage(
		request.Context(), request.PathValue("databaseID"), service.QueryPageInput(input),
	)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) explainDatabase(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[queryDatabaseRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result, err := s.app.ExplainDatabase(request.Context(), request.PathValue("databaseID"), input.SQL)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) exportDatabaseCSV(writer http.ResponseWriter, request *http.Request) {
	input, err := decodeJSON[queryDatabaseRequest](writer, request)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	content, err := s.app.ExportDatabaseCSV(request.Context(), request.PathValue("databaseID"), input.SQL)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	writer.Header().Set("Content-Disposition", `attachment; filename="query.csv"`)
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(content); err != nil { //nolint:gosec // response is text/csv and formula-leading cells are escaped.
		s.logger.ErrorContext(request.Context(), "CSV response write failed", slog.Any("err", err))
	}
}

func (s *Server) listQueryHistory(writer http.ResponseWriter, request *http.Request) {
	history, err := s.app.ListQueryHistory(request.Context(), request.PathValue("databaseID"), 50)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	result := make([]queryHistoryDTO, 0, len(history))
	for _, record := range history {
		result = append(result, queryHistoryDTO{
			ID: string(record.ID()), DatabaseID: string(record.DatabaseID()), Statement: record.Statement(),
			RowCount: record.RowCount(), DurationMillis: record.DurationMillis(),
			Truncated: record.Truncated(), Success: record.Success(), CreatedAt: record.CreatedAt(),
		})
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: result})
}

func (s *Server) listDatabaseTables(writer http.ResponseWriter, request *http.Request) {
	tables, err := s.app.ListDatabaseTables(request.Context(), request.PathValue("databaseID"))
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: tables})
}

func (s *Server) listDatabaseColumns(writer http.ResponseWriter, request *http.Request) {
	columns, err := s.app.ListDatabaseColumns(
		request.Context(), request.PathValue("databaseID"), request.PathValue("tableName"),
	)
	if err != nil {
		s.fail(request.Context(), writer, err)
		return
	}
	s.respond(request.Context(), writer, response{status: http.StatusOK, body: columns})
}

func toDatabaseDTO(connection domain.DatabaseConnection) databaseDTO {
	return databaseDTO{
		ID: string(connection.ID()), Name: connection.Name().String(), Driver: connection.Driver(),
		Host: connection.Host().String(), Port: connection.Port().Int(), User: connection.User(),
		Database: connection.Database(), SQLitePath: connection.SQLitePath(), TLS: connection.TLS(),
		HasSecret: connection.SecretName() != "", CreatedAt: connection.CreatedAt(),
	}
}
