import { Loader2, CheckCircle2, XCircle } from 'lucide-react'
import Modal from '../../components/ui/Modal'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'

// IMPORT_FIELDS mirrors importColumnDefs in backend/modules/peserta/import.go
// (minus NIK, which is the match key and always read) — kept in sync by hand.
export const IMPORT_FIELDS = [
  { key: 'name', label: 'Name' },
  { key: 'nomor_ktp', label: 'Nomor KTP' },
  { key: 'email', label: 'Email' },
  { key: 'title', label: 'Title' },
  { key: 'first_name', label: 'Nama Depan' },
  { key: 'middle_name', label: 'Nama Tengah' },
  { key: 'last_name', label: 'Nama Belakang' },
  { key: 'birth_date', label: 'Tanggal Lahir' },
  { key: 'origin_city', label: 'Kota Asal' },
  { key: 'nearest_airport', label: 'Bandara Terdekat' },
  { key: 'dietary_restriction', label: 'Pantangan Makanan' },
  { key: 'phone_number', label: 'Nomor HP' },
  { key: 'region', label: 'Region' },
  { key: 'cabang', label: 'Cabang' },
  { key: 'position', label: 'Position' },
  { key: 'passport_number', label: 'Nomor Passport' },
  { key: 'passport_expiry', label: 'Masa Berlaku Passport' },
  { key: 'blazer_size', label: 'Blazer Size' },
  { key: 'nomor_meja', label: 'Nomor Meja' },
  { key: 'description', label: 'Deskripsi' },
]

const IMPORT_MODES = [
  { value: 'upsert', label: 'Update & Insert', hint: 'NIK yang sudah ada diperbarui, NIK baru ditambahkan' },
  { value: 'insert', label: 'Insert saja', hint: 'Hanya menambah peserta baru, NIK yang sudah ada dilewati' },
  { value: 'update', label: 'Update saja', hint: 'Hanya memperbarui peserta yang NIK-nya sudah ada, NIK baru dilewati' },
]

// Name is forced on whenever the mode can create rows — the backend does the
// same (ParseImportOptions), since a new peserta can't exist without one.
export const isFieldLocked = (mode, key) => key === 'name' && mode !== 'update'

// Drives the whole "verify before import" flow through a single modal
// rather than juggling separate loading/preview/result modals — the phase
// prop tells it which of the five states to render. All-or-nothing by
// design: onConfirm is only reachable once preview.errors is empty, and the
// caller (Peserta.jsx) never calls the real import endpoint otherwise.
function PesertaImportModal({
  open,
  onClose,
  phase,
  preview,
  importResult,
  onConfirm,
  fileName,
  mode,
  onModeChange,
  fields,
  onFieldsChange,
  onValidate,
  onBack,
}) {
  if (!open) return null

  const errors = preview?.errors || []
  const hasErrors = errors.length > 0
  const isChecked = (key) => isFieldLocked(mode, key) || fields.includes(key)
  const checkedCount = IMPORT_FIELDS.filter((f) => isChecked(f.key)).length
  const allChecked = checkedCount === IMPORT_FIELDS.length

  const toggleField = (key) =>
    onFieldsChange(fields.includes(key) ? fields.filter((k) => k !== key) : [...fields, key])
  const toggleAll = () => onFieldsChange(allChecked ? [] : IMPORT_FIELDS.map((f) => f.key))

  return (
    <Modal open={open} onClose={onClose} title="Import Peserta" size="lg">
      {phase === 'options' && (
        <div className="space-y-5">
          {fileName && (
            <p className="text-sm text-text-secondary">
              File: <span className="font-medium text-text-primary">{fileName}</span>
            </p>
          )}

          <div className="space-y-2">
            <p className="text-sm font-medium text-text-primary">Mode Import</p>
            <div className="grid gap-2 sm:grid-cols-3">
              {IMPORT_MODES.map((m) => (
                <label
                  key={m.value}
                  className={`flex cursor-pointer flex-col gap-1 rounded-md border px-3 py-2 text-sm ${
                    mode === m.value ? 'border-primary bg-primary/5' : 'border-surface-border'
                  }`}
                >
                  <span className="flex items-center gap-2 font-medium text-text-primary">
                    <input
                      type="radio"
                      name="import_mode"
                      className="h-4 w-4 accent-primary"
                      checked={mode === m.value}
                      onChange={() => onModeChange(m.value)}
                    />
                    {m.label}
                  </span>
                  <span className="text-xs text-text-secondary">{m.hint}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-text-primary">
                Kolom yang diimport <span className="font-normal text-text-secondary">({checkedCount} dipilih)</span>
              </p>
              <button type="button" className="text-xs font-medium text-primary hover:underline" onClick={toggleAll}>
                {allChecked ? 'Kosongkan' : 'Pilih semua'}
              </button>
            </div>
            <p className="text-xs text-text-secondary">
              NIK selalu dipakai sebagai kunci pencocokan. Kolom yang tidak dicentang diabaikan — data lama tidak
              diubah.
            </p>
            <div className="grid grid-cols-2 gap-x-4 gap-y-2 rounded-md border border-surface-border p-3 sm:grid-cols-3">
              {IMPORT_FIELDS.map((f) => {
                const locked = isFieldLocked(mode, f.key)
                return (
                  <label
                    key={f.key}
                    className={`flex items-center gap-2 text-sm ${locked ? 'cursor-not-allowed text-text-secondary' : 'cursor-pointer text-text-primary'}`}
                    title={locked ? 'Wajib untuk peserta baru' : undefined}
                  >
                    <input
                      type="checkbox"
                      className="h-4 w-4 accent-primary"
                      checked={isChecked(f.key)}
                      disabled={locked}
                      onChange={() => toggleField(f.key)}
                    />
                    {f.label}
                  </label>
                )
              })}
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={onClose}>
              Batal
            </Button>
            <Button type="button" onClick={onValidate} disabled={checkedCount === 0}>
              Periksa File
            </Button>
          </div>
        </div>
      )}

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
            {importResult?.skipped > 0 && `, ${importResult.skipped} dilewati`}
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
            {preview.to_skip > 0 && <Badge variant="neutral">{preview.to_skip} dilewati</Badge>}
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
            <Button type="button" variant="secondary" onClick={onBack}>
              Ubah Opsi
            </Button>
            <Button type="button" variant="secondary" onClick={onClose}>
              Batal
            </Button>
            <Button type="button" onClick={onConfirm} disabled={hasErrors || preview.valid === 0}>
              Import {preview.valid} Peserta
            </Button>
          </div>
        </div>
      )}
    </Modal>
  )
}

export default PesertaImportModal
