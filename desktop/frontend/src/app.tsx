import {
  Archive,
  CloudArrowUp,
  Code,
  Database,
  GitBranch,
  HardDrives,
  Pulse,
  Robot,
  ShieldCheck,
  Stack,
} from "@phosphor-icons/react"
import { useState } from "react"

import { api } from "./api"
import { desktopApi } from "./api/desktop-client"
import { ActionButton, InlineError, LoadingState, StatusBadge } from "./components/primitives"
import { AgentsScreen } from "./features/agents-screen"
import { AIPolicyScreen } from "./features/ai-policy-screen"
import { AuditScreen } from "./features/audit-screen"
import { DatabasesScreen } from "./features/databases-screen"
import { DeployScreen } from "./features/deploy-screen"
import { GitScreen } from "./features/git-screen"
import { IntegrationsScreen } from "./features/integrations-screen"
import { MaintenanceScreen } from "./features/maintenance-screen"
import { OverviewScreen } from "./features/overview-screen"
import { RegistryScreen } from "./features/registry-screen"
import type { Services } from "./features/types"
import { useDashboard } from "./state/use-dashboard"

const navigation = [
  { id: "overview", label: "Overview", icon: Pulse },
  { id: "registry", label: "Projects", icon: Stack },
  { id: "deploy", label: "Deploy", icon: CloudArrowUp },
  { id: "agents", label: "Agents", icon: Robot },
  { id: "databases", label: "Databases", icon: Database },
  { id: "git", label: "Git", icon: GitBranch },
  { id: "ai", label: "AI & policy", icon: ShieldCheck },
  { id: "integrations", label: "Integrations", icon: Code },
  { id: "maintenance", label: "Maintenance", icon: Archive },
  { id: "audit", label: "Audit", icon: HardDrives },
] as const

type Section = (typeof navigation)[number]["id"]

const services: Services = { api, desktop: desktopApi }

export const App = () => {
  const [section, setSection] = useState<Section>("overview")
  const dashboard = useDashboard(api)
  const current = navigation.find((item) => item.id === section)

  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>
      <nav aria-label="Primary" className="app-navigation">
        <div className="brand">
          <span className="brand__mark">C</span>
          <span>Coffe</span>
        </div>
        <div className="app-navigation__items">
          {navigation.map(({ id, label, icon: Icon }) => (
            <button
              className={section === id ? "nav-item nav-item--active" : "nav-item"}
              key={id}
              onClick={() => setSection(id)}
              type="button"
            >
              <Icon aria-hidden="true" size={18} />
              <span>{label}</span>
            </button>
          ))}
        </div>
        <p className="navigation-note">Local state only</p>
      </nav>

      <aside className="project-rail">
        <header>
          <span className="meta-label">Workspace</span>
          <StatusBadge status={dashboard.refreshing ? "running" : "ready"} />
        </header>
        {dashboard.state.kind === "ready" ? (
          <>
            <div className="rail-section">
              <h2>Projects</h2>
              {dashboard.state.snapshot.projects.map((project) => (
                <button
                  className={
                    dashboard.projectId === project.id ? "rail-item rail-item--active" : "rail-item"
                  }
                  key={project.id}
                  onClick={() => dashboard.selectProject(project.id)}
                  type="button"
                >
                  <span>{project.name}</span>
                  <small className="technical">{project.root_path}</small>
                </button>
              ))}
            </div>
            <div className="rail-section">
              <h2>Targets</h2>
              {dashboard.state.snapshot.targets.map((target) => (
                <button
                  className={
                    dashboard.targetId === target.id ? "rail-item rail-item--active" : "rail-item"
                  }
                  key={target.id}
                  onClick={() => dashboard.selectTarget(target.id)}
                  type="button"
                >
                  <span>{target.name}</span>
                  <small>{target.host}</small>
                </button>
              ))}
            </div>
          </>
        ) : null}
      </aside>

      <main className="app-main" id="main-content">
        <header className="titlebar">
          <div>
            <span className="meta-label">Control plane</span>
            <h1>{current?.label ?? "Coffe"}</h1>
          </div>
          <ActionButton busy={dashboard.refreshing} onClick={dashboard.refresh}>
            Refresh
          </ActionButton>
        </header>
        <div className="app-main__body">
          {dashboard.state.kind === "loading" ? (
            <LoadingState label="Opening local workspace" />
          ) : null}
          {dashboard.state.kind === "error" ? (
            <InlineError message={dashboard.state.message} />
          ) : null}
          {dashboard.state.kind === "ready"
            ? renderSection(section, dashboard, dashboard.state.snapshot)
            : null}
        </div>
      </main>
    </div>
  )
}

const renderSection = (
  section: Section,
  dashboard: ReturnType<typeof useDashboard>,
  snapshot: Extract<ReturnType<typeof useDashboard>["state"], { kind: "ready" }>["snapshot"],
) => {
  const props = { dashboard, snapshot, services }
  switch (section) {
    case "overview":
      return <OverviewScreen {...props} />
    case "registry":
      return <RegistryScreen {...props} />
    case "deploy":
      return <DeployScreen {...props} />
    case "agents":
      return <AgentsScreen {...props} />
    case "databases":
      return <DatabasesScreen {...props} />
    case "git":
      return <GitScreen {...props} />
    case "ai":
      return <AIPolicyScreen {...props} />
    case "integrations":
      return <IntegrationsScreen {...props} />
    case "maintenance":
      return <MaintenanceScreen {...props} />
    case "audit":
      return <AuditScreen {...props} />
  }
}
