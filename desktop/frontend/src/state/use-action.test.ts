import { beforeEach, describe, expect, it, mock } from "bun:test"

import type { ActionState } from "./use-action"

let renderedState: ActionState = { kind: "idle" }

mock.module("react", () => ({
  useCallback: <Callback>(callback: Callback): Callback => callback,
  useRef: <Value>(initialValue: Value): { current: Value } => ({ current: initialValue }),
  useState: (initialValue: ActionState): readonly [ActionState, (next: ActionState) => void] => [
    initialValue,
    (next) => {
      renderedState = next
    },
  ],
}))

const { reconcileSelection, selectionChannel, useAction } = await import("./use-action")

const deferred = <Value>() => {
  let resolve = (_value: Value): void => undefined
  let reject = (_reason: unknown): void => undefined
  const promise = new Promise<Value>((nextResolve, nextReject) => {
    resolve = nextResolve
    reject = nextReject
  })
  return { promise, reject, resolve }
}

describe("reconcileSelection", () => {
  it("retains a current selection that remains available", () => {
    // Given refreshed options that still contain the current selection.
    const options = [{ id: "first" }, { id: "current" }]

    // When the selection is reconciled.
    const result = reconcileSelection("current", options)

    // Then the current selection remains unchanged.
    expect(result).toBe("current")
  })

  it("chooses the first option when the current selection is invalid", () => {
    // Given refreshed options that replaced the current selection.
    const options = [{ id: "first" }, { id: "second" }]

    // When the selection is reconciled.
    const result = reconcileSelection("removed", options)

    // Then the first available option becomes current.
    expect(result).toBe("first")
  })

  it("clears the current selection when no options remain", () => {
    // Given an empty refreshed option list.
    const options: readonly SelectionOptionFixture[] = []

    // When the selection is reconciled.
    const result = reconcileSelection("removed", options)

    // Then no selection remains.
    expect(result).toBe("")
  })
})

type SelectionOptionFixture = Readonly<{ id: string }>

describe("useAction", () => {
  beforeEach(() => {
    renderedState = { kind: "idle" }
  })

  it("keeps the latest success when an older request succeeds later", async () => {
    // Given two overlapping actions that will complete newest first.
    const first = deferred<string>()
    const second = deferred<string>()
    const successfulResults: string[] = []
    let refreshCount = 0
    const action = useAction(() => {
      refreshCount += 1
    })
    action.run(
      () => first.promise,
      "First completed",
      (result) => successfulResults.push(result),
    )
    action.run(
      () => second.promise,
      "Second completed",
      (result) => successfulResults.push(result),
    )

    // When the latest action succeeds before the stale action.
    second.resolve("second")
    await second.promise
    first.resolve("first")
    await first.promise

    // Then only the latest action controls callbacks and state.
    expect(successfulResults).toEqual(["second"])
    expect(renderedState).toEqual({ kind: "success", message: "Second completed" })
    expect(refreshCount).toBe(1)
  })

  it("keeps the latest success when an older request fails later", async () => {
    // Given two overlapping actions that will complete newest first.
    const first = deferred<string>()
    const second = deferred<string>()
    let refreshCount = 0
    const action = useAction(() => {
      refreshCount += 1
    })
    action.run(() => first.promise, "First completed")
    action.run(() => second.promise, "Second completed")

    // When the latest action succeeds before the stale action fails.
    second.resolve("second")
    await second.promise
    first.reject(new Error("stale failure"))
    await first.promise.catch(() => undefined)

    // Then the stale failure cannot replace the latest status.
    expect(renderedState).toEqual({ kind: "success", message: "Second completed" })
    expect(refreshCount).toBe(1)
  })

  it("allows the same operation on different selections to complete independently", async () => {
    // Given overlapping database queries for different selections.
    const first = deferred<string>()
    const second = deferred<string>()
    const successfulResults: string[] = []
    let refreshCount = 0
    const action = useAction(() => {
      refreshCount += 1
    })
    action.run(
      () => first.promise,
      "First completed",
      (result) => successfulResults.push(result),
      selectionChannel("database-result", "database-a"),
    )
    action.run(
      () => second.promise,
      "Second completed",
      (result) => successfulResults.push(result),
      selectionChannel("database-result", "database-b"),
    )

    // When both actions complete newest first.
    second.resolve("second")
    await second.promise
    first.resolve("first")
    await first.promise

    // Then both channels complete, but only the globally latest status remains visible.
    expect(successfulResults).toEqual(["second", "first"])
    expect(renderedState).toEqual({ kind: "success", message: "Second completed" })
    expect(refreshCount).toBe(2)
  })
})
