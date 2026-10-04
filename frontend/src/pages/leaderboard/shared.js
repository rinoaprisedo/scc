export const DISPLAY_PATH = '/leaderboard/display'

export function formatPoints(n) {
  return new Intl.NumberFormat('id-ID').format(n ?? 0)
}

// Podium medal tones for ranks 1–3 — fixed metal colors rather than the
// runtime primary color, since gold/silver/bronze carry meaning on their own.
export const MEDALS = {
  1: { label: 'Gold', ring: '#e7b416', bg: 'linear-gradient(135deg,#fde68a,#e7b416 55%,#b7860b)', text: '#4a3500' },
  2: { label: 'Silver', ring: '#b8bcc4', bg: 'linear-gradient(135deg,#f3f4f6,#b8bcc4 55%,#8a8f98)', text: '#2b2e33' },
  3: { label: 'Bronze', ring: '#c27a3e', bg: 'linear-gradient(135deg,#f5c79b,#c27a3e 55%,#8c5223)', text: '#3b1f08' },
}

export function initials(name = '') {
  return (
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((w) => w[0].toUpperCase())
      .join('') || '?'
  )
}
