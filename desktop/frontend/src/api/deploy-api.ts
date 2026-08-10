import { z } from "zod"

import { eventSchema, planSchema, projectSchema, targetSchema } from "./core-schemas"
import { swarmNodeSchema, swarmServiceSchema } from "./system-schemas"
import type { Requester } from "./transport"

type PreviewInput = {
  readonly projectId: string
  readonly targetId: string
  readonly prompt: string
}

export type CreateProjectInput = {
  readonly name: string
  readonly rootPath: string
}

export type CreateTargetInput = {
  readonly projectId: string
  readonly name: string
  readonly host: string
  readonly port: number
  readonly user: string
  readonly remotePath: string
  readonly branch: string
}

const segment = (value: string): string => encodeURIComponent(value)

export const createDeployApi = (request: Requester) => ({
  createProject: (input: CreateProjectInput) =>
    request({
      method: "POST",
      path: "/api/v1/projects",
      body: { name: input.name, root_path: input.rootPath },
      schema: projectSchema,
    }),
  listProjects: () =>
    request({ method: "GET", path: "/api/v1/projects", schema: z.array(projectSchema).readonly() }),
  listTargets: (projectId: string) =>
    request({
      method: "GET",
      path: `/api/v1/projects/${segment(projectId)}/targets`,
      schema: z.array(targetSchema).readonly(),
    }),
  createTarget: (input: CreateTargetInput) =>
    request({
      method: "POST",
      path: "/api/v1/targets",
      body: {
        project_id: input.projectId,
        name: input.name,
        host: input.host,
        port: input.port,
        user: input.user,
        remote_path: input.remotePath,
        branch: input.branch,
      },
      schema: targetSchema,
    }),
  listPlans: () =>
    request({ method: "GET", path: "/api/v1/plans", schema: z.array(planSchema).readonly() }),
  getPlan: (planId: string) =>
    request({ method: "GET", path: `/api/v1/plans/${segment(planId)}`, schema: planSchema }),
  listEvents: (planId: string) =>
    request({
      method: "GET",
      path: `/api/v1/plans/${segment(planId)}/events`,
      schema: z.array(eventSchema).readonly(),
    }),
  previewDeploy: (input: PreviewInput) =>
    request({
      method: "POST",
      path: "/api/v1/plans/preview",
      body: { project_id: input.projectId, target_id: input.targetId, prompt: input.prompt },
      schema: planSchema,
    }),
  approvePlan: (planId: string) =>
    request({
      method: "POST",
      path: `/api/v1/plans/${segment(planId)}/approve`,
      schema: planSchema,
    }),
  executePlan: (planId: string) =>
    request({
      method: "POST",
      path: `/api/v1/plans/${segment(planId)}/execute`,
      schema: planSchema,
    }),
  cancelPlan: (planId: string) =>
    request({
      method: "POST",
      path: `/api/v1/plans/${segment(planId)}/cancel`,
      schema: z.object({ status: z.literal("cancellation_requested") }).readonly(),
    }),
  retryPlan: (planId: string) =>
    request({ method: "POST", path: `/api/v1/plans/${segment(planId)}/retry`, schema: planSchema }),
  previewRollback: (targetId: string) =>
    request({
      method: "POST",
      path: `/api/v1/targets/${segment(targetId)}/rollback-preview`,
      schema: planSchema,
    }),
  listSwarmNodes: (targetId: string) =>
    request({
      method: "GET",
      path: `/api/v1/targets/${segment(targetId)}/swarm/nodes`,
      schema: z.array(swarmNodeSchema).readonly(),
    }),
  listSwarmServices: (targetId: string) =>
    request({
      method: "GET",
      path: `/api/v1/targets/${segment(targetId)}/swarm/services`,
      schema: z.array(swarmServiceSchema).readonly(),
    }),
})
