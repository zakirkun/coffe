import { z } from "zod"

const timestampSchema = z.iso.datetime({ offset: true })

export const projectSchema = z
  .object({
    id: z.string().min(1),
    name: z.string().min(1),
    root_path: z.string().min(1),
    created_at: timestampSchema,
  })
  .readonly()

export const targetSchema = z
  .object({
    id: z.string().min(1),
    project_id: z.string().min(1),
    name: z.string().min(1),
    host: z.string().min(1),
    port: z.number().int().positive(),
    user: z.string().min(1),
    remote_path: z.string().min(1),
    branch: z.string().min(1),
    created_at: timestampSchema,
  })
  .readonly()

export const planStatusSchema = z.enum([
  "draft",
  "approved",
  "running",
  "succeeded",
  "failed",
  "canceled",
])

export const planSchema = z
  .object({
    id: z.string().min(1),
    project_id: z.string().min(1),
    target_id: z.string().min(1),
    source_plan_id: z.string().optional(),
    kind: z.enum(["compose_deploy", "compose_rollback", "swarm_deploy"]),
    status: planStatusSchema,
    revision: z.string().optional(),
    steps: z.array(z.string().min(1)).readonly(),
    commands: z.array(z.string()).readonly().optional(),
    created_at: timestampSchema,
    approved_at: timestampSchema.optional(),
  })
  .readonly()

export const eventSchema = z
  .object({
    id: z.string().min(1),
    plan_id: z.string().min(1),
    kind: z.string().min(1),
    message: z.string().min(1),
    occurred_at: timestampSchema,
  })
  .readonly()

export type Project = z.infer<typeof projectSchema>
export type Target = z.infer<typeof targetSchema>
export type Plan = z.infer<typeof planSchema>
export type Event = z.infer<typeof eventSchema>
