// Compares against a <input type="datetime-local"> value ("" = no deadline).
// Naive datetime strings (no timezone suffix) parse as local wall-clock time
// in JS by spec, matching the backend's time.ParseInLocation(..., time.Local)
// — so this stays consistent with the server-side enforcement in
// backend/modules/peserta/handler.go without needing to share code across languages.
export function isPastDeadline(value) {
  if (!value) return false
  const deadline = new Date(value)
  if (Number.isNaN(deadline.getTime())) return false
  return new Date() > deadline
}
