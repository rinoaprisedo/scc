import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Select from '../ui/Select'
import { shirtSchema } from '../../utils/validation'
import { updateMyProfile } from '../../api/peserta'
import blazerSizeGuide from '../../assets/blazer-size.webp'

// Reused as-is with the admin panel's peserta form so a value entered on
// either side always matches the same size domain.
const SIZE_OPTIONS = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL']

function ShirtSizeTab({ user, formId, onSaved, onSubmittingChange }) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(shirtSchema),
    defaultValues: { blazer_size: user.blazer_size || '' },
  })

  const [error, setError] = useState('')

  useEffect(() => () => onSubmittingChange(false), [onSubmittingChange])

  const submit = async (values) => {
    setError('')
    onSubmittingChange(true)
    try {
      await updateMyProfile(values)
      await onSaved()
    } catch (err) {
      setError(err.response?.data?.message || 'Gagal menyimpan, silakan coba lagi')
    } finally {
      onSubmittingChange(false)
    }
  }

  return (
    <form id={formId} className="flex flex-col gap-5" onSubmit={handleSubmit(submit)}>
      <div className="rounded-2xl border border-surface-border bg-surface-bg p-4">
        <p className="text-center text-sm font-bold text-navy">BLAZER — PRIA / WANITA</p>
        <img src={blazerSizeGuide} alt="Panduan ukuran blazer" className="mt-3 w-full rounded-lg" />
      </div>

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
        <Select label="Select Blazer Size" required error={errors.blazer_size?.message} {...register('blazer_size')}>
          <option value="">Select</option>
          {SIZE_OPTIONS.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </Select>
      </div>

      {error && <p className="text-sm font-medium text-red-600">{error}</p>}
    </form>
  )
}

export default ShirtSizeTab
