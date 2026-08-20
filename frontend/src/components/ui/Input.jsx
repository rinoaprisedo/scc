import { forwardRef } from 'react'
import { cn } from '../../utils/cn'

const Input = forwardRef(function Input({ label, error, className, ...props }, ref) {
  return (
    <div className="flex flex-col gap-1">
      {label && (
        <label className="text-xs font-medium uppercase tracking-[.05em] text-text-secondary">{label}</label>
      )}
      <input
        ref={ref}
        className={cn(
          'h-[42px] rounded-md border-[1.5px] border-surface-border bg-white px-3 text-sm text-text-primary outline-none transition-colors duration-150 focus:border-primary',
          error && 'border-danger focus:border-danger',
          className,
        )}
        {...props}
      />
      {error && <span className="text-xs text-danger">{error}</span>}
    </div>
  )
})

export default Input
