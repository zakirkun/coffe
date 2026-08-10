import { CircleNotch, Tray } from "@phosphor-icons/react"
import type { MouseEventHandler, ReactNode } from "react"

type PanelProps = Readonly<{
  title: string
  actions?: ReactNode
  children: ReactNode
  variant?: "standard" | "recessed" | "approval"
}>

export const Panel = ({ title, actions, children, variant = "standard" }: PanelProps) => (
  <section className={`panel panel--${variant}`}>
    <header className="panel__header">
      <h2>{title}</h2>
      {actions === undefined ? null : <div className="action-cluster">{actions}</div>}
    </header>
    <div className="panel__body">{children}</div>
  </section>
)

type ActionButtonProps = Readonly<{
  children: ReactNode
  variant?: "primary" | "secondary" | "destructive"
  type?: "button" | "submit"
  disabled?: boolean
  busy?: boolean
  onClick?: MouseEventHandler<HTMLButtonElement>
  ariaLabel?: string
}>

export const ActionButton = ({
  children,
  variant = "secondary",
  type = "button",
  disabled = false,
  busy = false,
  onClick,
  ariaLabel,
}: ActionButtonProps) => (
  <button
    aria-label={ariaLabel}
    className={`button button--${variant}`}
    disabled={disabled || busy}
    onClick={onClick}
    type={type}
  >
    {busy ? <CircleNotch aria-hidden="true" className="spinner" size={16} /> : null}
    <span>{children}</span>
  </button>
)

const successStatuses = new Set(["ready", "succeeded", "healthy", "success"])
const warningStatuses = new Set(["draft", "approved", "pending", "running"])
const errorStatuses = new Set(["failed", "denied", "canceled", "stopped", "error"])

const statusTone = (status: string): "neutral" | "success" | "warning" | "error" => {
  const normalized = status.toLowerCase()
  if (successStatuses.has(normalized)) {
    return "success"
  }
  if (warningStatuses.has(normalized)) {
    return "warning"
  }
  if (errorStatuses.has(normalized)) {
    return "error"
  }
  return "neutral"
}

export const StatusBadge = ({ status }: Readonly<{ status: string }>) => (
  <span className={`status status--${statusTone(status)}`}>{status}</span>
)

type EmptyStateProps = Readonly<{
  title: string
  detail: string
}>

export const EmptyState = ({ title, detail }: EmptyStateProps) => (
  <div className="empty-state">
    <Tray aria-hidden="true" size={24} />
    <strong>{title}</strong>
    <span>{detail}</span>
  </div>
)

export const InlineError = ({ message }: Readonly<{ message: string }>) => (
  <p className="inline-error" role="alert">
    {message}
  </p>
)

export const LoadingState = ({ label = "Loading" }: Readonly<{ label?: string }>) => (
  <div aria-live="polite" className="loading-state">
    <CircleNotch aria-hidden="true" className="spinner" size={18} />
    <span>{label}</span>
  </div>
)
