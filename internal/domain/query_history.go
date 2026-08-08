package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidQueryHistory = errors.New("domain: invalid query history")

type QueryHistoryID string

type QueryHistory struct {
	createdAt      time.Time
	id             QueryHistoryID
	databaseID     DatabaseID
	statement      string
	rowCount       int
	durationMillis int64
	truncated      bool
	success        bool
}

type QueryHistoryValues struct {
	ID             QueryHistoryID
	DatabaseID     DatabaseID
	Statement      string
	RowCount       int
	DurationMillis int64
	Truncated      bool
	Success        bool
	CreatedAt      time.Time
}

func NewQueryHistory(values QueryHistoryValues) (QueryHistory, error) {
	statement := strings.TrimSpace(values.Statement)
	if values.ID == "" || values.DatabaseID == "" || statement == "" || len(statement) > 4096 ||
		values.RowCount < 0 || values.DurationMillis < 0 || values.CreatedAt.IsZero() || hasControl(statement) {
		return QueryHistory{}, fmt.Errorf("query history: %w", ErrInvalidQueryHistory)
	}
	return QueryHistory{
		id: values.ID, databaseID: values.DatabaseID, statement: statement,
		rowCount: values.RowCount, durationMillis: values.DurationMillis,
		truncated: values.Truncated, success: values.Success, createdAt: values.CreatedAt.UTC(),
	}, nil
}

func (h QueryHistory) ID() QueryHistoryID     { return h.id }
func (h QueryHistory) DatabaseID() DatabaseID { return h.databaseID }
func (h QueryHistory) Statement() string      { return h.statement }
func (h QueryHistory) RowCount() int          { return h.rowCount }
func (h QueryHistory) DurationMillis() int64  { return h.durationMillis }
func (h QueryHistory) Truncated() bool        { return h.truncated }
func (h QueryHistory) Success() bool          { return h.success }
func (h QueryHistory) CreatedAt() time.Time   { return h.createdAt }
