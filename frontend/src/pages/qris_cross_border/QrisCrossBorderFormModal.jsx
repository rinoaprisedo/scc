import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import toast from 'react-hot-toast'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import Select from '../../components/ui/Select'
import Button from '../../components/ui/Button'
import FileUpload from '../../components/ui/FileUpload'
import AsyncSelect from '../../components/ui/AsyncSelect'
import { qrisCrossBorderSchema } from '../../utils/validation'
import { searchPesertaOptions } from '../../api/peserta'
import { compressImage } from '../../utils/imageCompress'
import { fileURL } from '../../utils/url'

function QrisCrossBorderFormModal({ open, onClose, onSubmit, initialData, loading }) {
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm({ resolver: zodResolver(qrisCrossBorderSchema) })

  const [imageFile, setImageFile] = useState(null)
  const [imagePreview, setImagePreview] = useState(null)
  const [imageError, setImageError] = useState('')
  const [compressing, setCompressing] = useState(false)

  const pesertaUuid = watch('peserta_uuid')
  const isPending = initialData?.status === 'pending'

  useEffect(() => {
    if (!open) return
    reset(
      initialData
        ? {
            peserta_uuid: initialData.peserta_uuid,
            nominal_asing: initialData.nominal_asing ?? 0,
            nominal_rupiah: initialData.nominal_rupiah ?? 0,
            merchant_name: initialData.merchant_name ?? '',
            reference_number: initialData.reference_number ?? '',
            status: initialData.status,
          }
        : { peserta_uuid: '' },
    )
    setImageFile(null)
    setImagePreview(initialData?.image ? fileURL(initialData.image) : null)
    setImageError('')
  }, [open, initialData, reset])

  const handleImageSelect = async (file) => {
    setCompressing(true)
    try {
      const compressed = await compressImage(file)
      setImageFile(compressed)
      setImagePreview(URL.createObjectURL(compressed))
      setImageError('')
    } catch {
      toast.error('Failed to process image')
    } finally {
      setCompressing(false)
    }
  }

  const submit = (values) => {
    if (!initialData && !imageFile) {
      setImageError('Image is required')
      return
    }
    const formData = new FormData()
    formData.append('peserta_uuid', values.peserta_uuid)
    if (initialData) {
      formData.append('nominal_asing', values.nominal_asing === '' ? '0' : String(values.nominal_asing))
      formData.append('nominal_rupiah', values.nominal_rupiah === '' ? '0' : String(values.nominal_rupiah))
      formData.append('merchant_name', values.merchant_name || '')
      formData.append('reference_number', values.reference_number || '')
      formData.append('status', values.status)
    }
    if (imageFile) formData.append('image', imageFile)
    onSubmit(formData)
  }

  return (
    <Modal open={open} onClose={onClose} title={initialData ? 'Edit Qris Cross Border' : 'Add Qris Cross Border'}>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit(submit)}>
        {!initialData && (
          <p className="-mt-1 text-xs text-text-secondary">
            Nominal, nama merchant, dan no. referensi akan diisi otomatis setelah bukti pembayaran diproses.
          </p>
        )}
        {isPending && (
          <p className="-mt-1 rounded-md bg-warning-bg px-3 py-2 text-xs text-warning">
            Sedang diproses otomatis oleh AI — data belum bisa diedit sampai proses selesai.
          </p>
        )}

        <AsyncSelect
          label="Peserta"
          placeholder="Search peserta by name/NIK..."
          value={pesertaUuid}
          initialLabel={initialData?.peserta_name}
          error={errors.peserta_uuid?.message}
          loadOptions={searchPesertaOptions}
          getOptionValue={(o) => o.uuid}
          getOptionLabel={(o) => o.name}
          getOptionSublabel={(o) => o.ktp_number}
          onChange={(uuid) => setValue('peserta_uuid', uuid, { shouldValidate: true })}
        />

        <FileUpload
          label="Proof of Payment Image"
          hint="PNG, JPEG or WebP — compressed automatically before upload"
          accept="image/png,image/jpeg,image/webp"
          preview={imagePreview}
          onFileSelect={handleImageSelect}
          previewClassName="h-20 w-20"
        />
        {compressing && <p className="-mt-2 text-xs text-text-secondary">Compressing image...</p>}
        {imageError && <p className="-mt-2 text-xs text-danger">{imageError}</p>}

        {initialData && (
          <>
            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Nama Merchant"
                disabled={isPending}
                error={errors.merchant_name?.message}
                {...register('merchant_name')}
              />
              <Input
                label="No. Referensi"
                disabled={isPending}
                error={errors.reference_number?.message}
                {...register('reference_number')}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Nominal Asing"
                type="number"
                step="0.01"
                min={0}
                disabled={isPending}
                error={errors.nominal_asing?.message}
                {...register('nominal_asing')}
              />
              <Input
                label="Nominal Rupiah (IDR)"
                type="number"
                step="0.01"
                min={0}
                disabled={isPending}
                error={errors.nominal_rupiah?.message}
                {...register('nominal_rupiah')}
              />
            </div>

            <Select label="Status" disabled={isPending} error={errors.status?.message} {...register('status')}>
              <option value="pending">Pending</option>
              <option value="waiting_approval">Waiting Approval</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
            </Select>
          </>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading || compressing || isPending}>
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

export default QrisCrossBorderFormModal
