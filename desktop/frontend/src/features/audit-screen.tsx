import { useState } from "react"

import { ActionButton, EmptyState, InlineError, Panel, StatusBadge } from "../components/primitives"
import { useAction } from "../state/use-action"
import { formatTimestamp, shortID } from "./format"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

export const AuditScreen = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [exportError, setExportError] = useState<string | null>(null)

  const exportAudit = (): void => {
    services.desktop
      .choosePath("audit_export")
      .then((path) => {
        setExportError(null)
        if (path !== "")
          action.run(() => services.desktop.exportAudit(path), "Audit ledger exported")
      })
      .catch((error: unknown) => {
        setExportError(error instanceof Error ? error.message : "Unable to export audit ledger")
      })
  }

  return (
    <Panel
      actions={<ActionButton onClick={exportAudit}>Export CSV</ActionButton>}
      title="Redacted audit ledger"
    >
      {exportError === null ? null : <InlineError message={exportError} />}
      <ActionStatus state={action.state} />
      {snapshot.audit.length === 0 ? (
        <EmptyState
          detail="Policy and operator decisions will appear here."
          title="No audit entries"
        />
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Category</th>
                <th>Action</th>
                <th>Resource</th>
                <th>Outcome</th>
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              {snapshot.audit.map((entry) => (
                <tr key={entry.id}>
                  <td>{formatTimestamp(entry.occurred_at)}</td>
                  <td>{entry.category}</td>
                  <td>{entry.action}</td>
                  <td className="technical">{shortID(entry.resource_id)}</td>
                  <td>
                    <StatusBadge status={entry.outcome} />
                  </td>
                  <td>{entry.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  )
}
