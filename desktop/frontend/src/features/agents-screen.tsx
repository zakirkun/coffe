import { useState } from "react"

import type { AgentEvent, AgentTool } from "../api/operation-schemas"
import { FieldFrame, SelectField } from "../components/forms"
import { ActionButton, EmptyState, Panel, StatusBadge } from "../components/primitives"
import { useAction } from "../state/use-action"
import { formatTimestamp, shortID } from "./format"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

const tools: readonly AgentTool[] = ["opencode", "claude", "codex", "gemini", "aider"]

export const AgentsScreen = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [tool, setTool] = useState<AgentTool>("opencode")
  const [prompt, setPrompt] = useState("")
  const [events, setEvents] = useState<readonly AgentEvent[]>([])

  return (
    <div className="screen-grid screen-grid--two">
      <Panel title="Start coding agent">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            const activeProjectId = dashboard.projectId
            if (activeProjectId === null) return
            action.run(
              () => services.api.startAgent({ projectId: activeProjectId, tool, prompt }),
              "Agent started",
            )
          }}
        >
          <SelectField
            label="Agent tool"
            onChange={(event) => {
              const selected = tools.find((candidate) => candidate === event.target.value)
              if (selected !== undefined) setTool(selected)
            }}
            value={tool}
          >
            {tools.map((candidate) => (
              <option key={candidate} value={candidate}>
                {candidate}
              </option>
            ))}
          </SelectField>
          <FieldFrame id="agent-prompt" label="Agent prompt">
            <textarea
              id="agent-prompt"
              onChange={(event) => setPrompt(event.target.value)}
              required
              value={prompt}
            />
          </FieldFrame>
          <div className="form-actions">
            <ActionButton disabled={dashboard.projectId === null} type="submit" variant="primary">
              Start agent
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Agent events" variant="recessed">
        {events.length === 0 ? (
          <EmptyState
            detail="Inspect a session to read its capped event stream."
            title="No session selected"
          />
        ) : (
          <ol className="activity-feed">
            {events.map((event) => (
              <li key={`${event.occurred_at}-${event.message}`}>
                <time dateTime={event.occurred_at}>{formatTimestamp(event.occurred_at)}</time>
                <div>
                  <span>{event.message}</span>
                </div>
              </li>
            ))}
          </ol>
        )}
      </Panel>

      <Panel title="Agent sessions">
        {snapshot.agents.length === 0 ? (
          <EmptyState
            detail="Start a coding agent against the selected project."
            title="No agent sessions"
          />
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Session</th>
                  <th>Tool</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {snapshot.agents.map((agent) => (
                  <tr key={agent.id}>
                    <td className="technical">{shortID(agent.id)}</td>
                    <td>{agent.tool}</td>
                    <td>
                      <StatusBadge status={agent.status} />
                    </td>
                    <td>{formatTimestamp(agent.created_at)}</td>
                    <td>
                      <div className="action-cluster">
                        <ActionButton
                          onClick={() =>
                            action.run(
                              () => services.api.listAgentEvents(agent.id),
                              "Agent events loaded",
                              setEvents,
                            )
                          }
                        >
                          Inspect
                        </ActionButton>
                        <ActionButton
                          disabled={agent.status !== "running"}
                          onClick={() =>
                            action.run(() => services.api.stopAgent(agent.id), "Stop requested")
                          }
                          variant="destructive"
                        >
                          Stop
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
