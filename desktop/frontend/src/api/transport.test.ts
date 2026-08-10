import { describe, expect, it } from "bun:test"
import { z } from "zod"

import { ApiError, createRequester, type Transport } from "./transport"

describe("createRequester", () => {
  it("parses a successful in-process API response", async () => {
    let captured: unknown
    const transport: Transport = {
      request: async (input) => {
        captured = input
        return {
          status: 201,
          content_type: "application/json",
          body: '{"status":"draft"}',
        }
      },
    }
    const request = createRequester(transport)

    const result = await request({
      method: "POST",
      path: "/api/v1/plans/preview",
      body: { prompt: "deploy" },
      schema: z.object({ status: z.literal("draft") }),
    })

    expect(captured).toEqual({
      method: "POST",
      path: "/api/v1/plans/preview",
      body: '{"prompt":"deploy"}',
    })
    expect(result).toEqual({ status: "draft" })
  })

  it("converts an API error response into ApiError", async () => {
    const transport: Transport = {
      request: async () => ({
        status: 403,
        content_type: "application/json",
        body: '{"code":"policy_denied","message":"push disabled"}',
      }),
    }
    const request = createRequester(transport)

    const operation = request({ method: "GET", path: "/api/v1/policy", schema: z.object({}) })

    await expect(operation).rejects.toBeInstanceOf(ApiError)
    await expect(operation).rejects.toHaveProperty("message", "push disabled")
  })
})
