import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Select from '../ui/Select'
import { shirtSchema } from '../../utils/validation'
import { updateMyProfile } from '../../api/peserta'

// Reused as-is with the admin panel's peserta form so a value entered on
// either side always matches the same size domain.
const SIZE_OPTIONS = ['S', 'M', 'L', 'XL', 'XXL', 'XXXL']

const BLAZER_GUIDE = [
  ['XS', '54.5', '64'],
  ['S', '57.5', '66'],
  ['M', '60.5', '68'],
  ['L', '63.5', '70'],
  ['XL', '67.5', '73'],
  ['XXL', '71.5', '74'],
]

function SizeGuideTable({ title, rows }) {
  return (
    <div className="rounded-2xl border border-surface-border bg-surface-bg p-4">
      <p className="text-center text-sm font-bold text-navy">{title}</p>
      <table className="mt-3 w-full text-xs text-text-primary">
        <thead>
          <tr className="border-b border-surface-border text-text-secondary">
            <th className="py-1 text-left font-semibold">Size</th>
            <th className="py-1 text-right font-semibold">Lebar</th>
            <th className="py-1 text-right font-semibold">Panjang</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(([size, lebar, panjang]) => (
            <tr key={size} className="border-b border-surface-border/60 last:border-0">
              <td className="py-1">{size}</td>
              <td className="py-1 text-right">{lebar}</td>
              <td className="py-1 text-right">{panjang}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className="mt-2 text-right text-[10px] text-text-secondary">*Satuan CM</p>
    </div>
  )
}

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
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <SizeGuideTable title="BLAZER — PRIA / WANITA" rows={BLAZER_GUIDE} />
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
