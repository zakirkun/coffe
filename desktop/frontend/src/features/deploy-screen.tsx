import { useState } from "react"
import type { SwarmNode, SwarmService } from "../api/system-schemas"
import { ActionButton, EmptyState, Panel, StatusBadge } from "../components/primitives"
import { useAction } from "../state/use-action"
import { formatTimestamp, shortID } from "./format"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

export const DeployScreen = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [prompt, setPrompt] = useState("")
  const [nodes, setNodes] = useState<readonly SwarmNode[]>([])
  const [swarmServices, setSwarmServices] = useState<readonly SwarmService[]>([])
  const plan = dashboard.selectedPlan

  const inspectSwarm = (): void => {
    if (dashboard.targetId === null) return
    action.run(
      () =>
        Promise.all([
          services.api.listSwarmNodes(dashboard.targetId ?? ""),
          services.api.listSwarmServices(dashboard.targetId ?? ""),
        ]),
      "Swarm state refreshed",
      ([nextNodes, nextServices]) => {
        setNodes(nextNodes)
        setSwarmServices(nextServices)
      },
    )
  }

  return (
    <div className="screen-grid screen-grid--deploy">
      <Panel title="Deployment command" variant="approval">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            if (dashboard.projectId === null || dashboard.targetId === null) return
            action.run(
              () =>
                services.api.previewDeploy({
                  projectId: dashboard.projectId ?? "",
                  targetId: dashboard.targetId ?? "",
                  prompt,
                }),
              "Deployment preview created",
              (created) => dashboard.selectPlan(created.id),
            )
          }}
        >
          <p className="context-line">
            <span>{dashboard.selectedProject?.name ?? "No project"}</span>
            <span>{dashboard.selectedTarget?.name ?? "No target"}</span>
          </p>
          <div className="field">
            <label className="field__label" htmlFor="deploy-prompt">
              Deployment request
            </label>
            <textarea
              id="deploy-prompt"
              onChange={(event) => setPrompt(event.target.value)}
              placeholder="Deploy the selected project to the selected target"
              required
              value={prompt}
            />
          </div>
          <div className="form-actions">
            <ActionButton
              busy={action.state.kind === "pending"}
              disabled={dashboard.targetId === null}
              type="submit"
              variant="primary"
            >
              Preview
            </ActionButton>
            <ActionButton
              disabled={dashboard.targetId === null}
              onClick={() => {
                if (dashboard.targetId !== null)
                  action.run(
                    () => services.api.previewRollback(dashboard.targetId ?? ""),
                    "Rollback preview created",
                    (created) => dashboard.selectPlan(created.id),
                  )
              }}
            >
              Preview rollback
            </ActionButton>
            <ActionButton disabled={dashboard.targetId === null} onClick={inspectSwarm}>
              Inspect Swarm
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Selected plan">
        {plan === null ? (
          <EmptyState detail="Choose or create a deployment plan." title="No plan selected" />
        ) : (
          <div className="plan-detail">
            <div className="detail-heading">
              <span className="technical">{shortID(plan.id)}</span>
              <StatusBadge status={plan.status} />
            </div>
            <dl className="detail-list">
              <div>
                <dt>Kind</dt>
                <dd>{plan.kind}</dd>
              </div>
              <div>
                <dt>Created</dt>
                <dd>{formatTimestamp(plan.created_at)}</dd>
              </div>
              <div>
                <dt>Revision</dt>
                <dd className="technical">{plan.revision ?? "pending"}</dd>
              </div>
            </dl>
            {plan.commands === undefined || plan.commands.length === 0 ? null : (
              <pre className="log-surface">{plan.commands.join("\n")}</pre>
            )}
            <div className="form-actions">
              <ActionButton
                disabled={plan.status !== "draft"}
                onClick={() => action.run(() => services.api.approvePlan(plan.id), "Plan approved")}
                variant="primary"
              >
                Approve
              </ActionButton>
              <ActionButton
                disabled={plan.status !== "approved"}
                onClick={() =>
                  action.run(() => services.api.executePlan(plan.id), "Execution started")
                }
                variant="primary"
              >
                Execute
              </ActionButton>
              <ActionButton
                disabled={plan.status !== "running"}
                onClick={() =>
                  action.run(() => services.api.cancelPlan(plan.id), "Cancellation requested")
                }
                variant="destructive"
              >
                Cancel
              </ActionButton>
              <ActionButton
                disabled={plan.status !== "failed" && plan.status !== "canceled"}
                onClick={() =>
                  action.run(
                    () => services.api.retryPlan(plan.id),
                    "Retry preview created",
                    (created) => dashboard.selectPlan(created.id),
                  )
                }
              >
                Retry
              </ActionButton>
            </div>
          </div>
        )}
      </Panel>

      <Panel title="Plan activity" variant="recessed">
        {snapshot.events.length === 0 ? (
          <EmptyState detail="Events appear as the plan progresses." title="No activity" />
        ) : (
          <ol className="activity-feed">
            {snapshot.events.map((event) => (
              <li key={event.id}>
                <time dateTime={event.occurred_at}>{formatTimestamp(event.occurred_at)}</time>
                <div>
                  <strong>{event.kind}</strong>
                  <span>{event.message}</span>
                </div>
              </li>
            ))}
          </ol>
        )}
      </Panel>

      <Panel title="Swarm inspection">
        {nodes.length === 0 && swarmServices.length === 0 ? (
          <EmptyState
            detail="Inspect the selected target when it uses Docker Swarm."
            title="No Swarm data"
          />
        ) : (
          <div className="split-list">
            <div>
              <h3>Nodes</h3>
              {nodes.map((node) => (
                <p key={node.id}>
                  <strong>{node.hostname}</strong>
                  <span>
                    {node.status} / {node.availability}
                  </span>
                </p>
              ))}
            </div>
            <div>
              <h3>Services</h3>
              {swarmServices.map((service) => (
                <p key={service.id}>
                  <strong>{service.name}</strong>
                  <span>
                    {service.replicas} / {service.image}
                  </span>
                </p>
              ))}
            </div>
          </div>
        )}
      </Panel>
    </div>
  )
}
