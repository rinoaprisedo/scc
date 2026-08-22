import { forwardRef } from 'react'

// forwardRef is required here — react-hook-form's register() hands back a
// ref that must land on the real <input> DOM node. Without forwardRef, a
// plain function component silently drops that ref, the field never
// actually registers, and it reports as "Required" (zod's own default
// message for an undefined value) even when the box visibly has text in it.
const Input = forwardRef(function Input(
  { label, error, required, endAdornment, className = '', type = 'text', ...props },
  ref,
) {
  return (
    <label className={`block ${className}`}>
      {label && (
        <span className="text-sm font-semibold text-navy">
          {label}
          {required && <span className="ml-0.5 text-red-500">*</span>}
        </span>
      )}
      <div className="relative mt-1.5">
        <input
          ref={ref}
          type={type}
          className={`w-full rounded-xl border bg-white px-4 py-2.5 text-sm text-text-primary outline-none transition focus:ring-2 ${
            endAdornment ? 'pr-10' : ''
          } ${error ? 'border-red-300 focus:ring-red-200' : 'border-surface-border focus:border-gold focus:ring-gold/20'}`}
          {...props}
        />
        {endAdornment && <div className="absolute right-3 top-1/2 -translate-y-1/2">{endAdornment}</div>}
      </div>
      {error && <p className="mt-1 text-xs font-medium text-red-600">{error}</p>}
    </label>
  )
})

export default Input
