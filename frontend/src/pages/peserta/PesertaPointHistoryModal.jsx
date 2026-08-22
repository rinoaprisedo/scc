import { useEffect, useState } from 'react'
import Modal from '../../components/ui/Modal'
import Skeleton from '../../components/ui/Skeleton'
import { getPesertaPointHistory } from '../../api/peserta'

function formatDate(value) {
  if (!value) return '-'
  return new Date(value).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function PesertaPointHistoryModal({ peserta, onClose }) {
  const [loading, setLoading] = useState(true)
  const [scans, setScans] = useState([])
  const [totalPoints, setTotalPoints] = useState(0)

  useEffect(() => {
    if (!peserta) return
    setLoading(true)
    getPesertaPointHistory(peserta.uuid)
      .then((res) => {
        setScans(res.data?.scans || [])
        setTotalPoints(res.data?.total_points || 0)
      })
      .finally(() => setLoading(false))
  }, [peserta])

  return (
    <Modal open={!!peserta} onClose={onClose} title={peserta ? `Point History — ${peserta.name}` : ''} size="md">
      {loading ? (
        <Skeleton className="h-40 w-full" />
      ) : (
        <div className="flex flex-col gap-4">
          <div className="rounded-lg bg-surface-bg px-4 py-3">
            <p className="text-xs font-medium uppercase tracking-wide text-text-secondary">Total Points</p>
            <p className="text-2xl font-bold text-text-primary">{totalPoints}</p>
          </div>

          {scans.length === 0 ? (
            <p className="py-6 text-center text-sm text-text-secondary">Belum ada QR gate yang berhasil discan.</p>
          ) : (
            <div className="flex flex-col gap-2">
              {scans.map((scan) => (
                <div
                  key={scan.uuid}
                  className="flex items-center justify-between rounded-lg border border-surface-border px-4 py-3"
                >
                  <div>
                    <p className="text-sm font-semibold text-text-primary">{scan.gate_name}</p>
                    <p className="text-xs text-text-secondary">{formatDate(scan.scanned_at)}</p>
                  </div>
                  <span className="text-sm font-bold text-primary">+{scan.points_awarded}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}

export default PesertaPointHistoryModal
