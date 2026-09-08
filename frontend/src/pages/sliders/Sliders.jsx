import { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Plus, Pencil, Trash2, GripVertical } from 'lucide-react'
import EmptyState from '../../components/ui/EmptyState'
import { TableSkeleton } from '../../components/ui/Skeleton'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import Badge from '../../components/ui/Badge'
import { Menubar, MenubarAction, MenubarLabel, MenubarSeparator } from '../../components/ui/Menubar'
import SliderFormModal from './SliderFormModal'
import { getSliders, createSlider, updateSlider, deleteSlider, reorderSliders } from '../../api/sliders'
import { fileURL } from '../../utils/url'
import usePermission from '../../hooks/usePermission'
import usePageActions from '../../hooks/usePageActions'

const TYPE_TABS = [
  { value: 'desktop', label: 'Desktop' },
  { value: 'mobile', label: 'Mobile' },
]

function Sliders() {
  const { canCreate, canEdit, canDelete } = usePermission()
  const queryClient = useQueryClient()
  const [type, setType] = useState('desktop')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [deleting, setDeleting] = useState(null)
  const [items, setItems] = useState([])
  const [dragOverIndex, setDragOverIndex] = useState(null)
  const dragIndexRef = useRef(null)
  const draggable = canEdit('sliders')

  const { data, isLoading } = useQuery({
    queryKey: ['sliders', type],
    queryFn: () => getSliders({ limit: 100, sort_by: 'order', sort_dir: 'asc', type }),
  })

  // Local copy so a drag can preview the new order immediately instead of
  // waiting on the reorder request round-trip.
  useEffect(() => {
    setItems(data?.data || [])
  }, [data])

  const saveMutation = useMutation({
    mutationFn: (payload) => (editing ? updateSlider(editing.uuid, payload) : createSlider(payload)),
    onSuccess: () => {
      toast.success(editing ? 'Slider updated' : 'Slider created')
      queryClient.invalidateQueries({ queryKey: ['sliders'] })
      setFormOpen(false)
      setEditing(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Save failed'),
  })

  const deleteMutation = useMutation({
    mutationFn: (uuid) => deleteSlider(uuid),
    onSuccess: () => {
      toast.success('Slider deleted')
      queryClient.invalidateQueries({ queryKey: ['sliders'] })
      setDeleting(null)
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Delete failed'),
  })

  const reorderMutation = useMutation({
    mutationFn: (payload) => reorderSliders(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['sliders'] }),
    onError: (err) => {
      toast.error(err.response?.data?.message || 'Reorder failed')
      setItems(data?.data || [])
    },
  })

  const handleDragStart = (index) => {
    dragIndexRef.current = index
  }

  const handleDragOver = (e, index) => {
    e.preventDefault()
    setDragOverIndex(index)
  }

  const handleDrop = (index) => {
    const from = dragIndexRef.current
    dragIndexRef.current = null
    setDragOverIndex(null)
    if (from === null || from === index) return

    const next = [...items]
    const [moved] = next.splice(from, 1)
    next.splice(index, 0, moved)
    setItems(next)
    reorderMutation.mutate(next.map((item, i) => ({ uuid: item.uuid, order: i })))
  }

  usePageActions(
    useMemo(
      () => (
        <Menubar>
          <MenubarLabel>Sliders</MenubarLabel>
          {canCreate('sliders') && (
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
      <div className="flex gap-1 border-b border-surface-border">
        {TYPE_TABS.map((t) => (
          <button
            key={t.value}
            onClick={() => setType(t.value)}
            className={`px-4 py-2 text-sm font-medium ${
              type === t.value ? 'border-b-2 border-primary text-primary' : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="rounded-lg border border-surface-border bg-surface-card shadow-card">
        {isLoading ? (
          <TableSkeleton cols={3} />
        ) : items.length === 0 ? (
          <EmptyState title="No records found" />
        ) : (
          <table className="w-full text-left text-[13px]">
            <thead>
              <tr className="border-b border-surface-border text-text-tertiary">
                <th className="w-10 px-5 py-[11px]" />
                <th className="px-5 py-[11px] text-[11px] font-semibold uppercase tracking-[.07em]">Image</th>
                <th className="px-5 py-[11px] text-[11px] font-semibold uppercase tracking-[.07em]">Status</th>
                <th className="px-5 py-[11px]" />
              </tr>
            </thead>
            <tbody>
              {items.map((row, i) => (
                <tr
                  key={row.uuid}
                  draggable={draggable}
                  onDragStart={() => handleDragStart(i)}
                  onDragOver={(e) => handleDragOver(e, i)}
                  onDrop={() => handleDrop(i)}
                  onDragEnd={() => {
                    dragIndexRef.current = null
                    setDragOverIndex(null)
                  }}
                  className={`border-b border-surface-border transition-colors duration-100 last:border-0 ${
                    dragOverIndex === i ? 'bg-surface-hover' : ''
                  }`}
                >
                  <td className="px-5 py-[13px] text-text-tertiary">
                    {draggable && <GripVertical size={16} className="cursor-grab active:cursor-grabbing" />}
                  </td>
                  <td className="px-5 py-[13px]">
                    <div className="h-14 w-24 overflow-hidden rounded-md border border-surface-border bg-surface-card">
                      <img src={fileURL(row.image)} alt="" className="h-full w-full object-cover" />
                    </div>
                  </td>
                  <td className="px-5 py-[13px] text-text-primary">
                    <Badge variant={row.is_active ? 'success' : 'neutral'}>{row.is_active ? 'active' : 'inactive'}</Badge>
                  </td>
                  <td className="px-5 py-[13px]">
                    <div className="flex justify-end gap-1">
                      {canEdit('sliders') && (
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
                      {canDelete('sliders') && (
                        <button
                          onClick={() => setDeleting(row)}
                          className="rounded p-1.5 text-danger hover:bg-danger/10"
                          title="Delete"
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <SliderFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={(values) => saveMutation.mutate(values)}
        initialData={editing}
        defaultType={type}
        loading={saveMutation.isPending}
      />

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleteMutation.mutate(deleting.uuid)}
        loading={deleteMutation.isPending}
        message="Delete this slider? This cannot be undone."
      />
    </div>
  )
}

export default Sliders
