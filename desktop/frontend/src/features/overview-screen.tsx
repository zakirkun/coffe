import { Database, Pulse, Robot, Stack } from "@phosphor-icons/react"

import { EmptyState, Panel, StatusBadge } from "../components/primitives"
import { formatTimestamp, shortID } from "./format"
import type { ScreenProps } from "./types"

export const OverviewScreen = ({ snapshot }: ScreenProps) => {
  const metrics = [
    { label: "Projects", value: snapshot.projects.length, icon: Stack },
    {
      label: "Active plans",
      value: snapshot.plans.filter((plan) => plan.status === "running").length,
      icon: Pulse,
    },
    {
      label: "Agents",
      value: snapshot.agents.filter((agent) => agent.status === "running").length,
      icon: Robot,
    },
    { label: "Databases", value: snapshot.databases.length, icon: Database },
  ] as const

  return (
    <div className="screen-grid">
      <section className="metric-grid" aria-label="Workspace metrics">
        {metrics.map(({ label, value, icon: Icon }) => (
          <div className="metric" key={label}>
            <Icon aria-hidden="true" size={18} />
            <span>{label}</span>
            <strong>{value}</strong>
          </div>
        ))}
      </section>

      <Panel title="Recent deployment plans">
        {snapshot.plans.length === 0 ? (
          <EmptyState
            detail="Preview a deployment to create the first plan."
            title="No plans yet"
          />
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Plan</th>
                  <th>Kind</th>
                  <th>Status</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {snapshot.plans.slice(0, 8).map((plan) => (
                  <tr key={plan.id}>
                    <td className="technical">{shortID(plan.id)}</td>
                    <td>{plan.kind.replaceAll("_", " ")}</td>
                    <td>
                      <StatusBadge status={plan.status} />
                    </td>
                    <td>{formatTimestamp(plan.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>

      <Panel title="Latest audit activity" variant="recessed">
        {snapshot.audit.length === 0 ? (
          <EmptyState detail="Operator actions will appear here." title="No audit entries" />
        ) : (
          <ol className="activity-feed">
            {snapshot.audit.slice(0, 10).map((entry) => (
              <li key={entry.id}>
                <time dateTime={entry.occurred_at}>{formatTimestamp(entry.occurred_at)}</time>
                <div>
                  <strong>{entry.action}</strong>
                  <span>{entry.message}</span>
                </div>
                <StatusBadge status={entry.outcome} />
              </li>
            ))}
          </ol>
        )}
      </Panel>
    </div>
  )
}
