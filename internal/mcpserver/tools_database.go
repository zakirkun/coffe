package mcpserver

import (
	"coffe/internal/domain"
	"coffe/internal/service"
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listDatabasesInput struct{}

type databaseOutput struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Driver   domain.DatabaseDriver `json:"driver"`
	Database string                `json:"database,omitempty"`
}

type databasesOutput struct {
	Databases []databaseOutput `json:"databases"`
}

type databaseInput struct {
	DatabaseID string `json:"database_id" jsonschema:"registered database connection identifier"`
}

type databaseQueryInput struct {
	DatabaseID string `json:"database_id" jsonschema:"registered database connection identifier"`
	SQL        string `json:"sql" jsonschema:"single read-only SELECT, SHOW, DESCRIBE, or EXPLAIN statement"`
	Offset     int    `json:"offset,omitempty" jsonschema:"zero-based rows to skip"`
	Limit      int    `json:"limit,omitempty" jsonschema:"page size capped by Coffe"`
}

type databaseColumnsInput struct {
	DatabaseID string `json:"database_id" jsonschema:"registered database connection identifier"`
	Table      string `json:"table" jsonschema:"table name"`
}

type (
	databaseTablesOutput struct {
		Tables []domain.DatabaseTable `json:"tables"`
	}
	databaseColumnsOutput struct {
		Columns []domain.DatabaseColumn `json:"columns"`
	}
	databaseQueryOutput struct {
		Result domain.QueryResult `json:"result"`
	}
	databaseHistoryOutput struct {
		History []queryHistoryRecord `json:"history"`
	}
)

type queryHistoryRecord struct {
	DatabaseID     string    `json:"database_id"`
	Statement      string    `json:"statement"`
	RowCount       int       `json:"row_count"`
	DurationMillis int64     `json:"duration_millis"`
	Truncated      bool      `json:"truncated"`
	Success        bool      `json:"success"`
	CreatedAt      time.Time `json:"created_at"`
}

func registerDatabaseTools(server *mcp.Server, app Application) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_databases", Description: "List registered database connections without credentials",
	}, listDatabasesHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "database_tables", Description: "List tables and views through a read-only connection",
	}, databaseTablesHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "database_columns", Description: "List columns for one database table",
	}, databaseColumnsHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "database_query_readonly", Description: "Execute one capped read-only database query",
	}, databaseQueryHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "database_explain", Description: "Run a read-only EXPLAIN for one statement",
	}, databaseExplainHandler(app))
	mcp.AddTool(server, &mcp.Tool{
		Name: "database_query_history", Description: "List local read-only query history",
	}, databaseHistoryHandler(app))
}

func listDatabasesHandler(app Application) mcp.ToolHandlerFor[listDatabasesInput, databasesOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		_ listDatabasesInput,
	) (*mcp.CallToolResult, databasesOutput, error) {
		connections, err := app.ListDatabases(ctx)
		if err != nil {
			return nil, databasesOutput{}, fmt.Errorf("list databases: %w", err)
		}
		output := make([]databaseOutput, 0, len(connections))
		for _, connection := range connections {
			output = append(output, databaseOutput{
				ID: string(connection.ID()), Name: connection.Name().String(),
				Driver: connection.Driver(), Database: connection.Database(),
			})
		}
		return nil, databasesOutput{Databases: output}, nil
	}
}

func databaseTablesHandler(app Application) mcp.ToolHandlerFor[databaseInput, databaseTablesOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input databaseInput,
	) (*mcp.CallToolResult, databaseTablesOutput, error) {
		tables, err := app.ListDatabaseTables(ctx, input.DatabaseID)
		if err != nil {
			return nil, databaseTablesOutput{}, fmt.Errorf("list database tables: %w", err)
		}
		return nil, databaseTablesOutput{Tables: tables}, nil
	}
}

func databaseColumnsHandler(app Application) mcp.ToolHandlerFor[databaseColumnsInput, databaseColumnsOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input databaseColumnsInput,
	) (*mcp.CallToolResult, databaseColumnsOutput, error) {
		columns, err := app.ListDatabaseColumns(ctx, input.DatabaseID, input.Table)
		if err != nil {
			return nil, databaseColumnsOutput{}, fmt.Errorf("list database columns: %w", err)
		}
		return nil, databaseColumnsOutput{Columns: columns}, nil
	}
}

func databaseQueryHandler(app Application) mcp.ToolHandlerFor[databaseQueryInput, databaseQueryOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input databaseQueryInput,
	) (*mcp.CallToolResult, databaseQueryOutput, error) {
		result, err := app.QueryDatabasePage(ctx, input.DatabaseID, service.QueryPageInput{
			SQL: input.SQL, Offset: input.Offset, Limit: input.Limit,
		})
		if err != nil {
			return nil, databaseQueryOutput{}, fmt.Errorf("query database: %w", err)
		}
		return nil, databaseQueryOutput{Result: result}, nil
	}
}

func databaseExplainHandler(app Application) mcp.ToolHandlerFor[databaseQueryInput, databaseQueryOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input databaseQueryInput,
	) (*mcp.CallToolResult, databaseQueryOutput, error) {
		result, err := app.ExplainDatabase(ctx, input.DatabaseID, input.SQL)
		if err != nil {
			return nil, databaseQueryOutput{}, fmt.Errorf("explain database query: %w", err)
		}
		return nil, databaseQueryOutput{Result: result}, nil
	}
}

func databaseHistoryHandler(app Application) mcp.ToolHandlerFor[databaseInput, databaseHistoryOutput] {
	return func(
		ctx context.Context,
		_ *mcp.CallToolRequest,
		input databaseInput,
	) (*mcp.CallToolResult, databaseHistoryOutput, error) {
		history, err := app.ListQueryHistory(ctx, input.DatabaseID, 50)
		if err != nil {
			return nil, databaseHistoryOutput{}, fmt.Errorf("list database query history: %w", err)
		}
		output := make([]queryHistoryRecord, 0, len(history))
		for _, record := range history {
			output = append(output, queryHistoryRecord{
				DatabaseID: string(record.DatabaseID()), Statement: record.Statement(),
				RowCount: record.RowCount(), DurationMillis: record.DurationMillis(),
				Truncated: record.Truncated(), Success: record.Success(), CreatedAt: record.CreatedAt(),
			})
		}
		return nil, databaseHistoryOutput{History: output}, nil
	}
}
