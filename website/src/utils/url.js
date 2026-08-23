const s3BaseURL = (import.meta.env.VITE_S3_BASE_URL || '').replace(/\/+$/, '')

// The backend stores each upload's relative object key (e.g.
// "qris_cross_border/abc123.png") — files are served directly from
// S3-compatible object storage, not a backend-hosted /uploads route.
export function fileURL(path) {
  if (!path) return null
  if (/^https?:\/\//.test(path)) return path
  return `${s3BaseURL}/${path.replace(/^\/+/, '')}`
}
