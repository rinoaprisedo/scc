const VARIANTS = {
  primary:
    'bg-gradient-to-r from-gold-light via-gold to-gold-dark text-navy-dark shadow-md hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed',
  secondary: 'bg-surface-bg text-text-primary hover:bg-surface-border disabled:opacity-60',
}

function Button({ variant = 'primary', className = '', ...props }) {
  return (
    <button
      className={`rounded-lg py-3 text-sm font-bold uppercase tracking-wide transition ${VARIANTS[variant]} ${className}`}
      {...props}
    />
  )
}

export default Button
