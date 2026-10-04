import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Check, Lock, Plus, QrCode, Trash2, UserPen } from 'lucide-react'
import Modal from '../../components/ui/Modal'
import Skeleton from '../../components/ui/Skeleton'
import Input from '../../components/ui/Input'
import Button from '../../components/ui/Button'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import {
  getPesertaPointHistory,
  addPesertaManualPoint,
  deletePesertaManualPoint,
  deletePesertaScan,
  checkinPesertaGate,
} from '../../api/peserta'
import usePermission from '../../hooks/usePermission'

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

function ManualPointForm({ onSubmit, loading }) {
  const [name, setName] = useState('')
  const [points, setPoints] = useState('')
  const [errors, setErrors] = useState({})

  const handleSubmit = (e) => {
    e.preventDefault()
    const next = {}
    if (!name.trim()) next.name = 'Nama point wajib diisi'
    const n = Number(points)
    if (points === '' || !Number.isInteger(n) || n === 0) next.points = 'Jumlah point harus angka bulat, bukan 0'
    setErrors(next)
    if (Object.keys(next).length > 0) return
    onSubmit({ name: name.trim(), points: n }, () => {
      setName('')
      setPoints('')
    })
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-surface-border p-4">
      <p className="mb-3 text-sm font-semibold text-text-primary">Tambah Point Manual</p>
      <div className="grid gap-3 sm:grid-cols-[1fr_140px]">
        <Input
          label="Nama Point"
          placeholder="mis. Juara Kuis"
          value={name}
          onChange={(e) => setName(e.target.value)}
          error={errors.name}
          maxLength={255}
        />
        <Input
          label="Jumlah Point"
          type="number"
          step="1"
          placeholder="mis. 50"
          value={points}
          onChange={(e) => setPoints(e.target.value)}
          error={errors.points}
        />
      </div>
      <div className="mt-3 flex items-center justify-between gap-3">
        <p className="text-xs text-text-tertiary">Isi angka negatif untuk mengurangi point.</p>
        <Button type="submit" disabled={loading}>
          <Plus size={14} />
          {loading ? 'Menyimpan...' : 'Tambah'}
        </Button>
      </div>
    </form>
  )
}

