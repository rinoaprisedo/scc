import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Link2, MonitorPlay } from 'lucide-react'
import Table from '../../components/ui/Table'
import { Menubar, MenubarAction, MenubarLabel, MenubarSeparator } from '../../components/ui/Menubar'
import { getLeaderboard } from '../../api/leaderboard'
import usePageActions from '../../hooks/usePageActions'
import { DISPLAY_PATH, MEDALS, formatPoints } from './shared'

const displayURL = () => `${window.location.origin}${DISPLAY_PATH}`

// navigator.clipboard only exists in secure contexts (https/localhost), so a
// monitor-setup admin opening the panel over a plain-http LAN IP needs the
// legacy execCommand fallback.
async function copyText(text) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text)
    return
  }
  const el = document.createElement('textarea')
  el.value = text
  el.style.position = 'fixed'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.select()
  const ok = document.execCommand('copy')
  document.body.removeChild(el)
  if (!ok) throw new Error('copy failed')
}

function RankBadge({ rank }) {
  const medal = MEDALS[rank]
  if (medal) {
    return (
      <span
        className="inline-flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold shadow-card"
        style={{ background: medal.bg, color: medal.text }}
      >
        {rank}
      </span>
    )
  }
  return (
    <span className="inline-flex h-7 min-w-7 items-center justify-center rounded-full bg-surface-hover px-1.5 text-xs font-semibold text-text-secondary">
      {rank}
    </span>
  )
}

function Leaderboard() {
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [search, setSearch] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['leaderboard'],
    queryFn: getLeaderboard,
    refetchInterval: 15000,
  })

  const entries = useMemo(() => data?.data || [], [data])
  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return entries
    return entries.filter((e) => e.name.toLowerCase().includes(q) || e.ktp_number?.toLowerCase().includes(q))
  }, [entries, search])
  const pageRows = filtered.slice((page - 1) * limit, page * limit)

  const totalPoints = entries.reduce((sum, e) => sum + e.points, 0)
  const activeCount = entries.filter((e) => e.scan_count > 0).length

  usePageActions(
    useMemo(
      () => (
        <Menubar>
          <MenubarLabel>Leaderboard</MenubarLabel>
          <MenubarSeparator />
          <MenubarAction
            label="Copy Display Link"
            icon={Link2}
            onClick={() =>
              copyText(displayURL())
                .then(() => toast.success('Display link copied'))
                .catch(() => toast.error('Copy failed — ' + displayURL()))
            }
          />
          <MenubarAction label="Open Display" icon={MonitorPlay} onClick={() => window.open(DISPLAY_PATH, '_blank', 'noopener')} />
        </Menubar>
      ),
      [],
    ),
  )

  const columns = [
    { key: 'rank', label: 'Rank', render: (row) => <RankBadge rank={row.rank} /> },
    { key: 'name', label: 'Nama', render: (row) => <span className="font-medium text-text-primary">{row.name}</span> },
    { key: 'ktp_number', label: 'NIP', render: (row) => <span className="font-mono">{row.ktp_number || '-'}</span> },
    { key: 'scan_count', label: 'QR Discan', render: (row) => formatPoints(row.scan_count) },
    {
      key: 'points',
      label: 'Poin',
      render: (row) => <span className="font-semibold text-primary">{formatPoints(row.points)}</span>,
    },
  ]

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 rounded-lg border border-surface-border bg-surface-card p-5 shadow-card sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <MonitorPlay size={20} />
          </div>
          <div className="min-w-0">
            <p className="text-sm font-semibold text-text-primary">Fullscreen display</p>
            <p className="truncate font-mono text-xs text-text-tertiary">{displayURL()}</p>
          </div>
        </div>
        <div className="flex gap-6 text-sm">
          <div>
            <p className="text-xs text-text-tertiary">Peserta aktif</p>
            <p className="text-lg font-semibold text-text-primary">
              {formatPoints(activeCount)}
              <span className="text-sm font-normal text-text-tertiary"> / {formatPoints(entries.length)}</span>
            </p>
          </div>
          <div>
            <p className="text-xs text-text-tertiary">Total poin</p>
            <p className="text-lg font-semibold text-text-primary">{formatPoints(totalPoints)}</p>
          </div>
        </div>
      </div>

      <Table
        columns={columns}
        data={pageRows}
        loading={isLoading}
        search={search}
        onSearchChange={(v) => {
          setSearch(v)
          setPage(1)
        }}
        page={page}
        limit={limit}
        total={filtered.length}
        onPageChange={setPage}
        onLimitChange={(l) => {
          setLimit(l)
          setPage(1)
        }}
      />
    </div>
  )
}

export default Leaderboard
