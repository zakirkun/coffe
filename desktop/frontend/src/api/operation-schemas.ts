import { z } from "zod"

const timestampSchema = z.iso.datetime({ offset: true })

export const agentToolSchema = z.enum(["opencode", "claude", "codex", "gemini", "aider"])

export const agentSchema = z
  .object({
    id: z.string().min(1),
    project_id: z.string().min(1),
    tool: agentToolSchema,
    status: z.enum(["pending", "running", "succeeded", "failed", "stopped"]),
    created_at: timestampSchema,
    started_at: timestampSchema.optional(),
    finished_at: timestampSchema.optional(),
  })
  .readonly()

export const agentEventSchema = z
  .object({
    message: z.string().min(1),
    occurred_at: timestampSchema,
  })
  .readonly()

export const databaseSchema = z
  .object({
    id: z.string().min(1),
    name: z.string().min(1),
    driver: z.enum(["sqlite", "mysql"]),
    host: z.string().optional(),
    port: z.number().int().optional(),
    user: z.string().optional(),
    database: z.string().optional(),
    sqlite_path: z.string().optional(),
    tls: z.boolean(),
    has_secret: z.boolean(),
    created_at: timestampSchema,
  })
  .readonly()

export const queryResultSchema = z
  .object({
    columns: z.array(z.string()).readonly(),
    rows: z.array(z.array(z.string()).readonly()).readonly(),
    truncated: z.boolean(),
  })
  .readonly()

export const databaseTableSchema = z
  .object({ name: z.string().min(1), type: z.string().min(1) })
  .readonly()

export const databaseColumnSchema = z
  .object({
    name: z.string().min(1),
    data_type: z.string(),
    nullable: z.boolean(),
  })
  .readonly()

export const queryHistorySchema = z
  .object({
    id: z.string().min(1),
    database_id: z.string().min(1),
    statement: z.string().min(1),
    row_count: z.number().int(),
    duration_millis: z.number().int(),
    truncated: z.boolean(),
    success: z.boolean(),
    created_at: timestampSchema,
  })
  .readonly()

export const gitStatusSchema = z
  .object({
    branch: z.string(),
    ahead: z.number().int(),
    behind: z.number().int(),
    staged: z.number().int(),
    modified: z.number().int(),
    untracked: z.number().int(),
  })
  .readonly()

export const gitOperationSchema = z
  .object({
    id: z.string().min(1),
    project_id: z.string().min(1),
    kind: z.enum(["pull", "push"]),
    remote: z.string().min(1),
    branch: z.string().min(1),
    status: z.enum(["draft", "approved", "running", "succeeded", "failed"]),
    command: z.string().optional(),
    created_at: timestampSchema,
    approved_at: timestampSchema.optional(),
  })
  .readonly()

export type Agent = z.infer<typeof agentSchema>
export type AgentTool = z.infer<typeof agentToolSchema>
export type AgentEvent = z.infer<typeof agentEventSchema>
export type Database = z.infer<typeof databaseSchema>
export type QueryResult = z.infer<typeof queryResultSchema>
export type DatabaseTable = z.infer<typeof databaseTableSchema>
export type DatabaseColumn = z.infer<typeof databaseColumnSchema>
export type QueryHistory = z.infer<typeof queryHistorySchema>
export type GitStatus = z.infer<typeof gitStatusSchema>
export type GitOperation = z.infer<typeof gitOperationSchema>
