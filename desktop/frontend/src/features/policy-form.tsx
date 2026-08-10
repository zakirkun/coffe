import { useState } from "react"

import type { Policy } from "../api/system-schemas"
import { CheckboxField, TextField } from "../components/forms"
import { ActionButton, Panel } from "../components/primitives"
import { useAction } from "../state/use-action"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

const agentTools = ["opencode", "claude", "codex", "gemini", "aider"] as const

type PolicyDraft = Readonly<{
  allowGitPush: boolean
  allowSwarmDeploy: boolean
  allowedTools: readonly string[]
  maxAgents: string
  auditDays: string
}>

const fromPolicy = (policy: Policy): PolicyDraft => ({
  allowGitPush: policy.allow_git_push,
  allowSwarmDeploy: policy.allow_swarm_deploy,
  allowedTools: policy.agent_tools,
  maxAgents: String(policy.max_concurrent_agents),
  auditDays: String(policy.audit_retention_days),
})

export const PolicyForm = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [edited, setEdited] = useState<PolicyDraft | null>(null)
  const draft = edited ?? fromPolicy(snapshot.policy)

  const toggleTool = (tool: string, checked: boolean): void => {
    setEdited({
      ...draft,
      allowedTools: checked
        ? [...draft.allowedTools, tool]
        : draft.allowedTools.filter((candidate) => candidate !== tool),
    })
  }

  return (
    <Panel title="Safety policy" variant="approval">
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          action.run(
            () =>
              services.api.updatePolicy({
                allowGitPush: draft.allowGitPush,
                allowSwarmDeploy: draft.allowSwarmDeploy,
                agentTools: draft.allowedTools,
                maxConcurrentAgents: Number(draft.maxAgents),
                auditRetentionDays: Number(draft.auditDays),
              }),
            "Policy updated",
            () => setEdited(null),
          )
        }}
      >
        <div className="form-grid">
          <CheckboxField
            checked={draft.allowGitPush}
            label="Allow approved Git push"
            onChange={(event) => setEdited({ ...draft, allowGitPush: event.target.checked })}
          />
          <CheckboxField
            checked={draft.allowSwarmDeploy}
            label="Allow approved Swarm deploy"
            onChange={(event) => setEdited({ ...draft, allowSwarmDeploy: event.target.checked })}
          />
          <TextField
            label="Maximum concurrent agents"
            onChange={(event) => setEdited({ ...draft, maxAgents: event.target.value })}
            required
            type="number"
            value={draft.maxAgents}
          />
          <TextField
            label="Audit retention days"
            onChange={(event) => setEdited({ ...draft, auditDays: event.target.value })}
            required
            type="number"
            value={draft.auditDays}
          />
        </div>
        <fieldset className="checkbox-group">
          <legend>Allowed agent tools</legend>
          {agentTools.map((tool) => (
            <CheckboxField
              checked={draft.allowedTools.includes(tool)}
              key={tool}
              label={tool}
              onChange={(event) => toggleTool(tool, event.target.checked)}
            />
          ))}
        </fieldset>
        <div className="form-actions">
          <ActionButton type="submit" variant="primary">
            Replace policy
          </ActionButton>
        </div>
        <ActionStatus state={action.state} />
      </form>
    </Panel>
  )
}
