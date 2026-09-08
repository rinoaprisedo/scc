import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import toast from 'react-hot-toast'
import Modal from '../../components/ui/Modal'
import Button from '../../components/ui/Button'
import Select from '../../components/ui/Select'
import FileUpload from '../../components/ui/FileUpload'
import { sliderSchema } from '../../utils/validation'
import { compressImage } from '../../utils/imageCompress'
import { fileURL } from '../../utils/url'

function SliderFormModal({ open, onClose, onSubmit, initialData, defaultType, loading }) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm({ resolver: zodResolver(sliderSchema) })

  const [imageFile, setImageFile] = useState(null)
  const [imagePreview, setImagePreview] = useState(null)
  const [imageError, setImageError] = useState('')
  const [compressing, setCompressing] = useState(false)

  useEffect(() => {
    if (!open) return
    reset({ type: initialData?.type || defaultType, is_active: initialData?.is_active ?? true })
    setImageFile(null)
    setImagePreview(initialData?.image ? fileURL(initialData.image) : null)
    setImageError('')
  }, [open, initialData, defaultType, reset])

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
    formData.append('type', values.type)
    formData.append('is_active', String(values.is_active))
    if (imageFile) formData.append('image', imageFile)
    onSubmit(formData)
  }

  return (
    <Modal open={open} onClose={onClose} title={initialData ? 'Edit Slider' : 'Add Slider'}>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit(submit)}>
        <FileUpload
          label="Image"
          hint="PNG, JPEG or WebP — compressed automatically before upload"
          accept="image/png,image/jpeg,image/webp"
          preview={imagePreview}
          onFileSelect={handleImageSelect}
          previewClassName="h-20 w-32"
        />
        {compressing && <p className="-mt-2 text-xs text-text-secondary">Compressing image...</p>}
        {imageError && <p className="-mt-2 text-xs text-danger">{imageError}</p>}
        {errors.is_active && <p className="-mt-2 text-xs text-danger">{errors.is_active.message}</p>}

        <Select label="Type" error={errors.type?.message} {...register('type')}>
          <option value="desktop">Desktop</option>
          <option value="mobile">Mobile</option>
        </Select>

        <label className="flex items-center gap-2 text-sm text-text-primary">
          <input type="checkbox" className="h-4 w-4 accent-primary" {...register('is_active')} />
          Active
        </label>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading || compressing}>
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

export default SliderFormModal
