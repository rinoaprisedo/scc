import { useState, useMemo, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Plus, Pencil, Trash2, Eye, Trophy, MoreHorizontal, FileSpreadsheet, FileText, FileUp, FileDown, FileArchive } from 'lucide-react'
import { format } from 'date-fns'
import Table from '../../components/ui/Table'
import Badge from '../../components/ui/Badge'
import { Menubar, MenubarAction, MenubarMenu, MenubarItem, MenubarLabel, MenubarSeparator } from '../../components/ui/Menubar'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import PesertaFormModal from './PesertaFormModal'
import PesertaDetailModal from './PesertaDetailModal'
import PesertaPointHistoryModal from './PesertaPointHistoryModal'
import PesertaImportModal from './PesertaImportModal'
import {
  getPeserta,
  createPeserta,
  updatePeserta,
  deletePeserta,
  uploadPesertaKtp,
  exportPesertaCsv,
  exportPesertaExcel,
  exportPesertaKtpZip,
  downloadPesertaImportTemplate,
  validatePesertaImport,
  importPesertaExcel,
} from '../../api/peserta'
import { getKotaAsal } from '../../api/kotaAsal'
import { getBandara } from '../../api/bandara'
import usePermission from '../../hooks/usePermission'
import usePageActions from '../../hooks/usePageActions'
import { attendanceLabel } from '../../utils/attendance'

