import { EmptyState, Panel } from "../components/primitives"

export type QueryDisplay = Readonly<{
  columns: readonly Readonly<{ id: string; label: string }>[]
  rows: readonly Readonly<{
    id: string
    cells: readonly Readonly<{ id: string; value: string }>[]
  }>[]
  truncated: boolean
}>

export const DatabaseResultPanel = ({ result }: Readonly<{ result: QueryDisplay | null }>) => (
  <Panel title="Query result">
    {result === null ? (
      <EmptyState detail="Run a capped read-only statement." title="No result" />
    ) : (
      <>
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                {result.columns.map((column) => (
                  <th key={column.id}>{column.label}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {result.rows.map((row) => (
                <tr key={row.id}>
                  {row.cells.map((cell) => (
                    <td className="technical" key={cell.id}>
                      {cell.value}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {result.truncated ? (
          <p className="form-status">Result truncated at the configured cap.</p>
        ) : null}
      </>
    )}
  </Panel>
)
