const timestampFormatter = new Intl.DateTimeFormat(undefined, {
  month: "short",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
})

export const formatTimestamp = (value: string): string => timestampFormatter.format(new Date(value))

export const shortID = (value: string): string => value.slice(0, 8)
