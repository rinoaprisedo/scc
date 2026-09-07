import { Loader2, CheckCircle2, XCircle } from 'lucide-react'
import Modal from '../../components/ui/Modal'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

// Drives the whole "verify before import" flow through a single modal
// rather than juggling separate loading/preview/result modals — the phase
// prop tells it which of the four states to render. All-or-nothing by
// design: onConfirm is only reachable once preview.errors is empty, and the
// caller (Peserta.jsx) never calls the real import endpoint otherwise.
function PesertaImportModal({ open, onClose, phase, preview, importResult, onConfirm }) {
  if (!open) return null

  const errors = preview?.errors || []
  const hasErrors = errors.length > 0

  return (
    <Modal open={open} onClose={onClose} title="Import Peserta" size="lg">
      {phase === 'validating' && (
        <div className="flex flex-col items-center justify-center gap-3 py-10 text-text-secondary">
          <Loader2 size={28} className="animate-spin text-primary" />
          <p className="text-sm">Memeriksa file, mohon tunggu...</p>
        </div>
      )}

      {phase === 'importing' && (
        <div className="flex flex-col items-center justify-center gap-3 py-10 text-text-secondary">
          <Loader2 size={28} className="animate-spin text-primary" />
          <p className="text-sm">Mengimpor data ke database, mohon tunggu...</p>
        </div>
      )}

      {phase === 'done' && (
        <div className="flex flex-col items-center justify-center gap-3 py-10 text-center">
          <CheckCircle2 size={36} className="text-success" />
          <p className="text-sm font-medium text-text-primary">
            {importResult?.created ?? 0} peserta baru ditambahkan, {importResult?.updated ?? 0} peserta diperbarui
          </p>
          <Button className="mt-2" onClick={onClose}>
            Selesai
          </Button>
        </div>
      )}

      {phase === 'preview' && preview && (
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            <Badge variant="neutral">{preview.total_rows} baris terbaca</Badge>
            <Badge variant="success">{preview.valid} valid</Badge>
            {preview.to_create > 0 && <Badge variant="neutral">{preview.to_create} peserta baru</Badge>}
            {preview.to_update > 0 && <Badge variant="neutral">{preview.to_update} akan diperbarui</Badge>}
            <Badge variant={hasErrors ? 'danger' : 'neutral'}>{errors.length} tidak valid</Badge>
          </div>

          {hasErrors ? (
            <>
              <div className="flex items-start gap-2 rounded-md border border-danger/30 bg-danger/5 px-3 py-2 text-sm text-danger">
                <XCircle size={16} className="mt-0.5 shrink-0" />
                <span>
                  Ada baris yang datanya belum sesuai. Perbaiki file lalu upload ulang — selama masih ada error, tidak
                  ada data yang akan dimasukkan ke database.
                </span>
              </div>
              <div className="max-h-72 overflow-y-auto rounded-md border border-surface-border">
                <table className="w-full text-sm">
                  <thead className="bg-surface-hover text-left text-xs uppercase text-text-secondary">
                    <tr>
                      <th className="px-3 py-2">Baris</th>
                      <th className="px-3 py-2">Error</th>
                    </tr>
                  </thead>
                  <tbody>
                    {errors.map((e, i) => (
                      <tr key={i} className="border-t border-surface-border">
                        <td className="px-3 py-2 font-medium">{e.row}</td>
                        <td className="px-3 py-2 text-text-secondary">{e.message}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          ) : (
            <div className="flex items-start gap-2 rounded-md border border-success/30 bg-success/5 px-3 py-2 text-sm text-success">
              <CheckCircle2 size={16} className="mt-0.5 shrink-0" />
              <span>Semua data valid dan siap diimport.</span>
            </div>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={onClose}>
              Batal
            </Button>
            <Button type="button" onClick={onConfirm} disabled={hasErrors}>
              Import {preview.valid} Peserta
            </Button>
          </div>
        </div>
      )}
    </Modal>
  )
}

export default PesertaImportModal
