import { createDeployApi } from "./deploy-api"
import { createOperationsApi } from "./operations-api"
import { createSystemApi } from "./system-api"
import { createRequester, type Transport, wailsTransport } from "./transport"

export const createApi = (transport: Transport) => {
  const request = createRequester(transport)
  return {
    ...createDeployApi(request),
    ...createOperationsApi(request),
    ...createSystemApi(request),
  }
}

export type Api = ReturnType<typeof createApi>

export const api = createApi(wailsTransport)
