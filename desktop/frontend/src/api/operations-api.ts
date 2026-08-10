import { z } from "zod"

import {
  type AgentTool,
  agentEventSchema,
  agentSchema,
  databaseColumnSchema,
  databaseSchema,
  databaseTableSchema,
  gitOperationSchema,
  gitStatusSchema,
  queryHistorySchema,
  queryResultSchema,
} from "./operation-schemas"
import type { Requester } from "./transport"

type StartAgentInput = {
  readonly projectId: string
  readonly tool: AgentTool
  readonly prompt: string
}

type QueryInput = {
  readonly databaseId: string
  readonly sql: string
  readonly offset: number
  readonly limit: number
}

export type CreateDatabaseInput = {
  readonly name: string
  readonly driver: "sqlite" | "mysql"
  readonly host: string
  readonly port: number
  readonly user: string
  readonly database: string
  readonly secretName: string
  readonly sqlitePath: string
  readonly tls: boolean
}

type PreviewGitInput = {
  readonly projectId: string
  readonly kind: "pull" | "push"
  readonly branch: string
}

const segment = (value: string): string => encodeURIComponent(value)

export const createOperationsApi = (request: Requester) => ({
  listAgents: () =>
    request({ method: "GET", path: "/api/v1/agents", schema: z.array(agentSchema).readonly() }),
  startAgent: (input: StartAgentInput) =>
    request({
      method: "POST",
      path: "/api/v1/agents",
      body: { project_id: input.projectId, tool: input.tool, prompt: input.prompt },
      schema: agentSchema,
    }),
  stopAgent: (sessionId: string) =>
    request({
      method: "POST",
      path: `/api/v1/agents/${segment(sessionId)}/stop`,
      schema: z.object({ status: z.literal("stop_requested") }).readonly(),
    }),
  listAgentEvents: (sessionId: string) =>
    request({
      method: "GET",
      path: `/api/v1/agents/${segment(sessionId)}/events`,
      schema: z.array(agentEventSchema).readonly(),
    }),
  listDatabases: () =>
    request({
      method: "GET",
      path: "/api/v1/databases",
      schema: z.array(databaseSchema).readonly(),
    }),
  createDatabase: (input: CreateDatabaseInput) =>
    request({
      method: "POST",
      path: "/api/v1/databases",
      body: {
        name: input.name,
        driver: input.driver,
        host: input.host,
        port: input.port,
        user: input.user,
        database: input.database,
        secret_name: input.secretName,
        sqlite_path: input.sqlitePath,
        tls: input.tls,
      },
      schema: databaseSchema,
    }),
  queryDatabase: (input: QueryInput) =>
    request({
      method: "POST",
      path: `/api/v1/databases/${segment(input.databaseId)}/query`,
      body: { sql: input.sql, offset: input.offset, limit: input.limit },
      schema: queryResultSchema,
    }),
  explainDatabase: (databaseId: string, sql: string) =>
    request({
      method: "POST",
      path: `/api/v1/databases/${segment(databaseId)}/explain`,
      body: { sql, offset: 0, limit: 200 },
      schema: queryResultSchema,
    }),
  listDatabaseTables: (databaseId: string) =>
    request({
      method: "GET",
      path: `/api/v1/databases/${segment(databaseId)}/tables`,
      schema: z.array(databaseTableSchema).readonly(),
    }),
  listDatabaseColumns: (databaseId: string, tableName: string) =>
    request({
      method: "GET",
      path: `/api/v1/databases/${segment(databaseId)}/tables/${segment(tableName)}/columns`,
      schema: z.array(databaseColumnSchema).readonly(),
    }),
  listQueryHistory: (databaseId: string) =>
    request({
      method: "GET",
      path: `/api/v1/databases/${segment(databaseId)}/history`,
      schema: z.array(queryHistorySchema).readonly(),
    }),
  gitStatus: (projectId: string) =>
    request({
      method: "GET",
      path: `/api/v1/projects/${segment(projectId)}/git/status`,
      schema: gitStatusSchema,
    }),
  gitDiff: (projectId: string) =>
    request({
      method: "GET",
      path: `/api/v1/projects/${segment(projectId)}/git/diff`,
      schema: z.object({ diff: z.string() }).readonly(),
    }),
  gitBranches: (projectId: string) =>
    request({
      method: "GET",
      path: `/api/v1/projects/${segment(projectId)}/git/branches`,
      schema: z.array(z.string()).readonly(),
    }),
  listGitOperations: () =>
    request({
      method: "GET",
      path: "/api/v1/git/operations",
      schema: z.array(gitOperationSchema).readonly(),
    }),
  previewGit: (input: PreviewGitInput) =>
    request({
      method: "POST",
      path: "/api/v1/git/operations/preview",
      body: {
        project_id: input.projectId,
        kind: input.kind,
        remote: "origin",
        branch: input.branch,
      },
      schema: gitOperationSchema,
    }),
  approveGit: (operationId: string) =>
    request({
      method: "POST",
      path: `/api/v1/git/operations/${segment(operationId)}/approve`,
      schema: gitOperationSchema,
    }),
  executeGit: (operationId: string) =>
    request({
      method: "POST",
      path: `/api/v1/git/operations/${segment(operationId)}/execute`,
      schema: gitOperationSchema,
    }),
})
