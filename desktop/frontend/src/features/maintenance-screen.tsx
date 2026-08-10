import { useEffect, useState } from "react"

import type { DesktopApi, PathKind } from "../api/desktop-client"
import { CheckboxField, PathField, TextField } from "../components/forms"
import { ActionButton, InlineError, Panel, StatusBadge } from "../components/primitives"
import { useAction } from "../state/use-action"
import { ActionStatus } from "./shared"
import type { ScreenProps } from "./types"

type Release = Awaited<ReturnType<DesktopApi["checkUpdate"]>>
type Version = Awaited<ReturnType<DesktopApi["versionInfo"]>>

export const MaintenanceScreen = ({ dashboard, services }: ScreenProps) => {
  const action = useAction(dashboard.refresh)
  const [secretName, setSecretName] = useState("")
  const [secretValue, setSecretValue] = useState("")
  const [backupOutput, setBackupOutput] = useState("")
  const [backupPassphrase, setBackupPassphrase] = useState("")
  const [backupRepeat, setBackupRepeat] = useState("")
  const [backupInput, setBackupInput] = useState("")
  const [importPassphrase, setImportPassphrase] = useState("")
  const [importConfirmed, setImportConfirmed] = useState(false)
  const [manifestURL, setManifestURL] = useState("")
  const [currentVersion, setCurrentVersion] = useState("")
  const [release, setRelease] = useState<Release | null>(null)
  const [version, setVersion] = useState<Version | null>(null)
  const [localError, setLocalError] = useState<string | null>(null)

  useEffect(() => {
    services.desktop
      .versionInfo()
      .then((info) => {
        setVersion(info)
        if (info.version !== "dev" && info.version !== "") {
          setCurrentVersion(info.version.startsWith("v") ? info.version : `v${info.version}`)
        }
      })
      .catch((error: unknown) => {
        setLocalError(error instanceof Error ? error.message : "Unable to read version information")
      })
  }, [services.desktop])

  const choosePath = (kind: PathKind, setPath: (path: string) => void): void => {
    services.desktop
      .choosePath(kind)
      .then((path) => {
        setLocalError(null)
        if (path !== "") setPath(path)
      })
      .catch((error: unknown) => {
        setLocalError(error instanceof Error ? error.message : "Unable to open native path dialog")
      })
  }

  const downloadUpdate = (): void => {
    choosePath("update_output", (path) => {
      action.run(
        () => services.desktop.downloadUpdate(manifestURL, currentVersion, path),
        "Verified update staged",
      )
    })
  }

  return (
    <div className="screen-grid screen-grid--two">
      <Panel title="OS keyring secret">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            action.run(
              () => services.desktop.storeSecret(secretName, secretValue),
              "Secret stored in OS keyring",
              () => setSecretValue(""),
            )
          }}
        >
          <TextField
            label="Secret name"
            onChange={(event) => setSecretName(event.target.value)}
            required
            value={secretName}
          />
          <TextField
            helper="The value never enters SQLite or logs."
            label="Secret value"
            onChange={(event) => setSecretValue(event.target.value)}
            required
            type="password"
            value={secretValue}
          />
          <div className="form-actions">
            <ActionButton type="submit" variant="primary">
              Store secret
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Export encrypted backup">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            action.run(
              () => services.desktop.exportBackup(backupOutput, backupPassphrase, backupRepeat),
              "Encrypted backup exported",
              () => {
                setBackupPassphrase("")
                setBackupRepeat("")
              },
            )
          }}
        >
          <PathField
            label="Exclusive output file"
            onBrowse={() => choosePath("backup_export", setBackupOutput)}
            value={backupOutput}
          />
          <div className="form-grid">
            <TextField
              helper="Minimum 12 characters."
              label="Passphrase"
              onChange={(event) => setBackupPassphrase(event.target.value)}
              required
              type="password"
              value={backupPassphrase}
            />
            <TextField
              label="Repeat passphrase"
              onChange={(event) => setBackupRepeat(event.target.value)}
              required
              type="password"
              value={backupRepeat}
            />
          </div>
          <div className="form-actions">
            <ActionButton type="submit" variant="primary">
              Export backup
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Import encrypted backup" variant="approval">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            action.run(
              () => services.desktop.importBackup(backupInput, importPassphrase, importConfirmed),
              "Backup imported and runtime reopened",
              () => setImportPassphrase(""),
            )
          }}
        >
          <PathField
            label="Encrypted backup file"
            onBrowse={() => choosePath("backup_import", setBackupInput)}
            value={backupInput}
          />
          <TextField
            label="Backup passphrase"
            onChange={(event) => setImportPassphrase(event.target.value)}
            required
            type="password"
            value={importPassphrase}
          />
          <div className="confirmation">
            <strong>Replace local workspace state</strong>
            <p>
              The current database is preserved beside the imported copy. Keyring values are not
              restored.
            </p>
            <CheckboxField
              checked={importConfirmed}
              label="I understand this replaces the active SQLite state"
              onChange={(event) => setImportConfirmed(event.target.checked)}
            />
          </div>
          <div className="form-actions">
            <ActionButton disabled={!importConfirmed} type="submit" variant="destructive">
              Import backup
            </ActionButton>
          </div>
          <ActionStatus state={action.state} />
        </form>
      </Panel>

      <Panel title="Verified update staging">
        <div className="form-stack">
          <dl className="detail-list">
            <div>
              <dt>Version</dt>
              <dd>{version?.version ?? "unknown"}</dd>
            </div>
            <div>
              <dt>Commit</dt>
              <dd className="technical">{version?.commit ?? "unknown"}</dd>
            </div>
            <div>
              <dt>Built</dt>
              <dd>{version?.date ?? "unknown"}</dd>
            </div>
          </dl>
          <TextField
            label="HTTPS manifest URL"
            onChange={(event) => setManifestURL(event.target.value)}
            required
            type="url"
            value={manifestURL}
          />
          <TextField
            helper="Required for development builds."
            label="Current semantic version"
            onChange={(event) => setCurrentVersion(event.target.value)}
            placeholder="v1.0.0"
            required
            value={currentVersion}
          />
          <div className="form-actions">
            <ActionButton
              onClick={() =>
                action.run(
                  () => services.desktop.checkUpdate(manifestURL, currentVersion),
                  "Update manifest checked",
                  setRelease,
                )
              }
              variant="primary"
            >
              Check update
            </ActionButton>
            <ActionButton disabled={release?.available !== true} onClick={downloadUpdate}>
              Download verified asset
            </ActionButton>
          </div>
          {release === null ? null : (
            <p className="context-line">
              <span>{release.version}</span>
              <StatusBadge status={release.available ? "ready" : "stopped"} />
            </p>
          )}
          {localError === null ? null : <InlineError message={localError} />}
          <ActionStatus state={action.state} />
        </div>
      </Panel>
    </div>
  )
}
