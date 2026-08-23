import { useState, useMemo } from 'react'
import { createPortal } from 'react-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Plus, Pencil, Trash2, Check, X, Filter, ZoomIn } from 'lucide-react'
import { format } from 'date-fns'
import Table from '../../components/ui/Table'
import Badge from '../../components/ui/Badge'
import Button from '../../components/ui/Button'
import Select from '../../components/ui/Select'
import AsyncSelect from '../../components/ui/AsyncSelect'
import { Menubar, MenubarAction, MenubarMenu, MenubarLabel, MenubarSeparator } from '../../components/ui/Menubar'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import QrisCrossBorderFormModal from './QrisCrossBorderFormModal'
import {
  getQrisCrossBorder,
  createQrisCrossBorder,
  updateQrisCrossBorder,
  deleteQrisCrossBorder,
  updateQrisCrossBorderStatus,
} from '../../api/qrisCrossBorder'
import { searchPesertaOptions } from '../../api/peserta'
import { fileURL } from '../../utils/url'
import usePermission from '../../hooks/usePermission'
import usePageActions from '../../hooks/usePageActions'

const statusVariant = { pending: 'neutral', waiting_approval: 'warning', approved: 'success', rejected: 'danger' }
const statusLabel = { pending: 'Pending (Diproses AI)', waiting_approval: 'Waiting Approval', approved: 'Approved', rejected: 'Rejected' }

