import { useState } from "react"

import { CheckboxField, TextField } from "../components/forms"
import { ActionButton, EmptyState, Panel, StatusBadge } from "../components/primitives"
import { useAction } from "../state/use-action"
import { PolicyForm } from "./policy-form"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

export const AIPolicyScreen = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [name, setName] = useState("")
  const [kind, setKind] = useState("openai")
  const [baseUrl, setBaseURL] = useState("")
  const [model, setModel] = useState("")
  const [secretName, setSecretName] = useState("")
  const [priority, setPriority] = useState("100")
  const [enabled, setEnabled] = useState(true)
  const [models, setModels] = useState<readonly string[]>([])
  return (
    <div className="screen-grid screen-grid--two">
      <Panel title="Register AI provider">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            action.run(
              () =>
                services.api.createProvider({
                  name,
                  kind,
                  baseUrl,
                  model,
                  secretName,
                  priority: Number(priority),
                  enabled,
                }),
              "AI provider registered",
            )
          }}
        >
          <div className="form-grid">
            <TextField
              label="Provider name"
              onChange={(event) => setName(event.target.value)}
              required
              value={name}
            />
            <TextField
              label="Kind"
              onChange={(event) => setKind(event.target.value)}
              required
              value={kind}
            />
            <TextField
              label="Base URL"
              onChange={(event) => setBaseURL(event.target.value)}
              required
              type="url"
              value={baseUrl}
            />
            <TextField
              label="Default model"
              onChange={(event) => setModel(event.target.value)}
              value={model}
            />
            <TextField
              helper="Optional for local providers."
              label="API key secret name"
              onChange={(event) => setSecretName(event.target.value)}
              value={secretName}
            />
            <TextField
              label="Priority"
              onChange={(event) => setPriority(event.target.value)}
              type="number"
              value={priority}
            />
          </div>
          <CheckboxField
            checked={enabled}
            label="Enable provider routing"
            onChange={(event) => setEnabled(event.target.checked)}
          />
          <div className="form-actions">
            <ActionButton type="submit" variant="primary">
              Register provider
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <PolicyForm dashboard={dashboard} services={services} snapshot={snapshot} />

      <Panel title="Provider routing order">
        {snapshot.providers.length === 0 ? (
          <EmptyState
            detail="Register a local or hosted OpenAI-compatible provider."
            title="No providers"
          />
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Priority</th>
                  <th>Name</th>
                  <th>Kind</th>
                  <th>Model</th>
                  <th>Status</th>
                  <th>Models</th>
                </tr>
              </thead>
              <tbody>
                {snapshot.providers.map((provider) => (
                  <tr key={provider.id}>
                    <td>{provider.priority}</td>
                    <td>{provider.name}</td>
                    <td>{provider.kind}</td>
                    <td>{provider.model}</td>
                    <td>
                      <StatusBadge status={provider.enabled ? "ready" : "stopped"} />
                    </td>
                    <td>
                      <ActionButton
                        onClick={() =>
                          action.run(
                            () => services.api.listModels(provider.id),
                            "Models discovered",
                            setModels,
                          )
                        }
                      >
                        Discover
                      </ActionButton>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {models.length === 0 ? null : (
          <div className="tag-list">
            {models.map((availableModel) => (
              <span className="technical" key={availableModel}>
                {availableModel}
              </span>
            ))}
          </div>
        )}
      </Panel>

      <Panel title="Recent AI usage" variant="recessed">
        {snapshot.usage.length === 0 ? (
          <EmptyState
            detail="Token and latency metrics appear after classification calls."
            title="No usage records"
          />
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Model</th>
                  <th>Tokens</th>
                  <th>Latency</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {snapshot.usage.slice(0, 20).map((usage) => (
                  <tr key={usage.id}>
                    <td>{usage.model}</td>
                    <td>{usage.total_tokens}</td>
                    <td>{usage.latency_millis} ms</td>
                    <td>
                      <StatusBadge status={usage.success ? "success" : "failed"} />
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
