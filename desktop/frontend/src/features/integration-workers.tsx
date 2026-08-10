import type { Worker } from "../api/system-schemas"
import { ActionButton, EmptyState, Panel, StatusBadge } from "../components/primitives"
import { shortID } from "./format"

type IntegrationWorkersProps = Readonly<{
  workers: readonly Worker[]
  onHeartbeat: (workerId: string) => void
}>

export const IntegrationWorkers = ({ workers, onHeartbeat }: IntegrationWorkersProps) => (
  <Panel title="Registered workers">
    {workers.length === 0 ? (
      <EmptyState detail="Map a target to an authenticated HTTPS worker." title="No workers" />
    ) : (
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Worker</th>
              <th>Target</th>
              <th>Version</th>
              <th>Status</th>
              <th>Probe</th>
            </tr>
          </thead>
          <tbody>
            {workers.map((worker) => (
              <tr key={worker.id}>
                <td>{worker.name}</td>
                <td className="technical">{shortID(worker.target_id)}</td>
                <td>{worker.version ?? "unknown"}</td>
                <td>
                  <StatusBadge status={worker.enabled ? "ready" : "stopped"} />
                </td>
                <td>
                  <ActionButton onClick={() => onHeartbeat(worker.id)}>Heartbeat</ActionButton>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )}
  </Panel>
)