// One button per QR gate — clicking records the same scan the participant
// would get from the camera, so each gate can only be checked in once.
function GateCheckinButtons({ gates, onCheckin, pendingUuid }) {
  if (gates.length === 0) return null
  return (
    <div className="rounded-lg border border-surface-border p-4">
      <p className="mb-3 text-sm font-semibold text-text-primary">Check-in QR Gate</p>
      <div className="grid gap-2 sm:grid-cols-2">
        {gates.map((gate) => {
          const done = gate.scanned
          const locked = gate.claimed_by_other
          const pending = pendingUuid === gate.uuid
          const disabled = done || locked || !!pendingUuid
          return (
            <button
              key={gate.uuid}
              type="button"
              disabled={disabled}
              onClick={() => onCheckin(gate)}
              title={
                done ? 'Sudah check-in' : locked ? 'QR one-time ini sudah diklaim peserta lain' : 'Klik untuk check-in'
              }
              className={`flex items-center justify-between gap-2 rounded-md border px-3 py-2.5 text-left text-sm transition-colors ${
                done
                  ? 'cursor-default border-success/30 bg-success-bg text-success'
                  : locked
                    ? 'cursor-not-allowed border-surface-border bg-surface-bg text-text-tertiary'
                    : 'border-primary/30 bg-primary/5 text-text-primary hover:border-primary hover:bg-primary/10 disabled:opacity-60'
              }`}
            >
              <span className="flex min-w-0 items-center gap-2">
                {done ? (
                  <Check size={15} className="shrink-0" />
                ) : locked ? (
                  <Lock size={15} className="shrink-0" />
                ) : (
                  <QrCode size={15} className="shrink-0 text-primary" />
                )}
                <span className="truncate font-medium">{gate.name}</span>
              </span>
              <span className="shrink-0 text-xs font-semibold">
                {pending ? 'Menyimpan...' : done ? 'Sudah' : locked ? 'Diklaim' : `+${gate.points}`}
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}

function PesertaPointHistoryModal({ peserta, onClose }) {
  const { canEdit } = usePermission()
  const queryClient = useQueryClient()
  const [deleting, setDeleting] = useState(null)

  const { data, isLoading } = useQuery({
    queryKey: ['peserta-points', peserta?.uuid],
    queryFn: () => getPesertaPointHistory(peserta.uuid),
    enabled: !!peserta,
  })
  const scans = data?.data?.scans || []
  const gates = data?.data?.gates || []
  const totalPoints = data?.data?.total_points || 0

  // total_points is also a column on the Peserta list, so both views refresh.
  const refresh = () => {
    queryClient.invalidateQueries({
      queryKey: ['peserta-points', peserta?.uuid],
    })
    queryClient.invalidateQueries({ queryKey: ['peserta'] })
  }

  const addMutation = useMutation({
    mutationFn: ({ payload }) => addPesertaManualPoint(peserta.uuid, payload),
    onSuccess: (_, { reset }) => {
      toast.success('Point ditambahkan')
      reset()
      refresh()
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Gagal menambahkan point'),
  })

  const checkinMutation = useMutation({
    mutationFn: (gate) => checkinPesertaGate(peserta.uuid, gate.uuid),
    onSuccess: (res) => {
      toast.success(`Check-in ${res.data?.gate_name} (+${res.data?.points_awarded})`)
      refresh()
    },
    onError: (err) => {
      toast.error(err.response?.data?.message || 'Gagal check-in')
      refresh()
    },
  })

  const deleteMutation = useMutation({
    // Deleting a scan entry frees its gate's check-in button again.
    mutationFn: (entry) =>
      entry.source === 'manual'
        ? deletePesertaManualPoint(peserta.uuid, entry.uuid)
        : deletePesertaScan(peserta.uuid, entry.uuid),
    onSuccess: () => {
      toast.success('Point dihapus')
      setDeleting(null)
      refresh()
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Gagal menghapus point'),
  })

  return (
    <>
      <Modal open={!!peserta} onClose={onClose} title={peserta ? `Point History — ${peserta.name}` : ''} size="md">
        {isLoading ? (
          <Skeleton className="h-40 w-full" />
        ) : (
          <div className="flex flex-col gap-4">
            <div className="rounded-lg bg-surface-bg px-4 py-3">
              <p className="text-xs font-medium uppercase tracking-wide text-text-secondary">Total Points</p>
              <p className="text-2xl font-bold text-text-primary">{totalPoints}</p>
            </div>

            {canEdit('peserta') && (
              <GateCheckinButtons
                gates={gates}
                onCheckin={(gate) => checkinMutation.mutate(gate)}
                pendingUuid={checkinMutation.isPending ? checkinMutation.variables?.uuid : null}
              />
            )}

            {canEdit('peserta') && (
              <ManualPointForm
                loading={addMutation.isPending}
                onSubmit={(payload, reset) => addMutation.mutate({ payload, reset })}
              />
            )}

            {scans.length === 0 ? (
              <p className="py-6 text-center text-sm text-text-secondary">Belum ada riwayat point.</p>
            ) : (
              <div className="flex flex-col gap-2">
                {scans.map((scan) => {
                  const isManual = scan.source === 'manual'
                  const Icon = isManual ? UserPen : QrCode
                  return (
                    <div
                      key={scan.uuid}
                      className="flex items-center justify-between gap-3 rounded-lg border border-surface-border px-4 py-3"
                    >
                      <div className="flex min-w-0 items-center gap-3">
                        <div
                          className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-md ${
                            isManual ? 'bg-warning-bg text-warning' : 'bg-primary/10 text-primary'
                          }`}
                          title={isManual ? 'Point manual' : 'Scan QR'}
                        >
                          <Icon size={15} />
                        </div>
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold text-text-primary">{scan.gate_name}</p>
                          <p className="text-xs text-text-secondary">
                            {isManual ? 'Manual · ' : 'Scan QR · '}
                            {formatDate(scan.scanned_at)}
                          </p>
                        </div>
                      </div>
                      <div className="flex shrink-0 items-center gap-2">
                        <span
                          className={`text-sm font-bold ${scan.points_awarded < 0 ? 'text-danger' : 'text-primary'}`}
                        >
                          {scan.points_awarded > 0 ? '+' : ''}
                          {scan.points_awarded}
                        </span>
                        {canEdit('peserta') && (
                          <button
                            onClick={() => setDeleting(scan)}
                            className="rounded p-1.5 text-danger hover:bg-danger/10"
                            title={isManual ? 'Hapus point manual' : 'Batalkan check-in (QR bisa di-check-in lagi)'}
                          >
                            <Trash2 size={15} />
                          </button>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        )}
      </Modal>

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleteMutation.mutate(deleting)}
        loading={deleteMutation.isPending}
        message={
          deleting?.source === 'manual'
            ? `Hapus point "${deleting?.gate_name}" (${deleting?.points_awarded})?`
            : `Batalkan check-in "${deleting?.gate_name}" (${deleting?.points_awarded} point)? QR ini bisa di-check-in lagi setelahnya.`
        }
      />
    </>
  )
}

export default PesertaPointHistoryModal
