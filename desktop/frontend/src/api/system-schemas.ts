import { z } from "zod"

const timestampSchema = z.iso.datetime({ offset: true })

export const providerSchema = z
  .object({
    id: z.string().min(1),
    name: z.string().min(1),
    kind: z.string().min(1),
    base_url: z.string().min(1),
    model: z.string(),
    priority: z.number().int(),
    enabled: z.boolean(),
    has_secret: z.boolean(),
    created_at: timestampSchema,
  })
  .readonly()

export const usageSchema = z
  .object({
    id: z.string().min(1),
    provider_id: z.string().min(1),
    model: z.string(),
    prompt_tokens: z.number().int(),
    completion_tokens: z.number().int(),
    total_tokens: z.number().int(),
    latency_millis: z.number().int(),
    success: z.boolean(),
    created_at: timestampSchema,
  })
  .readonly()

export const policySchema = z
  .object({
    allow_git_push: z.boolean(),
    allow_swarm_deploy: z.boolean(),
    agent_tools: z.array(z.string()).readonly(),
    max_concurrent_agents: z.number().int(),
    audit_retention_days: z.number().int(),
    updated_at: timestampSchema,
  })
  .readonly()

export const auditSchema = z
  .object({
    id: z.string().min(1),
    category: z.string().min(1),
    action: z.string().min(1),
    resource_id: z.string().min(1),
    outcome: z.enum(["success", "failure", "denied"]),
    message: z.string().min(1),
    occurred_at: timestampSchema,
  })
  .readonly()

export const workerSchema = z
  .object({
    id: z.string().min(1),
    target_id: z.string().min(1),
    name: z.string().min(1),
    base_url: z.string().min(1),
    version: z.string().optional(),
    enabled: z.boolean(),
    last_seen: timestampSchema.optional(),
    created_at: timestampSchema,
  })
  .readonly()

export const pluginSchema = z
  .object({
    id: z.string().min(1),
    name: z.string().min(1),
    executable_path: z.string().min(1),
    sha256: z.string().min(1),
    methods: z.array(z.string()).readonly(),
    enabled: z.boolean(),
    created_at: timestampSchema,
  })
  .readonly()

export const swarmNodeSchema = z
  .object({
    id: z.string().min(1),
    hostname: z.string().min(1),
    status: z.string().min(1),
    availability: z.string().min(1),
    manager: z.boolean(),
  })
  .readonly()

export const swarmServiceSchema = z
  .object({
    id: z.string().min(1),
    name: z.string().min(1),
    mode: z.string().min(1),
    replicas: z.string().min(1),
    image: z.string().min(1),
  })
  .readonly()

export type Provider = z.infer<typeof providerSchema>
export type Usage = z.infer<typeof usageSchema>
export type Policy = z.infer<typeof policySchema>
export type Audit = z.infer<typeof auditSchema>
export type Worker = z.infer<typeof workerSchema>
export type Plugin = z.infer<typeof pluginSchema>
export type SwarmNode = z.infer<typeof swarmNodeSchema>
export type SwarmService = z.infer<typeof swarmServiceSchema>
