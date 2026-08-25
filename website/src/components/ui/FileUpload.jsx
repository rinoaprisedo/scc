import { useRef } from 'react'
import { UploadCloud, X } from 'lucide-react'

function FileUpload({ label, hint, error, required, className = '', preview, alt = 'Preview', onFileSelect, onClear }) {
  const inputRef = useRef(null)

  const handleChange = (e) => {
    const file = e.target.files?.[0]
    if (file) onFileSelect(file)
    e.target.value = ''
  }

  return (
    <div className={className}>
      {label && (
        <span className="text-sm font-semibold text-navy">
          {label}
          {required && <span className="ml-0.5 text-red-500">*</span>}
        </span>
      )}
      <div
        className={`mt-1.5 flex items-center gap-4 rounded-xl border border-dashed bg-white p-3 transition ${
          error ? 'border-red-300' : 'border-surface-border'
        }`}
      >
        {preview ? (
          <img src={preview} alt={alt} className="h-14 w-20 shrink-0 rounded-lg border border-surface-border object-cover" />
        ) : (
          <div className="flex h-14 w-20 shrink-0 items-center justify-center rounded-lg bg-surface-bg text-text-secondary">
            <UploadCloud size={22} strokeWidth={1.7} />
          </div>
        )}

        <div className="min-w-0 flex-1">
          <button
            type="button"
            onClick={() => inputRef.current?.click()}
            className="rounded-lg bg-navy px-3 py-1.5 text-xs font-semibold text-white transition hover:opacity-90"
          >
            {preview ? 'Ganti File' : 'Choose File'}
          </button>
          {hint && !error && <p className="mt-1.5 text-xs text-text-secondary">{hint}</p>}
          {error && <p className="mt-1.5 text-xs font-medium text-red-600">{error}</p>}
        </div>

        {preview && (
          <button
            type="button"
            onClick={onClear}
            aria-label="Remove file"
            className="shrink-0 rounded-full p-1.5 text-text-secondary transition hover:bg-surface-bg hover:text-red-600"
          >
            <X size={16} />
          </button>
        )}

        <input ref={inputRef} type="file" accept="image/png,image/jpeg,image/webp" className="hidden" onChange={handleChange} />
      </div>
    </div>
  )
}

export default FileUpload
