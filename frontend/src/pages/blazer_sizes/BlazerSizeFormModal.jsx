import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import Button from '../../components/ui/Button'
import { blazerSizeSchema } from '../../utils/validation'

function BlazerSizeFormModal({ open, onClose, onSubmit, initialData, loading }) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({ resolver: zodResolver(blazerSizeSchema) })

  useEffect(() => {
    if (open) {
      reset(
        initialData
          ? { size: initialData.size, stock: initialData.stock }
          : { size: '', stock: 0 },
      )
    }
  }, [open, initialData, reset])

  return (
    <Modal open={open} onClose={onClose} title={initialData ? 'Edit Blazer Size' : 'Add Blazer Size'}>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit(onSubmit)}>
        <Input label="Size" error={errors.size?.message} {...register('size')} />
        <Input label="Stock" type="number" min={0} error={errors.stock?.message} {...register('stock')} />
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

export default BlazerSizeFormModal
