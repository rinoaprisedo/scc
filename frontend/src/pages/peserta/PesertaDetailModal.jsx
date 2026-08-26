import { useState } from 'react'
import { X, ZoomIn } from 'lucide-react'
import Modal from '../../components/ui/Modal'
import Badge from '../../components/ui/Badge'
import { fileURL } from '../../utils/url'
import { attendanceLabel } from '../../utils/attendance'

const statusVariant = { active: 'success', inactive: 'neutral', suspended: 'danger' }

function Field({ label, value }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-wide text-text-secondary">{label}</p>
      <p className="mt-0.5 text-sm font-medium text-text-primary">{value || '-'}</p>
    </div>
  )
}

function formatDate(value) {
  if (!value) return null
  return new Date(value).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })
}

function fullName(p) {
  return [p.title, p.first_name, p.middle_name, p.last_name].filter(Boolean).join(' ') || p.name
}

function PesertaDetailModal({ open, onClose, peserta }) {
  const [zoomedUrl, setZoomedUrl] = useState(null)

  if (!peserta) return null
  const attendance = attendanceLabel(peserta.attendance_status)
  const originCity = peserta.origin_city
    ? `${peserta.origin_city.name}${peserta.origin_city.province ? `, ${peserta.origin_city.province}` : ''}`
    : peserta.origin_city_other
  const dietary = peserta.dietary_restriction
  const ktpUrl = peserta.ktp_file ? fileURL(peserta.ktp_file) : null
  const passportUrl = peserta.passport_file ? fileURL(peserta.passport_file) : null

  return (
    <Modal open={open} onClose={onClose} title="Detail Peserta" size="lg">
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-primary/10 text-lg font-semibold text-primary">
            {peserta.avatar ? (
              <img src={peserta.avatar} alt={peserta.name} className="h-14 w-14 rounded-full object-cover" />
            ) : (
              peserta.name?.charAt(0).toUpperCase()
            )}
          </div>
          <div>
            <p className="text-base font-semibold text-text-primary">{fullName(peserta)}</p>
            <div className="mt-1.5 flex gap-1.5">
              <Badge variant={statusVariant[peserta.status] || 'neutral'}>{peserta.status}</Badge>
              <Badge variant={attendance.variant}>{attendance.text}</Badge>
            </div>
          </div>
        </div>

        <div>
          <p className="mb-3 text-sm font-semibold text-text-primary">Data Diri</p>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
            <Field label="Tanggal Lahir" value={formatDate(peserta.birth_date)} />
            <Field label="Nomor HP" value={peserta.phone_number} />
            <Field label="NIK" value={peserta.ktp_number} />
            <Field label="Nomor KTP" value={peserta.nomor_ktp} />
            <Field label="Kota Asal" value={originCity} />
            <Field label="Bandara Terdekat" value={peserta.nearest_airport?.name} />
            <Field label="Pantangan Makanan" value={dietary} />
          </div>
        </div>

        <div>
          <p className="mb-3 text-sm font-semibold text-text-primary">Paspor & Ukuran Blazer</p>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
            <Field label="Nomor Paspor" value={peserta.passport_number} />
            <Field label="Masa Berlaku Paspor" value={formatDate(peserta.passport_expiry)} />
            <Field label="Blazer Size" value={peserta.blazer_size} />
            <Field label="Nomor Meja" value={peserta.nomor_meja} />
          </div>
        </div>

        <div>
          <p className="mb-3 text-sm font-semibold text-text-primary">Catatan Admin</p>
          <Field label="Deskripsi" value={peserta.description} />
        </div>

        <div className="flex flex-wrap gap-6">
          {ktpUrl && (
            <div>
              <p className="mb-2 text-sm font-semibold text-text-primary">KTP</p>
              <button
                type="button"
                onClick={() => setZoomedUrl(ktpUrl)}
                className="group relative inline-block overflow-hidden rounded-md border border-surface-border"
              >
                <img src={ktpUrl} alt="KTP" className="h-40 w-auto object-cover" />
                <span className="absolute inset-0 flex items-center justify-center bg-black/0 text-white opacity-0 transition group-hover:bg-black/30 group-hover:opacity-100">
                  <ZoomIn size={22} />
                </span>
              </button>
            </div>
          )}
          {passportUrl && (
            <div>
              <p className="mb-2 text-sm font-semibold text-text-primary">Paspor</p>
              <button
                type="button"
                onClick={() => setZoomedUrl(passportUrl)}
                className="group relative inline-block overflow-hidden rounded-md border border-surface-border"
              >
                <img src={passportUrl} alt="Paspor" className="h-40 w-auto object-cover" />
                <span className="absolute inset-0 flex items-center justify-center bg-black/0 text-white opacity-0 transition group-hover:bg-black/30 group-hover:opacity-100">
                  <ZoomIn size={22} />
                </span>
              </button>
            </div>
          )}
        </div>
      </div>

      {zoomedUrl && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-6"
          onClick={() => setZoomedUrl(null)}
        >
          <button
            type="button"
            onClick={() => setZoomedUrl(null)}
            aria-label="Close"
            className="absolute right-6 top-6 text-white/80 hover:text-white"
          >
            <X size={24} />
          </button>
          <img src={zoomedUrl} alt="Zoomed" className="max-h-full max-w-full rounded-md object-contain" />
        </div>
      )}
    </Modal>
  )
}

export default PesertaDetailModal
