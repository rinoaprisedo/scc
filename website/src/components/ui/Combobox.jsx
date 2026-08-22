import { useEffect, useRef, useState } from 'react'
import { ChevronDown, X } from 'lucide-react'

// A searchable <select> — options are filtered as the user types, but the
// bound value is still just the picked option's `value` (a uuid), so it
// drops into react-hook-form via <Controller> exactly like a plain select.
function Combobox({ label, required, error, options, value, onChange, placeholder = 'Cari...', className = '' }) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const inputRef = useRef(null)

  const selected = options.find((o) => o.value === value)

  useEffect(() => {
    if (!open) setQuery(selected ? selected.label : '')
  }, [selected, open])

  const filtered = query.trim()
    ? options.filter((o) => o.label.toLowerCase().includes(query.trim().toLowerCase()))
    : options

  const handleSelect = (opt) => {
    onChange(opt.value)
    setQuery(opt.label)
    setOpen(false)
  }

  // Typing a search term and then clicking away (Save, another field, etc.)
  // without explicitly clicking a dropdown option used to silently discard
  // the typed text and leave the field empty — looked filled in, validated
  // as if it wasn't. Auto-commit the best match on blur instead.
  const closeAndCommit = () => {
    const trimmed = query.trim()
    const alreadyMatches = selected && selected.label.toLowerCase() === trimmed.toLowerCase()
    if (trimmed && filtered.length > 0 && !alreadyMatches) {
      const exact = filtered.find((o) => o.label.toLowerCase() === trimmed.toLowerCase())
      handleSelect(exact || filtered[0])
      return
    }
    setOpen(false)
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      closeAndCommit()
      inputRef.current?.blur()
    } else if (e.key === 'Escape') {
      setQuery(selected ? selected.label : '')
      setOpen(false)
    }
  }

  return (
    <div className={`relative ${className}`}>
      {label && (
        <span className="text-sm font-semibold text-navy">
          {label}
          {required && <span className="ml-0.5 text-red-500">*</span>}
        </span>
      )}
      <div className="relative mt-1.5">
        <input
          ref={inputRef}
          type="text"
          value={open ? query : selected?.label || ''}
          placeholder={placeholder}
          onFocus={() => {
            setOpen(true)
            setQuery('')
          }}
          onChange={(e) => setQuery(e.target.value)}
          onBlur={closeAndCommit}
          onKeyDown={handleKeyDown}
          className={`w-full rounded-xl border bg-white px-4 py-2.5 pr-9 text-sm text-text-primary outline-none transition focus:ring-2 ${
            error ? 'border-red-300 focus:ring-red-200' : 'border-surface-border focus:border-gold focus:ring-gold/20'
          }`}
        />
        {selected && !open ? (
          <button
            type="button"
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => {
              onChange('')
              setQuery('')
            }}
            aria-label="Clear"
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary hover:text-red-600"
          >
            <X size={15} />
          </button>
        ) : (
          <ChevronDown
            size={16}
            className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary"
          />
        )}

        {open && (
          <div className="absolute z-10 mt-1 max-h-56 w-full overflow-y-auto rounded-xl border border-surface-border bg-white py-1 shadow-lg">
            {filtered.length === 0 && <p className="px-4 py-2 text-sm text-text-secondary">Tidak ditemukan</p>}
            {filtered.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => handleSelect(opt)}
                className={`block w-full truncate px-4 py-2 text-left text-sm hover:bg-surface-bg ${
                  opt.value === value ? 'bg-gold/10 font-semibold text-navy' : 'text-text-primary'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        )}
      </div>
      {error && <p className="mt-1 text-xs font-medium text-red-600">{error}</p>}
    </div>
  )
}

export default Combobox
