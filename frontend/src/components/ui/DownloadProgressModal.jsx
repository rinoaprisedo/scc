import { Loader2 } from 'lucide-react'
import { createPortal } from 'react-dom'

function formatBytes(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i++
  }
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

// Standalone (not built on Modal.jsx) because a download in progress isn't
// meant to be dismissible via backdrop click/X — the underlying request
// keeps running either way, so closing early would just hide state without
// stopping anything. The backend streams the file without a Content-Length
// header (size isn't known upfront), so there's no percentage to show —
// just bytes received and speed, enough to tell the user it's still moving.
function DownloadProgressModal({ open, title, loaded = 0, rate = 0 }) {
  if (!open) return null

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/25 p-4 backdrop-blur-[2px]">
      <div className="w-full max-w-sm rounded-xl bg-surface-card p-6 shadow-md">
        <div className="flex flex-col items-center gap-3 text-center">
          <Loader2 size={28} className="animate-spin text-primary" />
          <p className="text-sm font-medium text-text-primary">{title}</p>
          <div className="h-1.5 w-full overflow-hidden rounded-full bg-surface-hover">
            <div className="h-full w-full animate-pulse rounded-full bg-primary/60" />
          </div>
          <p className="text-xs text-text-secondary">
            {formatBytes(loaded)} diunduh{rate > 0 ? ` — ${formatBytes(rate)}/s` : ''}
          </p>
        </div>
      </div>
    </div>,
    document.body,
  )
}

export default DownloadProgressModal
