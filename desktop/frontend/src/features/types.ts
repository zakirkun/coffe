import type { Api } from "../api"
import type { DesktopApi } from "../api/desktop-client"
import type { DashboardController, DashboardSnapshot } from "../state/use-dashboard"

export type Services = Readonly<{
  api: Api
  desktop: DesktopApi
}>

export type ScreenProps = Readonly<{
  dashboard: DashboardController
  snapshot: DashboardSnapshot
  services: Services
}>
