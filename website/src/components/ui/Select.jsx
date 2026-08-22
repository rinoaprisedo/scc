import { forwardRef } from 'react'
import { ChevronDown } from 'lucide-react'

// See Input.jsx for why forwardRef is required — register()'s ref must
// reach the real <select> or react-hook-form won't pick up its value.
const Select = forwardRef(function Select({ label, error, required, className = '', children, ...props }, ref) {
  return (
    <label className={`block ${className}`}>
      {label && (
        <span className="text-sm font-semibold text-navy">
          {label}
          {required && <span className="ml-0.5 text-red-500">*</span>}
        </span>
      )}
      <div className="relative mt-1.5">
        <select
          ref={ref}
          className={`w-full appearance-none rounded-xl border bg-white px-4 py-2.5 pr-9 text-sm text-text-primary outline-none transition focus:ring-2 ${
            error ? 'border-red-300 focus:ring-red-200' : 'border-surface-border focus:border-gold focus:ring-gold/20'
          }`}
          {...props}
        >
          {children}
        </select>
        <ChevronDown size={16} className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary" />
      </div>
      {error && <p className="mt-1 text-xs font-medium text-red-600">{error}</p>}
    </label>
  )
})

export default Select
