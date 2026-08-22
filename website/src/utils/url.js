const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8880/api/v1'
export const apiOrigin = apiBase.replace(/\/api\/v1\/?$/, '')

// The backend stores relative storage paths (e.g. "ktp/abc123.png"); files
// are served by the backend at /uploads/<path>, not the website's own origin.
export function fileURL(path) {
  if (!path) return null
  if (/^https?:\/\//.test(path)) return path
  return `${apiOrigin}/uploads/${path.replace(/^\/+/, '')}`
}