function Peserta() {
  const { canCreate, canEdit, canDelete } = usePermission()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState('created_at')
  const [sortDir, setSortDir] = useState('desc')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [deleting, setDeleting] = useState(null)
  const [viewing, setViewing] = useState(null)
  const [viewingPoints, setViewingPoints] = useState(null)
  const [importOpen, setImportOpen] = useState(false)
  const [importPhase, setImportPhase] = useState('validating') // 'validating' | 'preview' | 'importing' | 'done'
  const [importPreview, setImportPreview] = useState(null)
  const [importCreated, setImportCreated] = useState(0)
  const [importFile, setImportFile] = useState(null)
  const fileInputRef = useRef(null)

  const { data, isLoading } = useQuery({
    queryKey: ['peserta', { page, limit, search, sortBy, sortDir }],
    queryFn: () => getPeserta({ page, limit, search, sort_by: sortBy, sort_dir: sortDir }),
  })

  const { data: kotaAsalData } = useQuery({ queryKey: ['kota-asal', 'all'], queryFn: () => getKotaAsal({ limit: 100 }) })
  const { data: bandaraData } = useQuery({ queryKey: ['bandara', 'all'], queryFn: () => getBandara({ limit: 100 }) })

  const saveMutation = useMutation({
    mutationFn: async ({ payload, ktpFile }) => {
      const saved = editing ? await updatePeserta(editing.uuid, payload) : await createPeserta(payload)
      const uuid = editing ? editing.uuid : saved?.data?.uuid
      if (ktpFile && uuid) {
        const formData = new FormData()
        formData.append('ktp_file', ktpFile)
        await uploadPesertaKtp(uuid, formData)
      }
      return saved
    },
    onSuccess: () => {
      toast.success(editing ? 'Peserta updated' : 'Peserta created')
      queryClient.invalidateQueries({ queryKey: ['peserta'] })
      setFormOpen(false)
      setEditing(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Save failed'),
  })

  // Import is a two-step, all-or-nothing flow: the file is first sent to a
  // dry-run validate endpoint that writes nothing, showing exactly which
  // rows (if any) are wrong; only once that comes back clean does confirming
  // hit the real import endpoint. The backend re-validates from scratch at
  // that point too (DB state can shift between preview and confirm), so
  // importMutation's onError still has to handle a fresh batch of row
  // errors, not just a hard failure.
  const validateMutation = useMutation({
    mutationFn: (file) => {
      const formData = new FormData()
      formData.append('file', file)
      return validatePesertaImport(formData)
    },
    onSuccess: (res) => {
      setImportPreview(res.data)
      setImportPhase('preview')
    },
    onError: (err) => {
      toast.error(err.response?.data?.message || 'Failed to validate file')
      setImportOpen(false)
    },
  })

  const importMutation = useMutation({
    mutationFn: (file) => {
      const formData = new FormData()
      formData.append('file', file)
      return importPesertaExcel(formData)
    },
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['peserta'] })
      setImportCreated(res.data.created)
      setImportPhase('done')
    },
    onError: (err) => {
      const errors = err.response?.data?.data?.errors
      if (errors) {
        setImportPreview((prev) => ({ ...prev, valid: 0, errors }))
        setImportPhase('preview')
        toast.error('Some data changed since validation — please review again')
      } else {
        toast.error(err.response?.data?.message || 'Import failed')
        setImportOpen(false)
      }
    },
  })

  const handleImportFile = (e) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    setImportFile(file)
    setImportPreview(null)
    setImportPhase('validating')
    setImportOpen(true)
    validateMutation.mutate(file)
  }

  const confirmImport = () => {
    setImportPhase('importing')
    importMutation.mutate(importFile)
  }

  const deleteMutation = useMutation({
    mutationFn: (uuid) => deletePeserta(uuid),
    onSuccess: () => {
      toast.success('Peserta deleted')
      queryClient.invalidateQueries({ queryKey: ['peserta'] })
      setDeleting(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Delete failed'),
  })

  const columns = [
    {
      key: 'name',
      label: 'Name',
      sortable: true,
      render: (row) => (
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
            {row.avatar ? (
              <img src={row.avatar} alt={row.name} className="h-8 w-8 rounded-full object-cover" />
            ) : (
              row.name?.charAt(0).toUpperCase()
            )}
          </div>
          <span className="font-medium">{row.name}</span>
        </div>
      ),
    },
    { key: 'ktp_number', label: 'NIK', sortable: true },
    { key: 'phone_number', label: 'Phone', render: (row) => row.phone_number || '-' },
    {
      key: 'attendance_status',
      label: 'Kehadiran',
      render: (row) => {
        const attendance = attendanceLabel(row.attendance_status)
        return <Badge variant={attendance.variant}>{attendance.text}</Badge>
      },
    },
    { key: 'total_points', label: 'Total Points', render: (row) => row.total_points ?? 0 },
    {
      key: 'last_login_at',
      label: 'Last Login',
      sortable: true,
      render: (row) => (row.last_login_at ? format(new Date(row.last_login_at), 'PP p') : 'Never'),
    },
    {
      key: 'actions',
      label: '',
      render: (row) => (
        <div className="flex justify-end gap-1">
          <button onClick={() => setViewing(row)} className="rounded p-1.5 hover:bg-surface-hover" title="Detail">
            <Eye size={16} />
          </button>
          <button
            onClick={() => setViewingPoints(row)}
            className="rounded p-1.5 hover:bg-surface-hover"
            title="Point history"
          >
            <Trophy size={16} />
          </button>
          {canEdit('peserta') && (
            <button
              onClick={() => {
                setEditing(row)
                setFormOpen(true)
              }}
              className="rounded p-1.5 hover:bg-surface-hover"
              title="Edit"
            >
              <Pencil size={16} />
            </button>
          )}
          {canDelete('peserta') && (
            <button onClick={() => setDeleting(row)} className="rounded p-1.5 text-danger hover:bg-danger/10" title="Delete">
              <Trash2 size={16} />
            </button>
          )}
        </div>
      ),
    },
  ]

  usePageActions(
    useMemo(
      () => (
        <Menubar>
          <MenubarLabel>Peserta</MenubarLabel>
          {canCreate('peserta') && (
            <>
              <MenubarSeparator />
              <MenubarAction
                label="Add New"
                icon={Plus}
                onClick={() => {
                  setEditing(null)
                  setFormOpen(true)
                }}
              />
            </>
          )}
          <MenubarMenu id="others" label="Others" icon={MoreHorizontal} align="right">
            <MenubarItem
              label="Export Excel"
              icon={FileSpreadsheet}
              onClick={() =>
                toast.promise(exportPesertaExcel({ search }), {
                  loading: 'Exporting...',
                  success: 'Excel file downloaded',
                  error: 'Export failed',
                })
              }
            />
            <MenubarItem
              label="Export CSV"
              icon={FileText}
              onClick={() =>
                toast.promise(exportPesertaCsv({ search }), {
                  loading: 'Exporting...',
                  success: 'CSV file downloaded',
                  error: 'Export failed',
                })
              }
            />
            <MenubarItem
              label="Export KTP (ZIP)"
              icon={FileArchive}
              onClick={() =>
                toast.promise(exportPesertaKtpZip({ search }), {
                  loading: 'Exporting...',
                  success: 'KTP ZIP downloaded',
                  error: 'Export failed',
                })
              }
            />
            {canCreate('peserta') && (
              <>
                <MenubarSeparator />
                <MenubarItem
                  label="Download Import Template"
                  icon={FileDown}
                  onClick={() =>
                    toast.promise(downloadPesertaImportTemplate(), {
                      loading: 'Downloading...',
                      success: 'Template downloaded',
                      error: 'Download failed',
                    })
                  }
                />
                <MenubarItem label="Import Excel" icon={FileUp} onClick={() => fileInputRef.current?.click()} />
              </>
            )}
          </MenubarMenu>
        </Menubar>
      ),
      [canCreate, search],
    ),
  )

  return (
    <div className="space-y-4">
      <input
        ref={fileInputRef}
        type="file"
        accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
        className="hidden"
        onChange={handleImportFile}
      />

      <Table
        columns={columns}
        data={data?.data || []}
        loading={isLoading}
        search={search}
        onSearchChange={(v) => {
          setSearch(v)
          setPage(1)
        }}
        sortBy={sortBy}
        sortDir={sortDir}
        onSortChange={(k, d) => {
          setSortBy(k)
          setSortDir(d)
        }}
        page={page}
        limit={limit}
        total={data?.meta?.total || 0}
        onPageChange={setPage}
        onLimitChange={(l) => {
          setLimit(l)
          setPage(1)
        }}
      />

      <PesertaDetailModal open={!!viewing} onClose={() => setViewing(null)} peserta={viewing} />

      <PesertaPointHistoryModal peserta={viewingPoints} onClose={() => setViewingPoints(null)} />

      <PesertaImportModal
        open={importOpen}
        onClose={() => setImportOpen(false)}
        phase={importPhase}
        preview={importPreview}
        createdCount={importCreated}
        onConfirm={confirmImport}
      />

      <PesertaFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={(payload, ktpFile) => saveMutation.mutate({ payload, ktpFile })}
        initialData={editing}
        kotaAsal={kotaAsalData?.data || []}
        bandara={bandaraData?.data || []}
        loading={saveMutation.isPending}
      />

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleteMutation.mutate(deleting.uuid)}
        loading={deleteMutation.isPending}
        message={`Delete peserta "${deleting?.name}"? This cannot be undone.`}
      />
    </div>
  )
}

export default Peserta
