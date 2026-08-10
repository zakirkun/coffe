import { useState } from "react"

import type {
  DatabaseColumn,
  DatabaseTable,
  QueryHistory,
  QueryResult,
} from "../api/operation-schemas"
import { SelectField } from "../components/forms"
import { ActionButton, InlineError, Panel, StatusBadge } from "../components/primitives"
import { reconcileSelection, selectionChannel, useAction } from "../state/use-action"
import { DatabaseResultPanel, type QueryDisplay } from "./database-result"
import { formatTimestamp } from "./format"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

type DatabaseData = Readonly<{
  result: QueryDisplay | null
  tables: readonly DatabaseTable[]
  columns: readonly DatabaseColumn[]
  history: readonly QueryHistory[]
}>

const emptyDatabaseData: DatabaseData = {
  result: null,
  tables: [],
  columns: [],
  history: [],
}

const displayResult = (result: QueryResult): QueryDisplay => ({
  columns: result.columns.map((label) => ({ id: crypto.randomUUID(), label })),
  rows: result.rows.map((row) => ({
    id: crypto.randomUUID(),
    cells: row.map((value) => ({ id: crypto.randomUUID(), value })),
  })),
  truncated: result.truncated,
})

export const DatabaseQuery = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [requestedDatabaseId, setDatabaseId] = useState(snapshot.databases[0]?.id ?? "")
  const databaseId = reconcileSelection(requestedDatabaseId, snapshot.databases)
  const [sql, setSQL] = useState("")
  const [databaseStates, setDatabaseStates] = useState<Readonly<Record<string, DatabaseData>>>({})
  const { result, tables, columns, history } = databaseStates[databaseId] ?? emptyDatabaseData
  const [exportError, setExportError] = useState<string | null>(null)

  const runQuery = (explain: boolean): void => {
    if (databaseId === "") return
    action.run(
      () =>
        explain
          ? services.api.explainDatabase(databaseId, sql)
          : services.api.queryDatabase({ databaseId, sql, offset: 0, limit: 200 }),
      explain ? "Explain plan loaded" : "Query completed",
      (nextResult) =>
        setDatabaseStates((current) => ({
          ...current,
          [databaseId]: {
            ...(current[databaseId] ?? emptyDatabaseData),
            result: displayResult(nextResult),
          },
        })),
      selectionChannel("database-result", databaseId),
    )
  }

  const inspectSchema = (): void => {
    if (databaseId === "") return
    action.run(
      () =>
        Promise.all([
          services.api.listDatabaseTables(databaseId),
          services.api.listQueryHistory(databaseId),
        ]),
      "Schema and history loaded",
      ([nextTables, nextHistory]) => {
        setDatabaseStates((current) => ({
          ...current,
          [databaseId]: {
            ...(current[databaseId] ?? emptyDatabaseData),
            tables: nextTables,
            history: nextHistory,
          },
        }))
      },
      selectionChannel("database-inspection", databaseId),
    )
  }

  const exportQuery = (): void => {
    if (databaseId === "" || sql === "") return
    services.desktop
      .choosePath("query_export")
      .then((path) => {
        if (path !== "") {
          action.run(
            () => services.desktop.exportQuery(databaseId, sql, path),
            "Query exported",
            undefined,
            selectionChannel("database-export", databaseId),
          )
        }
      })
      .catch((error: unknown) =>
        setExportError(error instanceof Error ? error.message : "Unable to export query"),
      )
  }

  return (
    <>
      <Panel title="Read-only query" variant="approval">
        <div className="form-stack">
          <SelectField
            label="Database"
            onChange={(event) => setDatabaseId(event.target.value)}
            value={databaseId}
          >
            <option value="">Select database</option>
            {snapshot.databases.map((database) => (
              <option key={database.id} value={database.id}>
                {database.name} ({database.driver})
              </option>
            ))}
          </SelectField>
          <div className="field">
            <label className="field__label" htmlFor="database-query">
              SQL statement
            </label>
            <textarea
              className="technical"
              id="database-query"
              onChange={(event) => setSQL(event.target.value)}
              placeholder="SELECT ..."
              value={sql}
            />
          </div>
          <div className="form-actions">
            <ActionButton onClick={() => runQuery(false)} variant="primary">
              Run query
            </ActionButton>
            <ActionButton onClick={() => runQuery(true)}>Explain</ActionButton>
            <ActionButton onClick={inspectSchema}>Inspect schema</ActionButton>
            <ActionButton onClick={exportQuery}>Export CSV</ActionButton>
          </div>
          {exportError === null ? null : <InlineError message={exportError} />}
          <ActionStatus state={action.state} />
        </div>
      </Panel>

      <DatabaseResultPanel result={result} />

      <Panel title="Schema and history" variant="recessed">
        <div className="split-list">
          <div>
            <h3>Tables</h3>
            {tables.map((table) => (
              <button
                className="list-button"
                key={table.name}
                onClick={() =>
                  action.run(
                    () => services.api.listDatabaseColumns(databaseId, table.name),
                    "Columns loaded",
                    (nextColumns) =>
                      setDatabaseStates((current) => ({
                        ...current,
                        [databaseId]: {
                          ...(current[databaseId] ?? emptyDatabaseData),
                          columns: nextColumns,
                        },
                      })),
                    selectionChannel("database-columns", databaseId),
                  )
                }
                type="button"
              >
                <strong>{table.name}</strong>
                <span>{table.type}</span>
              </button>
            ))}
          </div>
          <div>
            <h3>Columns</h3>
            {columns.map((column) => (
              <p key={column.name}>
                <strong>{column.name}</strong>
                <span>
                  {column.data_type} / {column.nullable ? "nullable" : "required"}
                </span>
              </p>
            ))}
          </div>
        </div>
        {history.length === 0 ? null : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Statement</th>
                  <th>Rows</th>
                  <th>Status</th>
                  <th>Run</th>
                </tr>
              </thead>
              <tbody>
                {history.slice(0, 10).map((entry) => (
                  <tr key={entry.id}>
                    <td className="technical">{entry.statement}</td>
                    <td>{entry.row_count}</td>
                    <td>
                      <StatusBadge status={entry.success ? "success" : "failed"} />
                    </td>
                    <td>{formatTimestamp(entry.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>
    </>
  )
}
