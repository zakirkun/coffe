import { z } from "zod"

const updateReleaseSchema = z
  .object({
    version: z.string().min(1),
    available: z.boolean(),
    asset_url: z.url(),
  })
  .readonly()

const versionInfoSchema = z
  .object({
    version: z.string(),
    commit: z.string(),
    date: z.string(),
  })
  .readonly()

export type PathKind =
  | "project_directory"
  | "sqlite_database"
  | "plugin_executable"
  | "backup_import"
  | "backup_export"
  | "audit_export"
  | "query_export"
  | "update_output"

export type PluginRegistration = Readonly<{
  name: string
  executable_path: string
  methods: readonly string[]
  enabled: boolean
}>

declare global {
  interface CoffeDesktopApp {
    readonly ChoosePath: (kind: PathKind) => Promise<string>
    readonly StoreSecret: (input: Readonly<{ name: string; value: string }>) => Promise<void>
    readonly ExportAudit: (outputPath: string) => Promise<void>
    readonly ExportQuery: (
      input: Readonly<{ database_id: string; sql: string; output_path: string }>,
    ) => Promise<void>
    readonly ExportBackup: (
      input: Readonly<{
        output_path: string
        passphrase: string
        repeated_passphrase: string
      }>,
    ) => Promise<void>
    readonly ImportBackup: (
      input: Readonly<{
        input_path: string
        passphrase: string
        confirmed: boolean
      }>,
    ) => Promise<void>
    readonly CheckUpdate: (
      input: Readonly<{
        manifest_url: string
        current_version: string
      }>,
    ) => Promise<unknown>
    readonly DownloadUpdate: (
      input: Readonly<{
        manifest_url: string
        current_version: string
        output_path: string
      }>,
    ) => Promise<void>
    readonly VersionInfo: () => Promise<unknown>
    readonly RegisterPlugin: (input: PluginRegistration) => Promise<string>
  }
}

export const desktopApi = {
  choosePath: (kind: PathKind) => window.go.main.App.ChoosePath(kind),
  storeSecret: (name: string, value: string) => window.go.main.App.StoreSecret({ name, value }),
  exportAudit: (outputPath: string) => window.go.main.App.ExportAudit(outputPath),
  exportQuery: (databaseId: string, sql: string, outputPath: string) =>
    window.go.main.App.ExportQuery({ database_id: databaseId, sql, output_path: outputPath }),
  exportBackup: (outputPath: string, passphrase: string, repeatedPassphrase: string) =>
    window.go.main.App.ExportBackup({
      output_path: outputPath,
      passphrase,
      repeated_passphrase: repeatedPassphrase,
    }),
  importBackup: (inputPath: string, passphrase: string, confirmed: boolean) =>
    window.go.main.App.ImportBackup({ input_path: inputPath, passphrase, confirmed }),
  checkUpdate: async (manifestUrl: string, currentVersion: string) =>
    updateReleaseSchema.parse(
      await window.go.main.App.CheckUpdate({
        manifest_url: manifestUrl,
        current_version: currentVersion,
      }),
    ),
  downloadUpdate: (manifestUrl: string, currentVersion: string, outputPath: string) =>
    window.go.main.App.DownloadUpdate({
      manifest_url: manifestUrl,
      current_version: currentVersion,
      output_path: outputPath,
    }),
  versionInfo: async () => versionInfoSchema.parse(await window.go.main.App.VersionInfo()),
  registerPlugin: (input: PluginRegistration) => window.go.main.App.RegisterPlugin(input),
}

export type DesktopApi = typeof desktopApi
