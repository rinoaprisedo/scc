import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import Button from '../../components/ui/Button'
import { qrGateSchema } from '../../utils/validation'

function QrGateFormModal({ open, onClose, onSubmit, initialData, loading }) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({ resolver: zodResolver(qrGateSchema) })

  useEffect(() => {
    if (open) {
      reset(
        initialData
          ? {
              name: initialData.name,
              code: initialData.code,
              points: initialData.points,
              is_reusable: initialData.is_reusable,
            }
          : { name: '', code: '', points: 0, is_reusable: true },
      )
    }
  }, [open, initialData, reset])

  return (
    <Modal open={open} onClose={onClose} title={initialData ? 'Edit QR Gate' : 'Add QR Gate'}>
      <form
        className="flex flex-col gap-4"
        onSubmit={handleSubmit((values) => onSubmit({ ...values, points: Number(values.points) }))}
      >
        <Input label="Name" error={errors.name?.message} {...register('name')} />
        <Input label="Code" error={errors.code?.message} {...register('code')} />
        <p className="-mt-3 text-xs text-text-secondary">
          Encoded into the QR image — letters, numbers, dash, underscore only.
        </p>
        <Input label="Points" type="number" min={0} error={errors.points?.message} {...register('points')} />
        <label className="flex items-center gap-3">
          <input type="checkbox" className="h-4 w-4 accent-primary" {...register('is_reusable')} />
          <span className="text-sm text-text-primary">
            Reusable (can be scanned by multiple participants, once each — otherwise only the first scan ever
            counts)
          </span>
        </label>
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

export default QrGateFormModal
