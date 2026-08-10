import { useCallback, useEffect, useMemo, useRef, useState } from "react"

import type { Api } from "../api"
import type { Event, Plan, Project, Target } from "../api/core-schemas"
import type { Agent, Database, GitOperation } from "../api/operation-schemas"
import type { Audit, Plugin, Policy, Provider, Usage, Worker } from "../api/system-schemas"

export type DashboardSnapshot = Readonly<{
  projects: readonly Project[]
  targets: readonly Target[]
  plans: readonly Plan[]
  events: readonly Event[]
  agents: readonly Agent[]
  databases: readonly Database[]
  gitOperations: readonly GitOperation[]
  providers: readonly Provider[]
  usage: readonly Usage[]
  policy: Policy
  audit: readonly Audit[]
  workers: readonly Worker[]
  plugins: readonly Plugin[]
}>

export type DashboardState =
  | Readonly<{ kind: "loading" }>
  | Readonly<{ kind: "error"; message: string }>
  | Readonly<{ kind: "ready"; snapshot: DashboardSnapshot }>

const messageFrom = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message
  }
  return "Unexpected dashboard error"
}

export const useDashboard = (client: Api) => {
  const [state, setState] = useState<DashboardState>({ kind: "loading" })
  const [refreshing, setRefreshing] = useState(false)
  const [projectId, setProjectId] = useState<string | null>(null)
  const [targetId, setTargetId] = useState<string | null>(null)
  const [planId, setPlanId] = useState<string | null>(null)
  const requestSequence = useRef(0)

  const refresh = useCallback((): void => {
    const sequence = requestSequence.current + 1
    requestSequence.current = sequence
    setRefreshing(true)
    Promise.all([
      client.listProjects(),
      client.listPlans(),
      client.listAgents(),
      client.listDatabases(),
      client.listGitOperations(),
      client.listProviders(),
      client.listUsage(),
      client.getPolicy(),
      client.listAudit(),
      client.listWorkers(),
      client.listPlugins(),
    ])
      .then(
        async ([
          projects,
          plans,
          agents,
          databases,
          gitOperations,
          providers,
          usage,
          policy,
          audit,
          workers,
          plugins,
        ]) => {
          const activeProjectId = projects.some((project) => project.id === projectId)
            ? projectId
            : (projects[0]?.id ?? null)
          const activePlanId = plans.some((plan) => plan.id === planId)
            ? planId
            : (plans[0]?.id ?? null)
          const [targets, events] = await Promise.all([
            activeProjectId === null ? Promise.resolve([]) : client.listTargets(activeProjectId),
            activePlanId === null ? Promise.resolve([]) : client.listEvents(activePlanId),
          ])
          const activeTargetId = targets.some((target) => target.id === targetId)
            ? targetId
            : (targets[0]?.id ?? null)
          if (sequence !== requestSequence.current) {
            return
          }
          setProjectId(activeProjectId)
          setTargetId(activeTargetId)
          setPlanId(activePlanId)
          setState({
            kind: "ready",
            snapshot: {
              projects,
              targets,
              plans,
              events,
              agents,
              databases,
              gitOperations,
              providers,
              usage,
              policy,
              audit,
              workers,
              plugins,
            },
          })
        },
      )
      .catch((error: unknown) => {
        if (sequence === requestSequence.current) {
          setState({ kind: "error", message: messageFrom(error) })
        }
      })
      .finally(() => {
        if (sequence === requestSequence.current) {
          setRefreshing(false)
        }
      })
  }, [client, planId, projectId, targetId])

  const refreshRef = useRef(refresh)

  useEffect(() => {
    refreshRef.current = refresh
  }, [refresh])

  useEffect(() => {
    queueMicrotask(refreshRef.current)
    const interval = window.setInterval(() => refreshRef.current(), 2000)
    return () => window.clearInterval(interval)
  }, [])

  const snapshot = state.kind === "ready" ? state.snapshot : null
  const selectedProject = useMemo(
    () => snapshot?.projects.find((project) => project.id === projectId) ?? null,
    [projectId, snapshot],
  )
  const selectedTarget = useMemo(
    () => snapshot?.targets.find((target) => target.id === targetId) ?? null,
    [snapshot, targetId],
  )
  const selectedPlan = useMemo(
    () => snapshot?.plans.find((plan) => plan.id === planId) ?? null,
    [planId, snapshot],
  )

  const selectProject = useCallback((id: string) => {
    setProjectId(id)
    setTargetId(null)
  }, [])

  return {
    state,
    refreshing,
    refresh,
    projectId,
    targetId,
    planId,
    selectedProject,
    selectedTarget,
    selectedPlan,
    selectProject,
    selectTarget: setTargetId,
    selectPlan: setPlanId,
  }
}

export type DashboardController = ReturnType<typeof useDashboard>
