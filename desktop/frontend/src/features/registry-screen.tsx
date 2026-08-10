import { useState } from "react"

import { PathField, TextField } from "../components/forms"
import { ActionButton, InlineError, Panel } from "../components/primitives"
import { useAction } from "../state/use-action"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

export const RegistryScreen = ({ dashboard, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [projectName, setProjectName] = useState("")
  const [rootPath, setRootPath] = useState("")
  const [targetName, setTargetName] = useState("")
  const [host, setHost] = useState("")
  const [port, setPort] = useState("22")
  const [user, setUser] = useState("")
  const [remotePath, setRemotePath] = useState("")
  const [branch, setBranch] = useState("main")
  const [pathError, setPathError] = useState<string | null>(null)

  const browseProject = (): void => {
    services.desktop
      .choosePath("project_directory")
      .then((path) => {
        setPathError(null)
        if (path !== "") setRootPath(path)
      })
      .catch((error: unknown) => {
        setPathError(error instanceof Error ? error.message : "Unable to open directory picker")
      })
  }

  return (
    <div className="screen-grid screen-grid--two">
      <Panel title="Register project">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            action.run(
              () => services.api.createProject({ name: projectName, rootPath }),
              "Project registered",
              (project) => dashboard.selectProject(project.id),
            )
          }}
        >
          <TextField
            label="Project name"
            onChange={(event) => setProjectName(event.target.value)}
            required
            value={projectName}
          />
          <PathField label="Repository root" onBrowse={browseProject} value={rootPath} />
          {pathError === null ? null : <InlineError message={pathError} />}
          <div className="form-actions">
            <ActionButton busy={action.state.kind === "pending"} type="submit" variant="primary">
              Register
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Register deployment target">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            const activeProjectId = dashboard.projectId
            if (activeProjectId === null) return
            action.run(
              () =>
                services.api.createTarget({
                  projectId: activeProjectId,
                  name: targetName,
                  host,
                  port: Number(port),
                  user,
                  remotePath,
                  branch,
                }),
              "Target registered",
            )
          }}
        >
          <div className="form-grid">
            <TextField
              label="Target name"
              onChange={(event) => setTargetName(event.target.value)}
              required
              value={targetName}
            />
            <TextField
              label="SSH host"
              onChange={(event) => setHost(event.target.value)}
              required
              value={host}
            />
            <TextField
              label="SSH port"
              onChange={(event) => setPort(event.target.value)}
              required
              type="number"
              value={port}
            />
            <TextField
              label="SSH user"
              onChange={(event) => setUser(event.target.value)}
              required
              value={user}
            />
            <TextField
              label="Remote path"
              onChange={(event) => setRemotePath(event.target.value)}
              required
              value={remotePath}
            />
            <TextField
              label="Branch"
              onChange={(event) => setBranch(event.target.value)}
              required
              value={branch}
            />
          </div>
          <div className="form-actions">
            <ActionButton disabled={dashboard.projectId === null} type="submit" variant="primary">
              Register target
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>
    </div>
  )
}
