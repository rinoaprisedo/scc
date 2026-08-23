import { useEffect, useState } from 'react'
import { X, Trash2, ZoomIn, CreditCard, ChevronLeft, ChevronRight } from 'lucide-react'
import Button from './ui/Button'
import FileUpload from './ui/FileUpload'
import { getMyQrisCrossBorder, createMyQrisCrossBorder, deleteMyQrisCrossBorder } from '../api/qrisCrossBorder'
import { fileURL } from '../utils/url'

const STATUS_STYLE = {
  pending: 'bg-amber-100 text-amber-700',
  waiting_approval: 'bg-amber-100 text-amber-700',
  approved: 'bg-green-100 text-green-700',
  rejected: 'bg-red-100 text-red-700',
}
const STATUS_LABEL = {
  pending: 'Diproses',
  waiting_approval: 'Menunggu Persetujuan',
  approved: 'Disetujui',
  rejected: 'Ditolak',
}

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

const PAGE_LIMIT = 5

function QrisCrossBorderModal({ onClose }) {
  const [list, setList] = useState([])
  const [loadingList, setLoadingList] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [file, setFile] = useState(null)
  const [preview, setPreview] = useState(null)
  const [fileError, setFileError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState('')
  const [confirmDelete, setConfirmDelete] = useState(null)
  const [deletingUuid, setDeletingUuid] = useState(null)
  const [zoomedImage, setZoomedImage] = useState(null)

  const loadList = (targetPage) => {
    setLoadingList(true)
    getMyQrisCrossBorder({ page: targetPage, limit: PAGE_LIMIT })
      .then((res) => {
        setList(res.data || [])
        setPage(res.meta?.page || targetPage)
        setTotalPages(res.meta?.total_pages || 1)
      })
      .finally(() => setLoadingList(false))
  }

  useEffect(() => {
    loadList(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSubmit = () => {
    if (!file) {
      setFileError('Bukti pembayaran wajib diunggah')
      return
    }
    setSubmitting(true)
    setSubmitError('')
    const formData = new FormData()
    formData.append('image', file)
    createMyQrisCrossBorder(formData)
      .then(() => {
        setFile(null)
        setPreview(null)
        loadList(1)
      })
      .catch((err) => setSubmitError(err.response?.data?.message || 'Gagal mengirim bukti pembayaran'))
      .finally(() => setSubmitting(false))
  }

  const handleDelete = (uuid) => {
    setDeletingUuid(uuid)
    deleteMyQrisCrossBorder(uuid)
      .then(() => {
        setConfirmDelete(null)
        loadList(list.length === 1 && page > 1 ? page - 1 : page)
      })
      .finally(() => setDeletingUuid(null))
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 py-8 backdrop-blur-sm">
      <div className="flex max-h-[90vh] w-full max-w-lg flex-col overflow-hidden rounded-2xl bg-white shadow-2xl">
        <div className="relative shrink-0 bg-gradient-to-br from-navy via-navy to-navy-light px-8 pb-6 pt-8 text-center">
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="absolute right-5 top-5 text-white/70 transition hover:text-white"
          >
            <X size={20} />
          </button>
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-gold-light via-gold to-gold-dark text-navy-dark shadow-md ring-4 ring-white/20">
            <CreditCard size={24} />
          </div>
          <p className="mt-3 text-xl font-bold text-white">QRIS Cross Border</p>
          <p className="text-xs font-medium uppercase tracking-widest text-gold-light">Upload bukti transaksi</p>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-6 py-6">
          <div className="flex flex-col gap-3 rounded-xl border border-surface-border p-4">
            <FileUpload
              label="Upload Screenshot QRIS"
              required
              hint="JPEG, PNG, atau WebP"
              error={fileError}
              preview={preview}
              onFileSelect={(f) => {
                setFile(f)
                setPreview(URL.createObjectURL(f))
                setFileError('')
                setSubmitError('')
              }}
              onClear={() => {
                setFile(null)
                setPreview(null)
              }}
            />
            {submitError && <p className="text-xs font-medium text-red-600">{submitError}</p>}
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? 'Mengirim...' : 'Kirim'}
            </Button>
          </div>

          <p className="mb-2 mt-6 text-xs font-semibold uppercase tracking-widest text-text-secondary">Riwayat</p>

          {loadingList ? (
            <p className="py-10 text-center text-sm text-text-secondary">Memuat...</p>
          ) : list.length === 0 ? (
            <p className="py-10 text-center text-sm text-text-secondary">Belum ada bukti pembayaran yang diunggah</p>
          ) : (
            <div className="flex flex-col gap-3">
              {list.map((row) => (
                <div key={row.uuid} className="flex items-center gap-3 rounded-lg border border-surface-border px-3 py-3">
                  <button
                    type="button"
                    onClick={() => setZoomedImage(fileURL(row.image))}
                    className="group relative h-14 w-14 shrink-0 overflow-hidden rounded-lg border border-surface-border"
                  >
                    <img src={fileURL(row.image)} alt="Bukti pembayaran" className="h-full w-full object-cover" />
                    <span className="absolute inset-0 flex items-center justify-center bg-black/0 text-white opacity-0 transition group-hover:bg-black/30 group-hover:opacity-100">
                      <ZoomIn size={16} />
                    </span>
                  </button>

                  <div className="min-w-0 flex-1">
                    {(row.nominal_rupiah > 0 || row.status !== 'rejected') && (
                      <p className="text-sm font-semibold text-navy">
                        {row.nominal_rupiah > 0 ? `Rp ${Number(row.nominal_rupiah).toLocaleString('id-ID')}` : 'Diproses...'}
                      </p>
                    )}
                    <p className="text-xs text-text-secondary">{formatDate(row.created_at)}</p>
                    {row.status === 'rejected' && row.reject_reason && (
                      <p className="mt-0.5 text-xs text-red-600">{row.reject_reason}</p>
                    )}
                  </div>

                  <span className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold ${STATUS_STYLE[row.status] || 'bg-surface-bg text-text-secondary'}`}>
                    {STATUS_LABEL[row.status] || row.status}
                  </span>

                  {confirmDelete === row.uuid ? (
                    <div className="flex shrink-0 items-center gap-1.5">
                      <button
                        type="button"
                        onClick={() => handleDelete(row.uuid)}
                        disabled={deletingUuid === row.uuid}
                        className="rounded-md bg-red-600 px-2 py-1 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-60"
                      >
                        {deletingUuid === row.uuid ? '...' : 'Ya'}
                      </button>
                      <button
                        type="button"
                        onClick={() => setConfirmDelete(null)}
                        className="rounded-md bg-surface-bg px-2 py-1 text-xs font-semibold text-text-secondary hover:bg-surface-border"
                      >
                        Batal
                      </button>
                    </div>
                  ) : (
                    <button
                      type="button"
                      onClick={() => setConfirmDelete(row.uuid)}
                      aria-label="Hapus"
                      className="shrink-0 rounded-full p-1.5 text-text-secondary transition hover:bg-red-50 hover:text-red-600"
                    >
                      <Trash2 size={16} />
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}

          {!loadingList && totalPages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <button
                type="button"
                onClick={() => loadList(page - 1)}
                disabled={page <= 1}
                className="flex items-center gap-1 rounded-md px-2 py-1 text-xs font-semibold text-navy transition hover:bg-surface-bg disabled:cursor-not-allowed disabled:opacity-40"
              >
                <ChevronLeft size={14} />
                Sebelumnya
              </button>
              <span className="text-xs text-text-secondary">
                Halaman {page} dari {totalPages}
              </span>
              <button
                type="button"
                onClick={() => loadList(page + 1)}
                disabled={page >= totalPages}
                className="flex items-center gap-1 rounded-md px-2 py-1 text-xs font-semibold text-navy transition hover:bg-surface-bg disabled:cursor-not-allowed disabled:opacity-40"
              >
                Berikutnya
                <ChevronRight size={14} />
              </button>
            </div>
          )}
        </div>
      </div>

      {zoomedImage && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-6"
          onClick={() => setZoomedImage(null)}
        >
          <button
            type="button"
            onClick={() => setZoomedImage(null)}
            aria-label="Close"
            className="absolute right-6 top-6 text-white/80 hover:text-white"
          >
            <X size={24} />
          </button>
          <img src={zoomedImage} alt="Bukti pembayaran (perbesar)" className="max-h-full max-w-full rounded-md object-contain" />
        </div>
      )}
    </div>
  )
}

export default QrisCrossBorderModal
