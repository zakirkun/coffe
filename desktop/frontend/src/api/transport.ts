import { z } from "zod"

const methodSchema = z.union([z.literal("GET"), z.literal("POST"), z.literal("PUT")])

const bridgeRequestSchema = z.object({
  method: methodSchema,
  path: z.string().startsWith("/"),
  body: z.string(),
})

const bridgeResponseSchema = z.object({
  status: z.number().int(),
  content_type: z.string(),
  body: z.string(),
})

const apiErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
})

export type BridgeRequest = z.infer<typeof bridgeRequestSchema>

export type Transport = {
  readonly request: (input: BridgeRequest) => Promise<unknown>
}

type RequestSpec<Result> = {
  readonly method: z.infer<typeof methodSchema>
  readonly path: string
  readonly schema: z.ZodType<Result>
  readonly body?: unknown
}

export type Requester = <Result>(spec: RequestSpec<Result>) => Promise<Result>

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.code = code
  }
}

export const createRequester = (transport: Transport): Requester => {
  return async <Result>(spec: RequestSpec<Result>): Promise<Result> => {
    const rawResponse = await transport.request(
      bridgeRequestSchema.parse({
        method: spec.method,
        path: spec.path,
        body: encodeBody(spec.body),
      }),
    )
    const response = bridgeResponseSchema.parse(rawResponse)
    const decoded = decodeBody(response.body, response.status)
    if (response.status < 200 || response.status >= 300) {
      const failure = apiErrorSchema.safeParse(decoded)
      if (failure.success) {
        throw new ApiError(response.status, failure.data.code, failure.data.message)
      }
      throw new ApiError(response.status, "request_failed", `Request failed (${response.status})`)
    }
    return spec.schema.parse(decoded)
  }
}

const encodeBody = (body: unknown): string => {
  if (body === undefined) {
    return ""
  }
  const encoded = JSON.stringify(body)
  if (encoded === undefined) {
    throw new ApiError(0, "invalid_request", "Request body is not JSON serializable")
  }
  return encoded
}

const decodeBody = (body: string, status: number): unknown => {
  try {
    return JSON.parse(body)
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new ApiError(status, "invalid_response", "Desktop API returned invalid JSON")
    }
    throw error
  }
}

declare global {
  interface CoffeDesktopApp {
    readonly Request: (input: BridgeRequest) => Promise<unknown>
  }

  interface Window {
    readonly go: {
      readonly main: {
        readonly App: CoffeDesktopApp
      }
    }
  }
}

export const wailsTransport: Transport = {
  request: (input) => window.go.main.App.Request(input),
}