function QrisCrossBorder() {
  const { canCreate, canEdit, canDelete } = usePermission()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState('created_at')
  const [sortDir, setSortDir] = useState('desc')
  const [status, setStatus] = useState('')
  const [draftStatus, setDraftStatus] = useState('')
  const [pesertaFilter, setPesertaFilter] = useState('')
  const [draftPeserta, setDraftPeserta] = useState('')
  const [draftPesertaLabel, setDraftPesertaLabel] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [deleting, setDeleting] = useState(null)
  const [zoomedImage, setZoomedImage] = useState(null)

  const { data, isLoading } = useQuery({
    queryKey: ['qris-cross-border', { page, limit, search, sortBy, sortDir, status, pesertaFilter }],
    queryFn: () =>
      getQrisCrossBorder({
        page,
        limit,
        search,
        sort_by: sortBy,
        sort_dir: sortDir,
        status,
        peserta_uuid: pesertaFilter,
      }),
  })

  const saveMutation = useMutation({
    mutationFn: (formData) => (editing ? updateQrisCrossBorder(editing.uuid, formData) : createQrisCrossBorder(formData)),
    onSuccess: () => {
      toast.success(editing ? 'Record updated' : 'Record created')
      queryClient.invalidateQueries({ queryKey: ['qris-cross-border'] })
      setFormOpen(false)
      setEditing(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Save failed'),
  })

  const statusMutation = useMutation({
    mutationFn: ({ uuid, newStatus }) => updateQrisCrossBorderStatus(uuid, newStatus),
    onSuccess: () => {
      toast.success('Status updated')
      queryClient.invalidateQueries({ queryKey: ['qris-cross-border'] })
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Update failed'),
  })

  const deleteMutation = useMutation({
    mutationFn: (uuid) => deleteQrisCrossBorder(uuid),
    onSuccess: () => {
      toast.success('Record deleted')
      queryClient.invalidateQueries({ queryKey: ['qris-cross-border'] })
      setDeleting(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Delete failed'),
  })

  const columns = [
    {
      key: 'image',
      label: 'Proof',
      render: (row) => (
        <button
          type="button"
          onClick={() => setZoomedImage(fileURL(row.image))}
          className="group relative block h-10 w-10 overflow-hidden rounded-md border border-surface-border"
        >
          <img src={fileURL(row.image)} alt="proof" className="h-full w-full object-cover" />
          <span className="absolute inset-0 flex items-center justify-center bg-black/0 text-white opacity-0 transition group-hover:bg-black/30 group-hover:opacity-100">
            <ZoomIn size={16} />
          </span>
        </button>
      ),
    },
    { key: 'peserta_name', label: 'Peserta', sortable: false, render: (row) => row.peserta_name || '-' },
    { key: 'peserta_ktp_number', label: 'NIK', render: (row) => row.peserta_ktp_number || '-' },
    { key: 'merchant_name', label: 'Merchant', sortable: false, render: (row) => row.merchant_name || '-' },
    { key: 'reference_number', label: 'No. Referensi', sortable: false, render: (row) => row.reference_number || '-' },
    {
      key: 'nominal_asing',
      label: 'Nominal (Asing)',
      sortable: true,
      render: (row) => Number(row.nominal_asing).toLocaleString('id-ID', { minimumFractionDigits: 2 }),
    },
    {
      key: 'nominal_rupiah',
      label: 'Nominal (IDR)',
      sortable: true,
      render: (row) => `Rp ${Number(row.nominal_rupiah).toLocaleString('id-ID')}`,
    },
    {
      key: 'status',
      label: 'Status',
      sortable: true,
      render: (row) => (
        <div className="flex flex-col gap-0.5">
          <Badge variant={statusVariant[row.status] || 'neutral'}>{statusLabel[row.status] || row.status}</Badge>
          {row.status === 'rejected' && row.reject_reason && (
            <span className="text-xs text-text-tertiary">{row.reject_reason}</span>
          )}
        </div>
      ),
    },
    {
      key: 'created_at',
      label: 'Created',
      sortable: true,
      render: (row) => format(new Date(row.created_at), 'PP p'),
    },
    {
      key: 'actions',
      label: '',
      render: (row) => (
        <div className="flex justify-end gap-1">
          {canEdit('qris-cross-border') && row.status === 'waiting_approval' && (
            <button
              onClick={() => statusMutation.mutate({ uuid: row.uuid, newStatus: 'approved' })}
              className="rounded p-1.5 text-success hover:bg-success-bg"
              title="Approve"
            >
              <Check size={16} />
            </button>
          )}
          {canEdit('qris-cross-border') && row.status === 'waiting_approval' && (
            <button
              onClick={() => statusMutation.mutate({ uuid: row.uuid, newStatus: 'rejected' })}
              className="rounded p-1.5 text-danger hover:bg-danger-bg"
              title="Reject"
            >
              <X size={16} />
            </button>
          )}
          {canEdit('qris-cross-border') && (
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
          {canDelete('qris-cross-border') && (
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
          <MenubarLabel>Qris Cross Border</MenubarLabel>
          {canCreate('qris-cross-border') && (
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
          <MenubarMenu id="filter" label="Filter" icon={Filter} align="sheet" active={!!(status || pesertaFilter)}>
            {(close) => (
              <div className="flex w-full flex-col gap-3 p-2 sm:w-64">
                <AsyncSelect
                  label="Peserta"
                  placeholder="All peserta"
                  value={draftPeserta}
                  initialLabel={draftPesertaLabel}
                  loadOptions={searchPesertaOptions}
                  getOptionValue={(o) => o.uuid}
                  getOptionLabel={(o) => o.name}
                  getOptionSublabel={(o) => o.ktp_number}
                  onChange={(uuid, opt) => {
                    setDraftPeserta(uuid)
                    setDraftPesertaLabel(opt.name)
                  }}
                />
                <Select label="Status" value={draftStatus} onChange={(e) => setDraftStatus(e.target.value)}>
                  <option value="">All</option>
                  <option value="pending">Pending</option>
                  <option value="waiting_approval">Waiting Approval</option>
                  <option value="approved">Approved</option>
                  <option value="rejected">Rejected</option>
                </Select>
                <div className="flex items-center justify-end gap-2 border-t border-surface-border pt-3">
                  <Button
                    variant="secondary"
                    className="h-[32px] px-3"
                    onClick={() => {
                      setDraftStatus('')
                      setDraftPeserta('')
                      setDraftPesertaLabel('')
                      setStatus('')
                      setPesertaFilter('')
                      setPage(1)
                      close()
                    }}
                  >
                    Reset
                  </Button>
                  <Button
                    className="h-[32px] px-3"
                    onClick={() => {
                      setStatus(draftStatus)
                      setPesertaFilter(draftPeserta)
                      setPage(1)
                      close()
                    }}
                  >
                    Apply
                  </Button>
                </div>
              </div>
            )}
          </MenubarMenu>
        </Menubar>
      ),
      [canCreate, status, draftStatus, pesertaFilter, draftPeserta, draftPesertaLabel],
    ),
  )

  return (
    <div className="space-y-4">
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

      <QrisCrossBorderFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={(formData) => saveMutation.mutate(formData)}
        initialData={editing}
        loading={saveMutation.isPending}
      />

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleteMutation.mutate(deleting.uuid)}
        loading={deleteMutation.isPending}
        message={`Delete this record for "${deleting?.peserta_name}"? This cannot be undone.`}
      />

      {zoomedImage &&
        createPortal(
          // Portalled to document.body — this page's root div uses
          // space-y-4, whose sibling selector injects margin-top on every
          // non-first child. For a position:fixed box that margin still
          // applies to its border box (top:0 only pins the margin edge),
          // pushing the whole overlay down and leaving a gap of bare page
          // above it. Same rationale as Modal.jsx's portal.
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
            <img src={zoomedImage} alt="Proof of payment (zoomed)" className="max-h-full max-w-full rounded-md object-contain" />
          </div>,
          document.body,
        )}
    </div>
  )
}

export default QrisCrossBorder
