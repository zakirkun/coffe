import { useState } from "react"

import { CheckboxField, FieldFrame, PathField, SelectField, TextField } from "../components/forms"
import { ActionButton, EmptyState, InlineError, Panel } from "../components/primitives"
import { reconcileSelection, selectionChannel, useAction } from "../state/use-action"
import { IntegrationWorkers } from "./integration-workers"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

type PluginInvocation = Readonly<{
  method: string
  result: string
}>

const emptyPluginInvocation: PluginInvocation = { method: "", result: "" }

export const IntegrationsScreen = ({ dashboard, snapshot, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [workerName, setWorkerName] = useState("")
  const [workerURL, setWorkerURL] = useState("")
  const [workerSecret, setWorkerSecret] = useState("")
  const [workerEnabled, setWorkerEnabled] = useState(true)
  const [pluginName, setPluginName] = useState("")
  const [pluginPath, setPluginPath] = useState("")
  const [pluginMethods, setPluginMethods] = useState("")
  const [pluginEnabled, setPluginEnabled] = useState(true)
  const [requestedPluginId, setPluginId] = useState(snapshot.plugins[0]?.id ?? "")
  const pluginId = reconcileSelection(requestedPluginId, snapshot.plugins)
  const [pluginStates, setPluginStates] = useState<Readonly<Record<string, PluginInvocation>>>({})
  const invocation = pluginStates[pluginId] ?? emptyPluginInvocation
  const [invokeInput, setInvokeInput] = useState("{}")
  const [localError, setLocalError] = useState<string | null>(null)

  const browsePlugin = (): void => {
    services.desktop
      .choosePath("plugin_executable")
      .then((path) => {
        setLocalError(null)
        if (path !== "") setPluginPath(path)
      })
      .catch((error: unknown) =>
        setLocalError(error instanceof Error ? error.message : "Unable to choose plugin"),
      )
  }

  const invoke = (): void => {
    let input: unknown
    try {
      input = JSON.parse(invokeInput)
    } catch (error) {
      if (error instanceof SyntaxError) {
        setLocalError("Plugin input must be valid JSON")
        return
      }
      throw error
    }
    action.run(
      () => services.api.invokePlugin(pluginId, invocation.method, input),
      "Plugin invoked",
      (result) =>
        setPluginStates((current) => ({
          ...current,
          [pluginId]: {
            ...(current[pluginId] ?? emptyPluginInvocation),
            result: JSON.stringify(result, null, 2),
          },
        })),
      selectionChannel("plugin-invoke", pluginId),
    )
  }

  const heartbeatWorker = (workerId: string): void => {
    action.run(
      () => services.api.heartbeatWorker(workerId),
      "Worker heartbeat succeeded",
      undefined,
      `worker-heartbeat:${workerId}`,
    )
  }

  return (
    <div className="screen-grid screen-grid--two">
      <Panel title="Register remote worker">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            const activeTargetId = dashboard.targetId
            if (activeTargetId === null) return
            action.run(
              () =>
                services.api.registerWorker({
                  targetId: activeTargetId,
                  name: workerName,
                  baseUrl: workerURL,
                  secretName: workerSecret,
                  enabled: workerEnabled,
                }),
              "Worker registered",
              undefined,
              "worker-register",
            )
          }}
        >
          <div className="form-grid">
            <TextField
              label="Worker name"
              onChange={(event) => setWorkerName(event.target.value)}
              required
              value={workerName}
            />
            <TextField
              label="HTTPS base URL"
              onChange={(event) => setWorkerURL(event.target.value)}
              required
              type="url"
              value={workerURL}
            />
            <TextField
              helper="Bearer token remains in the OS keyring."
              label="Token secret name"
              onChange={(event) => setWorkerSecret(event.target.value)}
              required
              value={workerSecret}
            />
          </div>
          <CheckboxField
            checked={workerEnabled}
            label="Route selected target through worker"
            onChange={(event) => setWorkerEnabled(event.target.checked)}
          />
          <div className="form-actions">
            <ActionButton disabled={dashboard.targetId === null} type="submit" variant="primary">
              Register worker
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Register plugin">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            action.run(
              () =>
                services.desktop.registerPlugin({
                  name: pluginName,
                  executable_path: pluginPath,
                  methods: pluginMethods
                    .split(",")
                    .map((method) => method.trim())
                    .filter((method) => method !== ""),
                  enabled: pluginEnabled,
                }),
              "Plugin hashed and registered",
              undefined,
              "plugin-register",
            )
          }}
        >
          <TextField
            label="Plugin name"
            onChange={(event) => setPluginName(event.target.value)}
            required
            value={pluginName}
          />
          <PathField label="Plugin executable" onBrowse={browsePlugin} value={pluginPath} />
          <TextField
            helper="Comma-separated JSON-stdio methods."
            label="Declared methods"
            onChange={(event) => setPluginMethods(event.target.value)}
            required
            value={pluginMethods}
          />
          <CheckboxField
            checked={pluginEnabled}
            label="Enable plugin invocation"
            onChange={(event) => setPluginEnabled(event.target.checked)}
          />
          <div className="form-actions">
            <ActionButton type="submit" variant="primary">
              Hash and register
            </ActionButton>
          </div>
          {localError === null ? null : <InlineError message={localError} />}
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <IntegrationWorkers onHeartbeat={heartbeatWorker} workers={snapshot.workers} />

      <Panel title="Invoke registered plugin" variant="approval">
        {snapshot.plugins.length === 0 ? (
          <EmptyState detail="Register a hash-pinned plugin first." title="No plugins" />
        ) : (
          <div className="form-stack">
            <SelectField
              label="Plugin"
              onChange={(event) => setPluginId(event.target.value)}
              value={pluginId}
            >
              {snapshot.plugins.map((plugin) => (
                <option key={plugin.id} value={plugin.id}>
                  {plugin.name}
                </option>
              ))}
            </SelectField>
            <TextField
              label="Declared method"
              onChange={(event) =>
                setPluginStates((current) => ({
                  ...current,
                  [pluginId]: { ...invocation, method: event.target.value },
                }))
              }
              required
              value={invocation.method}
            />
            <FieldFrame id="plugin-input" label="JSON input">
              <textarea
                className="technical"
                id="plugin-input"
                onChange={(event) => setInvokeInput(event.target.value)}
                value={invokeInput}
              />
            </FieldFrame>
            <div className="form-actions">
              <ActionButton onClick={invoke} variant="primary">
                Invoke
              </ActionButton>
            </div>
            {invocation.result === "" ? null : (
              <pre className="log-surface">{invocation.result}</pre>
            )}
          </div>
        )}
      </Panel>
    </div>
  )
}
