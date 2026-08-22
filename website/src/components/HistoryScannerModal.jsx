import { useEffect, useState } from 'react'
import { X, Trophy } from 'lucide-react'
import { getMyScans } from '../api/qrGate'

function formatDate(value) {
  if (!value) return null
  return new Date(value).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function HistoryScannerModal({ onClose }) {
  const [loading, setLoading] = useState(true)
  const [scans, setScans] = useState([])
  const [totalPoints, setTotalPoints] = useState(0)

  useEffect(() => {
    getMyScans()
      .then((res) => {
        setScans(res.data?.scans || [])
        setTotalPoints(res.data?.total_points || 0)
      })
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 py-8 backdrop-blur-sm">
      <div className="flex max-h-[88vh] w-full max-w-lg flex-col overflow-hidden rounded-2xl bg-white shadow-2xl">
        <div className="relative shrink-0 bg-gradient-to-br from-navy via-navy to-navy-light px-8 pb-8 pt-8 text-center">
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="absolute right-5 top-5 text-white/70 transition hover:text-white"
          >
            <X size={20} />
          </button>
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-gold-light via-gold to-gold-dark text-navy-dark shadow-md ring-4 ring-white/20">
            <Trophy size={26} />
          </div>
          <p className="mt-3 text-2xl font-bold text-white">{totalPoints} Poin</p>
          <p className="text-xs font-medium uppercase tracking-widest text-gold-light">History Scanner</p>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-6 py-6">
          {loading ? (
            <p className="py-10 text-center text-sm text-text-secondary">Memuat...</p>
          ) : scans.length === 0 ? (
            <p className="py-10 text-center text-sm text-text-secondary">Belum ada QR yang berhasil discan</p>
          ) : (
            <div className="flex flex-col gap-3">
              {scans.map((scan) => (
                <div
                  key={scan.uuid}
                  className="flex items-center justify-between rounded-lg border border-surface-border px-4 py-3"
                >
                  <div>
                    <p className="text-sm font-semibold text-text-primary">{scan.gate_name || 'QR Gate'}</p>
                    <p className="text-xs text-text-secondary">{formatDate(scan.scanned_at)}</p>
                  </div>
                  <span className="text-sm font-bold text-gold-dark">+{scan.points_awarded}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default HistoryScannerModal
