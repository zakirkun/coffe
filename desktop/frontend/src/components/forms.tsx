import { FolderOpen } from "@phosphor-icons/react"
import { type ChangeEventHandler, type ReactNode, useId } from "react"

import { ActionButton } from "./primitives"

type FieldFrameProps = Readonly<{
  label: string
  id: string
  helper?: string | undefined
  children: ReactNode
}>

export const FieldFrame = ({ label, id, helper, children }: FieldFrameProps) => (
  <div className="field">
    <label className="field__label" htmlFor={id}>
      {label}
    </label>
    {children}
    {helper === undefined ? null : <span className="field__helper">{helper}</span>}
  </div>
)

type TextFieldProps = Readonly<{
  label: string
  value: string
  onChange: ChangeEventHandler<HTMLInputElement>
  type?: "text" | "password" | "url" | "number"
  placeholder?: string
  helper?: string | undefined
  required?: boolean
}>

export const TextField = ({
  label,
  value,
  onChange,
  type = "text",
  placeholder,
  helper,
  required = false,
}: TextFieldProps) => {
  const id = useId()
  return (
    <FieldFrame helper={helper} id={id} label={label}>
      <input
        id={id}
        onChange={onChange}
        placeholder={placeholder}
        required={required}
        type={type}
        value={value}
      />
    </FieldFrame>
  )
}

type SelectFieldProps = Readonly<{
  label: string
  value: string
  onChange: ChangeEventHandler<HTMLSelectElement>
  children: ReactNode
}>

export const SelectField = ({ label, value, onChange, children }: SelectFieldProps) => {
  const id = useId()
  return (
    <FieldFrame id={id} label={label}>
      <select id={id} onChange={onChange} value={value}>
        {children}
      </select>
    </FieldFrame>
  )
}

type PathFieldProps = Readonly<{
  label: string
  value: string
  onBrowse: () => void
}>

export const PathField = ({ label, value, onBrowse }: PathFieldProps) => {
  const id = useId()
  return (
    <FieldFrame id={id} label={label}>
      <div className="path-field">
        <input className="technical" id={id} readOnly value={value} />
        <ActionButton ariaLabel={`Browse for ${label}`} onClick={onBrowse}>
          <FolderOpen aria-hidden="true" size={16} />
          Browse
        </ActionButton>
      </div>
    </FieldFrame>
  )
}

type CheckboxFieldProps = Readonly<{
  label: string
  checked: boolean
  onChange: ChangeEventHandler<HTMLInputElement>
}>

export const CheckboxField = ({ label, checked, onChange }: CheckboxFieldProps) => (
  <label className="checkbox-field">
    <input checked={checked} onChange={onChange} type="checkbox" />
    <span>{label}</span>
  </label>
)
