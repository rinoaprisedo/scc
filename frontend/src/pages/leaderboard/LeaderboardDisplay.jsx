import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Maximize, Trophy } from 'lucide-react'
import { getPublicLeaderboard } from '../../api/leaderboard'
import { MEDALS, formatPoints, initials } from './shared'

const LIMIT = 5
const REFRESH_MS = 10000

const glow = (alpha) => `rgb(var(--color-primary-rgb) / ${alpha})`

// All five winners share the same gold tone — no gold/silver/bronze split.
const GOLD = MEDALS[1]
const gold = (alpha) => `${GOLD.ring}${Math.round(alpha * 255).toString(16).padStart(2, '0')}`

function WinnerCard({ entry, place }) {
  return (
    <div
      className="flex h-full min-w-0 flex-1 flex-col items-center justify-center rounded-3xl px-[1vw] py-[3vh]"
      style={{
        background: `linear-gradient(180deg, ${gold(0.22)}, ${gold(0.03)})`,
        borderTop: `0.5vh solid ${GOLD.ring}`,
      }}
    >
      <span className="text-[9vh] font-black leading-none" style={{ color: GOLD.ring }}>
        {place}
      </span>
      <div
        className="mt-[3vh] flex h-[13vh] w-[13vh] shrink-0 items-center justify-center rounded-full text-[4.5vh] font-bold shadow-2xl"
        style={{
          background: GOLD.bg,
          color: GOLD.text,
          boxShadow: `0 0 0 0.6vh ${gold(0.2)}, 0 1vh 4vh ${gold(0.33)}`,
        }}
      >
        {entry ? initials(entry.name) : '–'}
      </div>
      <p className="mt-[3vh] line-clamp-2 min-h-[6.6vh] w-full text-center text-[2.8vh] font-semibold leading-tight text-white">
        {entry?.name || 'Belum ada'}
      </p>
      <p className="mt-[0.5vh] h-[2vh] font-mono text-[1.6vh] text-white/45">{entry?.ktp_number}</p>
      <p className="mt-[2vh] text-[4.5vh] font-bold tabular-nums" style={{ color: GOLD.ring }}>
        {formatPoints(entry?.points)}
        <span className="ml-[0.5vw] text-[1.8vh] font-medium text-white/50">poin</span>
      </p>
    </div>
  )
}

function Clock() {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(id)
  }, [])
  return <span className="tabular-nums">{now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}</span>
}

// Public, login-free page meant to run unattended on an event monitor — it
// sits outside AppLayout/ProtectedRoute and polls the unauthenticated
// /leaderboard/public endpoint. Sized in viewport units so it fills any
// screen without scrolling.
function LeaderboardDisplay() {
  const [isFullscreen, setIsFullscreen] = useState(!!document.fullscreenElement)

  const { data, isError, dataUpdatedAt } = useQuery({
    queryKey: ['leaderboard-public'],
    queryFn: () => getPublicLeaderboard(LIMIT),
    refetchInterval: REFRESH_MS,
    refetchIntervalInBackground: true,
    retry: true,
  })

  useEffect(() => {
    const onChange = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onChange)
    return () => document.removeEventListener('fullscreenchange', onChange)
  }, [])

  const entries = data?.data || []

  return (
    <div className="relative flex h-screen w-screen flex-col overflow-hidden bg-[#0b0d12] px-[3vw] py-[3vh] text-white">
      <div
        className="pointer-events-none absolute -top-[30vh] left-1/2 h-[70vh] w-[90vw] -translate-x-1/2 rounded-full blur-3xl"
        style={{ background: glow(0.28) }}
      />
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{ backgroundImage: 'radial-gradient(#fff 1px, transparent 1px)', backgroundSize: '3vh 3vh' }}
      />

      {/* 3-column grid (not justify-between) so the centered title stays put
          while the ticking clock / Live-Offline text changes width. */}
      <header className="relative grid shrink-0 grid-cols-[1fr_auto_1fr] items-center">
        <div />
        <div className="flex items-center gap-[1vw]">
          <Trophy className="h-[5vh] w-[5vh] text-primary" />
          <h1 className="text-[5vh] font-black uppercase tracking-[0.15em]">LEADERBOARD SCC 2026</h1>
        </div>
        <div className="flex items-center justify-end gap-[1vw] text-[2.2vh] font-medium text-white/70">
          <span className="relative flex h-[1.4vh] w-[1.4vh]">
            <span className={`absolute inline-flex h-full w-full rounded-full opacity-75 ${isError ? 'bg-red-500' : 'animate-ping bg-emerald-400'}`} />
            <span className={`relative inline-flex h-full w-full rounded-full ${isError ? 'bg-red-500' : 'bg-emerald-400'}`} />
          </span>
          {isError ? 'Offline' : 'Live'}
          <span className="text-white/30">·</span>
          <Clock />
        </div>
      </header>

      {entries.length === 0 ? (
        <div className="relative flex flex-1 flex-col items-center justify-center gap-[2vh] text-white/40">
          <Trophy className="h-[12vh] w-[12vh]" strokeWidth={1.2} />
          <p className="text-[3vh]">{data ? 'Belum ada peserta' : 'Memuat leaderboard...'}</p>
        </div>
      ) : (
        <main className="relative mt-[3vh] flex min-h-0 flex-1 gap-[1.5vw]">
          {Array.from({ length: LIMIT }, (_, i) => (
            <WinnerCard key={i + 1} place={i + 1} entry={entries[i]} />
          ))}
        </main>
      )}

      <footer className="relative mt-[2vh] flex shrink-0 items-center justify-between text-[1.6vh] text-white/35">
        <span>Poin diperbarui otomatis setiap {REFRESH_MS / 1000} detik</span>
        {dataUpdatedAt > 0 && (
          <span className="tabular-nums">
            Update terakhir {new Date(dataUpdatedAt).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
          </span>
        )}
      </footer>

      {!isFullscreen && (
        <button
          type="button"
          onClick={() => document.documentElement.requestFullscreen?.().catch(() => {})}
          className="absolute bottom-[2vh] left-1/2 inline-flex -translate-x-1/2 items-center gap-2 rounded-full border border-white/10 bg-white/5 px-4 py-2 text-xs font-medium text-white/60 backdrop-blur transition hover:bg-white/10 hover:text-white"
        >
          <Maximize size={14} />
          Fullscreen
        </button>
      )}
    </div>
  )
}

export default LeaderboardDisplay
