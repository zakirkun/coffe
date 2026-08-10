import { CloudArrowUp, GitBranch, Robot, ShieldCheck } from "@phosphor-icons/react"

import {
  ActionButton,
  EmptyState,
  InlineError,
  LoadingState,
  Panel,
  StatusBadge,
} from "./primitives"

const navigation = [
  { label: "Deployments", icon: CloudArrowUp },
  { label: "Agents", icon: Robot },
  { label: "Git", icon: GitBranch },
  { label: "Policy", icon: ShieldCheck },
] as const

export const PrimitiveShowcase = () => (
  <div className="showcase">
    <header className="showcase__header">
      <div>
        <span className="meta-label">Coffe desktop system</span>
        <h1>Operational primitives</h1>
      </div>
      <StatusBadge status="ready" />
    </header>

    <main className="showcase__grid">
      <Panel title="Navigation states">
        <nav aria-label="Showcase navigation" className="navigation-sample">
          {navigation.map(({ label, icon: Icon }, index) => (
            <button
              className={index === 0 ? "nav-item nav-item--active" : "nav-item"}
              key={label}
              type="button"
            >
              <Icon aria-hidden="true" size={18} />
              <span>{label}</span>
              {index === 0 ? <span className="nav-item__count">4</span> : null}
            </button>
          ))}
        </nav>
      </Panel>

      <Panel
        actions={<ActionButton variant="primary">Preview</ActionButton>}
        title="Action states"
        variant="approval"
      >
        <div className="showcase__stack">
          <div className="action-cluster">
            <ActionButton variant="primary">Approve</ActionButton>
            <ActionButton>Refresh</ActionButton>
            <ActionButton variant="destructive">Cancel</ActionButton>
          </div>
          <div className="action-cluster">
            <ActionButton busy>Working</ActionButton>
            <ActionButton disabled>Unavailable</ActionButton>
          </div>
          <InlineError message="The operation must be approved before execution." />
        </div>
      </Panel>

      <Panel title="Status language">
        <div className="status-list">
          {["draft", "approved", "running", "succeeded", "failed", "canceled"].map((status) => (
            <StatusBadge key={status} status={status} />
          ))}
        </div>
      </Panel>

      <Panel title="Empty and loading" variant="recessed">
        <div className="showcase__states">
          <EmptyState detail="Create a deployment preview to begin." title="No plans yet" />
          <LoadingState label="Refreshing local state" />
        </div>
      </Panel>

      <Panel title="Dense data table">
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Resource</th>
                <th>Status</th>
                <th>Revision</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>payments / office</td>
                <td>
                  <StatusBadge status="running" />
                </td>
                <td className="technical">71bf4c8</td>
                <td className="technical">14:32:08</td>
              </tr>
              <tr>
                <td>analytics / production-readonly</td>
                <td>
                  <StatusBadge status="succeeded" />
                </td>
                <td className="technical">22ad993</td>
                <td className="technical">14:29:44</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Panel>
    </main>
  </div>
)
