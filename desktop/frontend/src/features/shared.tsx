import { InlineError } from "../components/primitives"
import type { ActionState } from "../state/use-action"

export const ActionStatus = ({ state }: Readonly<{ state: ActionState }>) => {
  switch (state.kind) {
    case "idle":
    case "pending":
      return null
    case "success":
      return (
        <p className="form-status form-status--success" role="status">
          {state.message}
        </p>
      )
    case "error":
      return <InlineError message={state.message} />
  }
}
