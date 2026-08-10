import { useState } from "react"

import type { GitStatus } from "../api/operation-schemas"
import { SelectField } from "../components/forms"
import { ActionButton, EmptyState, Panel, StatusBadge } from "../components/primitives"
import { useAction } from "../state/use-action"
import { formatTimestamp, shortID } from "./format"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

export const GitScreen = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [status, setStatus] = useState<GitStatus | null>(null)
  const [diff, setDiff] = useState("")
  const [branches, setBranches] = useState<readonly string[]>([])
  const [branch, setBranch] = useState("main")

  const inspect = (): void => {
    const activeProjectId = dashboard.projectId
    if (activeProjectId === null) return
    action.run(
      () =>
        Promise.all([
          services.api.gitStatus(activeProjectId),
          services.api.gitDiff(activeProjectId),
          services.api.gitBranches(activeProjectId),
        ]),
      "Git state refreshed",
      ([nextStatus, nextDiff, nextBranches]) => {
        setStatus(nextStatus)
        setDiff(nextDiff.diff)
        setBranches(nextBranches)
        setBranch(nextStatus.branch)
      },
    )
  }

  const preview = (kind: "pull" | "push"): void => {
    const activeProjectId = dashboard.projectId
    if (activeProjectId === null) return
    action.run(
      () => services.api.previewGit({ projectId: activeProjectId, kind, branch }),
      `${kind} preview created`,
    )
  }

  return (
    <div className="screen-grid screen-grid--two">
      <Panel
        actions={<ActionButton onClick={inspect}>Refresh Git</ActionButton>}
        title="Repository status"
      >
        {status === null ? (
          <EmptyState
            detail="Inspect the selected project repository."
            title="Git state not loaded"
          />
        ) : (
          <>
            <dl className="detail-list">
              <div>
                <dt>Branch</dt>
                <dd>{status.branch}</dd>
              </div>
              <div>
                <dt>Ahead / behind</dt>
                <dd>
                  {status.ahead} / {status.behind}
                </dd>
              </div>
              <div>
                <dt>Staged</dt>
                <dd>{status.staged}</dd>
              </div>
              <div>
                <dt>Modified / untracked</dt>
                <dd>
                  {status.modified} / {status.untracked}
                </dd>
              </div>
            </dl>
            <SelectField
              label="Operation branch"
              onChange={(event) => setBranch(event.target.value)}
              value={branch}
            >
              {branches.map((candidate) => (
                <option key={candidate} value={candidate}>
                  {candidate}
                </option>
              ))}
            </SelectField>
            <div className="form-actions">
              <ActionButton onClick={() => preview("pull")} variant="primary">
                Preview pull
              </ActionButton>
              <ActionButton
                disabled={!snapshot.policy.allow_git_push}
                onClick={() => preview("push")}
              >
                Preview push
              </ActionButton>
            </div>
            <ActionStatus state={action.state} />
          </>
        )}
      </Panel>

      <Panel title="Capped diff" variant="recessed">
        {diff === "" ? (
          <EmptyState detail="Refresh Git to inspect tracked changes." title="No diff loaded" />
        ) : (
          <pre className="log-surface">{diff}</pre>
        )}
      </Panel>

      <Panel title="Git operations">
        {snapshot.gitOperations.length === 0 ? (
          <EmptyState
            detail="Preview pull or push to create an operation."
            title="No Git operations"
          />
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Operation</th>
                  <th>Kind</th>
                  <th>Branch</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {snapshot.gitOperations.map((operation) => (
                  <tr key={operation.id}>
                    <td className="technical">{shortID(operation.id)}</td>
                    <td>{operation.kind}</td>
                    <td>{operation.branch}</td>
                    <td>
                      <StatusBadge status={operation.status} />
                    </td>
                    <td>{formatTimestamp(operation.created_at)}</td>
                    <td>
                      <div className="action-cluster">
                        <ActionButton
                          disabled={operation.status !== "draft"}
                          onClick={() =>
                            action.run(
                              () => services.api.approveGit(operation.id),
                              "Git operation approved",
                            )
                          }
                        >
                          Approve
                        </ActionButton>
                        <ActionButton
                          disabled={operation.status !== "approved"}
                          onClick={() =>
                            action.run(
                              () => services.api.executeGit(operation.id),
                              "Git operation executed",
                            )
                          }
                          variant="primary"
                        >
                          Execute
                        </ActionButton>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>
    </div>
  )
}
