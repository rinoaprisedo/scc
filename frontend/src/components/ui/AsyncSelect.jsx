import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Loader2, Search } from 'lucide-react'
import { cn } from '../../utils/cn'

// Searchable select backed by an API call instead of a static option list —
// e.g. picking one peserta out of a list too large to load in full. Debounces
// the query, keeps the label of the currently-selected option even while the
// dropdown is closed (loadOptions never runs for it), and closes on
// click-outside/Escape like the rest of this app's dropdowns.
function AsyncSelect({
  label,
  error,
  value,
  initialLabel,
  onChange,
  loadOptions,
  placeholder = 'Search...',
  getOptionLabel = (o) => o.label,
  getOptionSublabel,
  getOptionValue = (o) => o.value,
  className,
}) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [options, setOptions] = useState([])
  const [loading, setLoading] = useState(false)
  const [selectedLabel, setSelectedLabel] = useState(initialLabel || '')
  const containerRef = useRef(null)
  const debounceRef = useRef(null)

  // Deliberately not keyed on `value` — handleSelect below already sets
  // selectedLabel the instant a user picks an option, and value changes on
  // every keystroke of that same selection via setValue(shouldValidate).
  // Re-running this on `value` would immediately stomp the just-picked
  // label back to initialLabel (empty on create), leaving the trigger
  // showing blank even though the right uuid was actually selected.
  useEffect(() => {
    setSelectedLabel(initialLabel || '')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialLabel])

  useEffect(() => {
    if (!open) return undefined
    const onClickOutside = (e) => {
      if (!containerRef.current?.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [open])

  useEffect(() => {
    if (!open) return undefined
    clearTimeout(debounceRef.current)
    setLoading(true)
    debounceRef.current = setTimeout(async () => {
      try {
        const result = await loadOptions(query)
        setOptions(result || [])
      } finally {
        setLoading(false)
      }
    }, 300)
    return () => clearTimeout(debounceRef.current)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, query])

  const handleSelect = (opt) => {
    setSelectedLabel(getOptionLabel(opt))
    onChange(getOptionValue(opt), opt)
    setOpen(false)
    setQuery('')
  }

  return (
    <div className="flex flex-col gap-1" ref={containerRef}>
      {label && <label className="text-xs font-medium uppercase tracking-[.05em] text-text-secondary">{label}</label>}
      <div className={cn('relative', className)}>
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className={cn(
            'flex h-[42px] w-full items-center justify-between rounded-md border-[1.5px] border-surface-border bg-white px-3 text-left text-sm outline-none transition-colors duration-150 focus:border-primary',
            error && 'border-danger',
          )}
        >
          <span className={cn('truncate', !value && 'text-text-tertiary')}>{value ? selectedLabel : placeholder}</span>
          <ChevronDown size={16} strokeWidth={1.7} className="shrink-0 text-text-tertiary" />
        </button>

        {open && (
          <div className="absolute left-0 right-0 top-[calc(100%+4px)] z-20 rounded-md border border-surface-border bg-white shadow-md">
            <div className="flex items-center gap-2 border-b border-surface-border px-3 py-2">
              <Search size={14} className="shrink-0 text-text-tertiary" />
              <input
                autoFocus
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Type to search..."
                className="w-full text-sm outline-none"
              />
              {loading && <Loader2 size={14} className="shrink-0 animate-spin text-text-tertiary" />}
            </div>
            <div className="max-h-56 overflow-y-auto py-1">
              {!loading && options.length === 0 && (
                <p className="px-3 py-2 text-xs text-text-secondary">No results</p>
              )}
              {options.map((opt) => (
                <button
                  key={getOptionValue(opt)}
                  type="button"
                  onClick={() => handleSelect(opt)}
                  className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-surface-hover"
                >
                  <span className="text-text-primary">{getOptionLabel(opt)}</span>
                  {getOptionSublabel && (
                    <span className="text-xs text-text-secondary">{getOptionSublabel(opt)}</span>
                  )}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
      {error && <span className="text-xs text-danger">{error}</span>}
    </div>
  )
}

export default AsyncSelect
