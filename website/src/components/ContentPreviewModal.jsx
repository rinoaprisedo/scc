import { X } from 'lucide-react'
import { fileURL } from '../utils/url'

function isPdf(path) {
  return !!path && path.toLowerCase().endsWith('.pdf')
}

function ContentPreviewModal({ title, path, onClose }) {
  const url = fileURL(path)
  const pdf = isPdf(path)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/80 p-4 backdrop-blur-sm">
      <div className="flex max-h-[90vh] w-full max-w-3xl flex-col overflow-hidden rounded-2xl bg-white shadow-2xl">
        <div className="flex shrink-0 items-center justify-between border-b border-surface-border px-6 py-4">
          <p className="text-base font-bold text-navy">{title}</p>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="text-text-secondary transition hover:text-text-primary"
          >
            <X size={20} />
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-auto bg-surface-bg">
          {!url ? (
            <div className="flex h-64 items-center justify-center px-6 text-center text-sm text-text-secondary">
              Konten belum tersedia. Silakan cek kembali nanti.
            </div>
          ) : pdf ? (
            <iframe src={url} title={title} className="h-[75vh] w-full" />
          ) : (
            <img src={url} alt={title} className="mx-auto max-h-[75vh] w-auto object-contain" />
          )}
        </div>
      </div>
    </div>
  )
}

export default ContentPreviewModal
