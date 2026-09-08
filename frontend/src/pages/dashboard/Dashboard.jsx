import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Users, UserCheck, UserX, QrCode, Shirt } from 'lucide-react'
import { Menubar, MenubarLabel } from '../../components/ui/Menubar'
import EmptyState from '../../components/ui/EmptyState'
import Skeleton from '../../components/ui/Skeleton'
import { getDashboardSummary } from '../../api/dashboard'
import { attendanceLabel } from '../../utils/attendance'
import usePageActions from '../../hooks/usePageActions'

// Fixed categorical order (not cycled per-render) for charts with no
// inherent status meaning — dietary categories and QR gates are identity,
// not state, so they get their own hues instead of borrowing success/danger.
const CATEGORY_COLORS = ['#2a78d6', '#eb6834', '#1baf7a', '#eda100', '#e87ba4', '#008300', '#4a3aa7', '#e34948']

function formatNumber(n) {
  return new Intl.NumberFormat('id-ID').format(n ?? 0)
}

function StatTile({ icon: Icon, label, value, sublabel, tone = 'primary' }) {
  const toneClasses = {
    primary: 'bg-primary/10 text-primary',
    success: 'bg-success-bg text-success',
    neutral: 'bg-surface-hover text-text-secondary',
  }
  return (
    <div className="rounded-lg border border-surface-border bg-surface-card p-5 shadow-card">
      <div className="flex items-center gap-3">
        <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${toneClasses[tone]}`}>
          <Icon size={20} strokeWidth={2} />
        </div>
        <p className="text-sm font-medium text-text-secondary">{label}</p>
      </div>
      <p className="mt-4 text-3xl font-semibold text-text-primary">{formatNumber(value)}</p>
      {sublabel && <p className="mt-1 text-xs text-text-tertiary">{sublabel}</p>}
    </div>
  )
}

function BarRow({ label, count, total, color }) {
  const pct = total > 0 ? Math.round((count / total) * 100) : 0
  return (
    <div title={`${label}: ${formatNumber(count)} peserta (${pct}%)`}>
      <div className="mb-1.5 flex items-center justify-between gap-3 text-sm">
        <span className="flex items-center gap-2 text-text-secondary">
          <span className="h-2 w-2 shrink-0 rounded-full" style={{ backgroundColor: color }} />
          {label}
        </span>
        <span className="shrink-0 font-medium text-text-primary">
          {formatNumber(count)} <span className="font-normal text-text-tertiary">({pct}%)</span>
        </span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-surface-hover">
        <div
          className="h-full rounded-full transition-all"
          style={{ width: `${pct}%`, backgroundColor: color }}
        />
      </div>
    </div>
  )
}

// StockRow mirrors BarRow's layout but the bar reflects a size's remaining
// stock relative to the size with the most remaining (not a share of total
// peserta, which BarRow assumes). initialStock (the admin-set quota) shows
// next to the size label for context, while the bar/value on the right is
// always the computed remaining (initialStock minus how many peserta already
// picked it) — flagged with the same "Habis" wording and text-danger color
// the Blazer Sizes admin list page uses once remaining hits zero.
function StockRow({ label, initialStock, remaining, maxRemaining, color }) {
  const outOfStock = remaining <= 0
  const pct = maxRemaining > 0 ? Math.round((Math.max(remaining, 0) / maxRemaining) * 100) : 0
  return (
    <div title={outOfStock ? `${label}: stok habis` : `${label}: sisa ${formatNumber(remaining)} pcs`}>
      <div className="mb-1.5 flex items-center justify-between gap-3 text-sm">
        <span className="flex items-center gap-2 text-text-secondary">
          <span className="h-2 w-2 shrink-0 rounded-full" style={{ backgroundColor: color }} />
          {label}
          <span className="text-xs font-normal text-text-tertiary">(stok awal: {formatNumber(initialStock)})</span>
        </span>
        <span className={`shrink-0 font-medium ${outOfStock ? 'text-danger' : 'text-text-primary'}`}>
          {outOfStock ? 'Habis' : `${formatNumber(remaining)} pcs`}
        </span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-surface-hover">
        <div className="h-full rounded-full transition-all" style={{ width: `${pct}%`, backgroundColor: color }} />
      </div>
    </div>
  )
}

function ChartCard({ title, children }) {
  return (
    <div className="rounded-lg border border-surface-border bg-surface-card p-5 shadow-card">
      <h3 className="mb-4 text-sm font-semibold text-text-primary">{title}</h3>
      {children}
    </div>
  )
}

function Dashboard() {
  usePageActions(
    useMemo(
      () => (
        <Menubar>
          <MenubarLabel>Dashboard</MenubarLabel>
        </Menubar>
      ),
      [],
    ),
  )

  const { data, isLoading } = useQuery({
    queryKey: ['dashboard', 'summary'],
    queryFn: getDashboardSummary,
    staleTime: 60_000,
  })

  const summary = data?.data
  const total = summary?.total_peserta ?? 0

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-lg" />
          ))}
        </div>
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-64 rounded-lg" />
          ))}
        </div>
      </div>
    )
  }

  const loginPct = total > 0 ? Math.round(((summary.logged_in ?? 0) / total) * 100) : 0
  const blazerSizes = summary.blazer_sizes ?? []
  const maxRemainingStock = Math.max(1, ...blazerSizes.map((b) => Math.max(b.remaining, 0)))

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatTile icon={Users} label="Total Peserta" value={total} tone="primary" />
        <StatTile
          icon={UserCheck}
          label="Sudah Login"
          value={summary.logged_in}
          sublabel={`${loginPct}% dari total peserta`}
          tone="success"
        />
        <StatTile
          icon={UserX}
          label="Belum Login"
          value={summary.not_logged_in}
          sublabel={`${100 - loginPct}% dari total peserta`}
          tone="neutral"
        />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ChartCard title="Status Kehadiran">
          <div className="space-y-4">
            {(summary.attendance ?? []).map((row) => {
              const { text, variant } = attendanceLabel(row.key)
              return (
                <BarRow
                  key={row.key}
                  label={text}
                  count={row.count}
                  total={total}
                  color={variantColor(variant)}
                />
              )
            })}
          </div>
        </ChartCard>

        <ChartCard title="Kategori Pantangan Makan">
          <div className="space-y-4">
            {(summary.dietary ?? []).map((row, i) => (
              <BarRow key={row.key} label={row.label} count={row.count} total={total} color={CATEGORY_COLORS[i % CATEGORY_COLORS.length]} />
            ))}
          </div>
        </ChartCard>

        <ChartCard title="Stok Blazer per Ukuran">
          {blazerSizes.length === 0 ? (
            <EmptyState icon={Shirt} title="Belum ada data ukuran" message="Tambahkan ukuran di menu Blazer Sizes untuk mulai melacak stok." />
          ) : (
            <div className="space-y-4">
              {blazerSizes.map((row, i) => (
                <StockRow
                  key={row.size}
                  label={row.size}
                  initialStock={row.stock}
                  remaining={row.remaining}
                  maxRemaining={maxRemainingStock}
                  color={CATEGORY_COLORS[i % CATEGORY_COLORS.length]}
                />
              ))}
            </div>
          )}
        </ChartCard>

        <ChartCard title="Scan per QR Gate">
          {(summary.qr_gates ?? []).length === 0 ? (
            <EmptyState icon={QrCode} title="Belum ada QR gate" message="Buat QR gate di menu QR Gate untuk mulai melacak scan peserta." />
          ) : (
            <div className="space-y-4">
              {summary.qr_gates.map((row, i) => (
                <BarRow key={row.key} label={row.label} count={row.count} total={total} color={CATEGORY_COLORS[i % CATEGORY_COLORS.length]} />
              ))}
            </div>
          )}
        </ChartCard>
      </div>
    </div>
  )
}

// variantColor resolves a Badge-style variant name to the same hex the
// tailwind token points at, since BarRow needs a raw color for inline width
// styling rather than a class name.
const VARIANT_HEX = { success: '#2d6a4f', danger: '#c0392b', warning: '#92400e', primary: '#1d4ed8', neutral: '#a09e98' }
function variantColor(variant) {
  return VARIANT_HEX[variant] || VARIANT_HEX.neutral
}

export default Dashboard
