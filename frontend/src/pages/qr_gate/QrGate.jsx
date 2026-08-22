import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Plus, Pencil, Trash2, QrCode } from 'lucide-react'
import Table from '../../components/ui/Table'
import Badge from '../../components/ui/Badge'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import { Menubar, MenubarAction, MenubarLabel, MenubarSeparator } from '../../components/ui/Menubar'
import QrGateFormModal from './QrGateFormModal'
import QrGateCodeModal from './QrGateCodeModal'
import { getQrGates, createQrGate, updateQrGate, deleteQrGate } from '../../api/qrGate'
import usePermission from '../../hooks/usePermission'
import usePageActions from '../../hooks/usePageActions'

function QrGate() {
  const { canCreate, canEdit, canDelete } = usePermission()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState('name')
  const [sortDir, setSortDir] = useState('asc')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [deleting, setDeleting] = useState(null)
  const [viewingCode, setViewingCode] = useState(null)

  const { data, isLoading } = useQuery({
    queryKey: ['qr-gate', { page, limit, search, sortBy, sortDir }],
    queryFn: () => getQrGates({ page, limit, search, sort_by: sortBy, sort_dir: sortDir }),
  })

  const saveMutation = useMutation({
    mutationFn: (payload) => (editing ? updateQrGate(editing.uuid, payload) : createQrGate(payload)),
    onSuccess: () => {
      toast.success(editing ? 'QR gate updated' : 'QR gate created')
      queryClient.invalidateQueries({ queryKey: ['qr-gate'] })
      setFormOpen(false)
      setEditing(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Save failed'),
  })

  const deleteMutation = useMutation({
    mutationFn: (uuid) => deleteQrGate(uuid),
    onSuccess: () => {
      toast.success('QR gate deleted')
      queryClient.invalidateQueries({ queryKey: ['qr-gate'] })
      setDeleting(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Delete failed'),
  })

  const columns = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'code', label: 'Code', sortable: true, render: (row) => <span className="font-mono">{row.code}</span> },
    { key: 'points', label: 'Points', sortable: true },
    {
      key: 'is_reusable',
      label: 'Scan Rule',
      render: (row) => (
        <Badge variant={row.is_reusable ? 'primary' : 'warning'}>{row.is_reusable ? 'Reusable' : 'One-time'}</Badge>
      ),
    },
    { key: 'scan_count', label: 'Scanned By', render: (row) => `${row.scan_count} user` },
    {
      key: 'actions',
      label: '',
      render: (row) => (
        <div className="flex justify-end gap-1">
          <button onClick={() => setViewingCode(row)} className="rounded p-1.5 hover:bg-surface-hover" title="View QR">
            <QrCode size={16} />
          </button>
          {canEdit('qr-gate') && (
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
          {canDelete('qr-gate') && (
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
          <MenubarLabel>QR Gate</MenubarLabel>
          {canCreate('qr-gate') && (
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
        </Menubar>
      ),
      [canCreate],
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

      <QrGateFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={(values) => saveMutation.mutate(values)}
        initialData={editing}
        loading={saveMutation.isPending}
      />

      <QrGateCodeModal gate={viewingCode} onClose={() => setViewingCode(null)} />

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleteMutation.mutate(deleting.uuid)}
        loading={deleteMutation.isPending}
        message={`Delete QR gate "${deleting?.name}"? This cannot be undone.`}
      />
    </div>
  )
}

export default QrGate
