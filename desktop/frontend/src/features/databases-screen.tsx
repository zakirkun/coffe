import { DatabaseQuery } from "./database-query"
import { DatabaseRegistration } from "./database-registration"
import type { ScreenProps } from "./types"

export const DatabasesScreen = (props: ScreenProps) => (
  <div className="screen-grid screen-grid--two">
    <DatabaseRegistration {...props} />
    <DatabaseQuery {...props} />
  </div>
)
