import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Plus, Pencil, Trash2, ArrowUp, ArrowDown } from 'lucide-react'
import Table from '../../components/ui/Table'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import { Menubar, MenubarAction, MenubarLabel, MenubarSeparator } from '../../components/ui/Menubar'
import MenuSectionFormModal from './MenuSectionFormModal'
import {
  getMenuSections,
  createMenuSection,
  updateMenuSection,
  deleteMenuSection,
  reorderMenuSections,
} from '../../api/menuSections'
import usePermission from '../../hooks/usePermission'
import usePageActions from '../../hooks/usePageActions'

function MenuSections() {
  const { canCreate, canEdit, canDelete } = usePermission()
  const queryClient = useQueryClient()
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [deleting, setDeleting] = useState(null)

  const { data, isLoading } = useQuery({ queryKey: ['menu-sections'], queryFn: () => getMenuSections({ limit: 100 }) })

  const saveMutation = useMutation({
    mutationFn: (payload) => (editing ? updateMenuSection(editing.uuid, payload) : createMenuSection(payload)),
    onSuccess: () => {
      toast.success(editing ? 'Section updated' : 'Section created')
      queryClient.invalidateQueries({ queryKey: ['menu-sections'] })
      setFormOpen(false)
      setEditing(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Save failed'),
  })

  const deleteMutation = useMutation({
    mutationFn: (uuid) => deleteMenuSection(uuid),
    onSuccess: () => {
      toast.success('Section deleted')
      queryClient.invalidateQueries({ queryKey: ['menu-sections'] })
      setDeleting(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Delete failed'),
  })

  const reorderMutation = useMutation({
    mutationFn: (payload) => reorderMenuSections(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['menu-sections'] }),
    onError: (err) => toast.error(err.response?.data?.message || 'Reorder failed'),
  })

  const sections = data?.data || []

  const move = (index, dir) => {
    const swapIndex = index + dir
    if (swapIndex < 0 || swapIndex >= sections.length) return
    const a = sections[index]
    const b = sections[swapIndex]
    reorderMutation.mutate([
      { uuid: a.uuid, order: b.order },
      { uuid: b.uuid, order: a.order },
    ])
  }

  const columns = [
    { key: 'name', label: 'Name' },
    { key: 'icon', label: 'Icon' },
    { key: 'order', label: 'Order' },
    { key: 'menu_count', label: 'Menus', render: (row) => row.menu_count ?? (row.menus?.length || 0) },
    {
      key: 'actions',
      label: '',
      render: (row, i) => (
        <div className="flex justify-end gap-1">
          <button onClick={() => move(i, -1)} disabled={i === 0} className="rounded p-1.5 hover:bg-surface-hover disabled:opacity-30">
            <ArrowUp size={14} />
          </button>
          <button onClick={() => move(i, 1)} disabled={i === sections.length - 1} className="rounded p-1.5 hover:bg-surface-hover disabled:opacity-30">
            <ArrowDown size={14} />
          </button>
          {canEdit('menu-sections') && (
            <button
              onClick={() => {
                setEditing(row)
                setFormOpen(true)
              }}
              className="rounded p-1.5 hover:bg-surface-hover"
            >
              <Pencil size={16} />
            </button>
          )}
          {canDelete('menu-sections') && (
            <button onClick={() => setDeleting(row)} className="rounded p-1.5 text-danger hover:bg-danger/10">
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
          <MenubarLabel>Menu Sections</MenubarLabel>
          {canCreate('menu-sections') && (
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
      <Table columns={columns} data={data?.data || []} loading={isLoading} />

      <MenuSectionFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={(values) => saveMutation.mutate(values)}
        initialData={editing}
        loading={saveMutation.isPending}
      />

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleteMutation.mutate(deleting.uuid)}
        loading={deleteMutation.isPending}
        message={`Delete section "${deleting?.name}"? This cannot be undone.`}
      />
    </div>
  )
}

export default MenuSections
