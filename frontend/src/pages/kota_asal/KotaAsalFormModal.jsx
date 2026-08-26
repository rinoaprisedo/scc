import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import Button from '../../components/ui/Button'
import { kotaAsalSchema } from '../../utils/validation'

function KotaAsalFormModal({ open, onClose, onSubmit, initialData, loading }) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({ resolver: zodResolver(kotaAsalSchema) })

  useEffect(() => {
    if (open) {
      reset(
        initialData
          ? {
              name: initialData.name,
              province: initialData.province || '',
              description: initialData.description || '',
            }
          : { name: '', province: '', description: '' },
      )
    }
  }, [open, initialData, reset])

  return (
    <Modal open={open} onClose={onClose} title={initialData ? 'Edit Kota Asal' : 'Add Kota Asal'}>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit(onSubmit)}>
        <Input label="Name" error={errors.name?.message} {...register('name')} />
        <Input label="Province" error={errors.province?.message} {...register('province')} />
        <div>
          <Input
            label="Description"
            placeholder="e.g. Jakarta, Bogor, Depok, Tangerang, Bekasi"
            error={errors.description?.message}
            {...register('description')}
          />
          <p className="mt-1 text-xs text-text-secondary">
            Extra search keywords — website users searching any of these words will also match this
            city (e.g. an entry named &quot;Jabodetabek&quot; found when searching &quot;Jakarta&quot;).
          </p>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

export default KotaAsalFormModal
