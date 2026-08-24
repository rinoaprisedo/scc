import { useEffect, useState } from 'react'

// Naive local-time parse (no timezone suffix), same convention as
// utils/deadline.js — target reached is simply "midnight local time" on the
// given date, no need to reconcile against a server clock for a marketing
// countdown like this.
function getRemaining(target) {
  const diff = Math.max(0, target.getTime() - Date.now())
  const totalSeconds = Math.floor(diff / 1000)
  return {
    days: Math.floor(totalSeconds / 86400),
    hours: Math.floor((totalSeconds % 86400) / 3600),
    minutes: Math.floor((totalSeconds % 3600) / 60),
    seconds: totalSeconds % 60,
    done: diff === 0,
  }
}

const UNITS = [
  { key: 'days', label: 'Hari' },
  { key: 'hours', label: 'Jam' },
  { key: 'minutes', label: 'Menit' },
  { key: 'seconds', label: 'Detik' },
]

function Countdown({ target, className = '' }) {
  const [remaining, setRemaining] = useState(() => getRemaining(target))

  useEffect(() => {
    const id = setInterval(() => setRemaining(getRemaining(target)), 1000)
    return () => clearInterval(id)
  }, [target])

  if (remaining.done) return null

  return (
    <div className={`flex items-center justify-center gap-2 sm:gap-3 ${className}`}>
      {UNITS.map(({ key, label }) => (
        <div
          key={key}
          className="flex w-14 flex-col items-center rounded-lg border border-gold/30 bg-navy-light/60 py-2 backdrop-blur-sm sm:w-16"
        >
          <span className="font-mono text-xl font-bold text-gold sm:text-2xl">
            {String(remaining[key]).padStart(2, '0')}
          </span>
          <span className="text-[10px] font-semibold uppercase tracking-wider text-white/70">{label}</span>
        </div>
      ))}
    </div>
  )
}

export default Countdown
