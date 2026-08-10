import { z } from "zod"

import {
  auditSchema,
  pluginSchema,
  policySchema,
  providerSchema,
  usageSchema,
  workerSchema,
} from "./system-schemas"
import type { Requester } from "./transport"

export type CreateProviderInput = {
  readonly name: string
  readonly kind: string
  readonly baseUrl: string
  readonly model: string
  readonly secretName: string
  readonly priority: number
  readonly enabled: boolean
}

export type UpdatePolicyInput = {
  readonly allowGitPush: boolean
  readonly allowSwarmDeploy: boolean
  readonly agentTools: readonly string[]
  readonly maxConcurrentAgents: number
  readonly auditRetentionDays: number
}

export type RegisterWorkerInput = {
  readonly targetId: string
  readonly name: string
  readonly baseUrl: string
  readonly secretName: string
  readonly enabled: boolean
}

export const createSystemApi = (request: Requester) => ({
  createProvider: (input: CreateProviderInput) =>
    request({
      method: "POST",
      path: "/api/v1/ai/providers",
      body: {
        name: input.name,
        kind: input.kind,
        base_url: input.baseUrl,
        model: input.model,
        secret_name: input.secretName,
        priority: input.priority,
        enabled: input.enabled,
      },
      schema: providerSchema,
    }),
  listProviders: () =>
    request({
      method: "GET",
      path: "/api/v1/ai/providers",
      schema: z.array(providerSchema).readonly(),
    }),
  listUsage: () =>
    request({ method: "GET", path: "/api/v1/ai/usage", schema: z.array(usageSchema).readonly() }),
  listModels: (providerId: string) =>
    request({
      method: "GET",
      path: `/api/v1/ai/providers/${encodeURIComponent(providerId)}/models`,
      schema: z.array(z.string()).readonly(),
    }),
  getPolicy: () => request({ method: "GET", path: "/api/v1/policy", schema: policySchema }),
  updatePolicy: (input: UpdatePolicyInput) =>
    request({
      method: "PUT",
      path: "/api/v1/policy",
      body: {
        allow_git_push: input.allowGitPush,
        allow_swarm_deploy: input.allowSwarmDeploy,
        agent_tools: input.agentTools,
        max_concurrent_agents: input.maxConcurrentAgents,
        audit_retention_days: input.auditRetentionDays,
      },
      schema: policySchema,
    }),
  listAudit: () =>
    request({ method: "GET", path: "/api/v1/audit", schema: z.array(auditSchema).readonly() }),
  listWorkers: () =>
    request({ method: "GET", path: "/api/v1/workers", schema: z.array(workerSchema).readonly() }),
  registerWorker: (input: RegisterWorkerInput) =>
    request({
      method: "POST",
      path: "/api/v1/workers",
      body: {
        target_id: input.targetId,
        name: input.name,
        base_url: input.baseUrl,
        secret_name: input.secretName,
        enabled: input.enabled,
      },
      schema: workerSchema,
    }),
  heartbeatWorker: (workerId: string) =>
    request({
      method: "POST",
      path: `/api/v1/workers/${encodeURIComponent(workerId)}/heartbeat`,
      schema: workerSchema,
    }),
  listPlugins: () =>
    request({ method: "GET", path: "/api/v1/plugins", schema: z.array(pluginSchema).readonly() }),
  invokePlugin: (pluginId: string, method: string, input: unknown) =>
    request({
      method: "POST",
      path: `/api/v1/plugins/${encodeURIComponent(pluginId)}/invoke`,
      body: { method, input },
      schema: z.unknown(),
    }),
})
