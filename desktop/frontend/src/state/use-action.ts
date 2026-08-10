import { useCallback, useRef, useState } from "react"

export type ActionState =
  | Readonly<{ kind: "idle" }>
  | Readonly<{ kind: "pending" }>
  | Readonly<{ kind: "success"; message: string }>
  | Readonly<{ kind: "error"; message: string }>

type SelectionOption = Readonly<{ id: string }>
const defaultActionChannel = "default"

export const reconcileSelection = (current: string, options: readonly SelectionOption[]): string =>
  options.some((option) => option.id === current) ? current : (options[0]?.id ?? "")

export const selectionChannel = (operation: string, selectionId: string): string =>
  `${operation}:${selectionId}`

const errorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message
  }
  return "Unexpected desktop error"
}

export const useAction = (refresh: () => void) => {
  const [state, setState] = useState<ActionState>({ kind: "idle" })
  const requestSequence = useRef(0)
  const channelSequences = useRef(new Map<string, number>())

  const run = useCallback(
    <Result>(
      action: () => Promise<Result>,
      message: string,
      onSuccess?: (result: Result) => void,
      channel: string = defaultActionChannel,
    ): void => {
      const sequence = requestSequence.current + 1
      requestSequence.current = sequence
      channelSequences.current.set(channel, sequence)
      setState({ kind: "pending" })
      action()
        .then((result) => {
          if (channelSequences.current.get(channel) !== sequence) return
          onSuccess?.(result)
          if (channelSequences.current.get(channel) !== sequence) return
          if (sequence === requestSequence.current) {
            setState({ kind: "success", message })
          }
          refresh()
        })
        .catch((error: unknown) => {
          if (
            channelSequences.current.get(channel) === sequence &&
            sequence === requestSequence.current
          ) {
            setState({ kind: "error", message: errorMessage(error) })
          }
        })
    },
    [refresh],
  )

  return { state, run }
}
