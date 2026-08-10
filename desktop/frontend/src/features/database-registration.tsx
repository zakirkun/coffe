import { useState } from "react"

import { CheckboxField, PathField, SelectField, TextField } from "../components/forms"
import { ActionButton, InlineError, Panel } from "../components/primitives"
import { useAction } from "../state/use-action"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

export const DatabaseRegistration = ({ dashboard, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [driver, setDriver] = useState<"sqlite" | "mysql">("sqlite")
  const [name, setName] = useState("")
  const [sqlitePath, setSQLitePath] = useState("")
  const [host, setHost] = useState("")
  const [port, setPort] = useState("3306")
  const [user, setUser] = useState("")
  const [database, setDatabase] = useState("")
  const [secretName, setSecretName] = useState("")
  const [tls, setTLS] = useState(true)
  const [pathError, setPathError] = useState<string | null>(null)

  const browse = (): void => {
    services.desktop
      .choosePath("sqlite_database")
      .then((path) => {
        setPathError(null)
        if (path !== "") setSQLitePath(path)
      })
      .catch((error: unknown) => {
        setPathError(error instanceof Error ? error.message : "Unable to open file picker")
      })
  }

  return (
    <Panel title="Register read-only database">
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          action.run(
            () =>
              services.api.createDatabase({
                name,
                driver,
                host,
                port: Number(port),
                user,
                database,
                secretName,
                sqlitePath,
                tls,
              }),
            "Database registered",
          )
        }}
      >
        <div className="form-grid">
          <TextField
            label="Connection name"
            onChange={(event) => setName(event.target.value)}
            required
            value={name}
          />
          <SelectField
            label="Driver"
            onChange={(event) => setDriver(event.target.value === "mysql" ? "mysql" : "sqlite")}
            value={driver}
          >
            <option value="sqlite">SQLite</option>
            <option value="mysql">MySQL</option>
          </SelectField>
        </div>
        {driver === "sqlite" ? (
          <>
            <PathField label="SQLite database" onBrowse={browse} value={sqlitePath} />
            {pathError === null ? null : <InlineError message={pathError} />}
          </>
        ) : (
          <>
            <div className="form-grid">
              <TextField
                label="Host"
                onChange={(event) => setHost(event.target.value)}
                required
                value={host}
              />
              <TextField
                label="Port"
                onChange={(event) => setPort(event.target.value)}
                required
                type="number"
                value={port}
              />
              <TextField
                label="User"
                onChange={(event) => setUser(event.target.value)}
                required
                value={user}
              />
              <TextField
                label="Database"
                onChange={(event) => setDatabase(event.target.value)}
                required
                value={database}
              />
              <TextField
                helper="Password remains in the OS keyring."
                label="Password secret name"
                onChange={(event) => setSecretName(event.target.value)}
                required
                value={secretName}
              />
            </div>
            <CheckboxField
              checked={tls}
              label="Require TLS"
              onChange={(event) => setTLS(event.target.checked)}
            />
          </>
        )}
        <div className="form-actions">
          <ActionButton type="submit" variant="primary">
            Register database
          </ActionButton>
        </div>
        <ActionStatus state={action.state} />
      </form>
    </Panel>
  )
}
