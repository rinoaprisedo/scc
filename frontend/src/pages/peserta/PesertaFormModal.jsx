import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import Select from '../../components/ui/Select'
import Button from '../../components/ui/Button'
import FileUpload from '../../components/ui/FileUpload'
import { pesertaSchema } from '../../utils/validation'

const DIETARY_OPTIONS = [
  'tidak ada pantangan',
  'tidak makan daging',
  'tidak makan ayam',
  'tidak makan seafood',
  'vegetarian',
]
const SIZE_OPTIONS = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL']
const OTHER_CITY_VALUE = 'other'
const DEFAULT_PASSWORD = 'scc2026'
const MAX_KTP_FILE_SIZE = 3 * 1024 * 1024

function Req({ children }) {
  return (
    <>
      {children} <span className="text-danger">*</span>
    </>
  )
}

function emptyValues() {
  return {
    name: '',
    ktp_number: '',
    nomor_ktp: '',
    password: '',
    status: 'active',
    email: '',
    title: 'Mr',
    first_name: '',
    middle_name: '',
    last_name: '',
    birth_date: '',
    origin_city_uuid: '',
    origin_city_other: '',
    nearest_airport_uuid: '',
    dietary_restriction: '',
    phone_number: '',
    passport_number: '',
    passport_expiry: '',
    blazer_size: '',
    nomor_meja: '',
  }
}

function PesertaFormModal({ open, onClose, onSubmit, initialData, kotaAsal = [], bandara = [], loading }) {
  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm({ resolver: zodResolver(pesertaSchema) })

  const [ktpFile, setKtpFile] = useState(null)
  const [ktpPreview, setKtpPreview] = useState(null)
  const [ktpError, setKtpError] = useState('')

  useEffect(() => {
    if (open) {
      reset(
        initialData
          ? {
              name: initialData.name,
              ktp_number: initialData.ktp_number || '',
              nomor_ktp: initialData.nomor_ktp || '',
              password: '',
              status: initialData.status || 'active',
              email: initialData.email || '',
              title: initialData.title || 'Mr',
              first_name: initialData.first_name || '',
              middle_name: initialData.middle_name || '',
              last_name: initialData.last_name || '',
              birth_date: initialData.birth_date ? initialData.birth_date.slice(0, 10) : '',
              origin_city_uuid: initialData.origin_city?.uuid || (initialData.origin_city_other ? OTHER_CITY_VALUE : ''),
              origin_city_other: initialData.origin_city_other || '',
              nearest_airport_uuid: initialData.nearest_airport?.uuid || '',
              dietary_restriction: initialData.dietary_restriction || '',
              phone_number: initialData.phone_number || '',
              passport_number: initialData.passport_number || '',
              passport_expiry: initialData.passport_expiry ? initialData.passport_expiry.slice(0, 10) : '',
              blazer_size: initialData.blazer_size || '',
              nomor_meja: initialData.nomor_meja || '',
            }
          : emptyValues(),
      )
      setKtpFile(null)
      setKtpPreview(initialData?.ktp_file || null)
      setKtpError('')
    }
  }, [open, initialData, reset])

  const originCityUuid = watch('origin_city_uuid')

  const submit = (values) => {
    const payload = { ...values }
    if (!initialData) {
      payload.password = DEFAULT_PASSWORD
    }
    if (payload.origin_city_uuid === OTHER_CITY_VALUE) {
      payload.origin_city_uuid = ''
    } else {
      payload.origin_city_other = ''
    }
    onSubmit(payload, ktpFile)
  }

  return (
    <Modal open={open} onClose={onClose} title={initialData ? 'Edit Peserta' : 'Add Peserta'} size="xl">
      <form className="flex flex-col gap-4" onSubmit={handleSubmit(submit)}>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Input label={<Req>Name</Req>} error={errors.name?.message} {...register('name')} />
          <Input label={<Req>NIK</Req>} error={errors.ktp_number?.message} {...register('ktp_number')} />
          <Input label="Nomor KTP" error={errors.nomor_ktp?.message} {...register('nomor_ktp')} />

          {initialData && (
            <Input
              label="Password (leave blank to keep current)"
              type="password"
              error={errors.password?.message}
              {...register('password')}
            />
          )}
          <Select label="Status" error={errors.status?.message} {...register('status')}>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="suspended">Suspended</option>
          </Select>

          <Select label="Title" error={errors.title?.message} {...register('title')}>
            <option value="Mr">Mr</option>
            <option value="Mrs">Mrs</option>
            <option value="Ms">Ms</option>
          </Select>

          <Input
            label="Nama Depan (Sesuai Passport)"
            error={errors.first_name?.message}
            {...register('first_name')}
          />
          <Input
            label="Nama Tengah (Sesuai Passport)"
            error={errors.middle_name?.message}
            {...register('middle_name')}
          />

          <Input
            label="Nama Belakang (Sesuai Passport)"
            error={errors.last_name?.message}
            {...register('last_name')}
          />
          <Input label="Tanggal Lahir" type="date" error={errors.birth_date?.message} {...register('birth_date')} />

          <Select label="Kota Asal Keberangkatan" error={errors.origin_city_uuid?.message} {...register('origin_city_uuid')}>
            <option value="">Select city</option>
            {kotaAsal.map((c) => (
              <option key={c.uuid} value={c.uuid}>
                {c.name}
                {c.province ? `, ${c.province}` : ''}
              </option>
            ))}
            <option value={OTHER_CITY_VALUE}>Other</option>
          </Select>
          <Select label="Bandara Terdekat" error={errors.nearest_airport_uuid?.message} {...register('nearest_airport_uuid')}>
            <option value="">Select airport</option>
            {bandara.map((b) => (
              <option key={b.uuid} value={b.uuid}>
                {b.name}
              </option>
            ))}
          </Select>

          {originCityUuid === OTHER_CITY_VALUE && (
            <Input
              className="sm:col-span-2"
              label="Kota Asal Lainnya (Kota, Provinsi)"
              placeholder="e.g. Kudus, Jawa Tengah"
              error={errors.origin_city_other?.message}
              {...register('origin_city_other')}
            />
          )}

          <Select
            label="Apakah ada pantangan makanan?"
            error={errors.dietary_restriction?.message}
            {...register('dietary_restriction')}
          >
            <option value="">Select</option>
            {DIETARY_OPTIONS.map((o) => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </Select>
          <Input
            label="Nomor HP (contoh: 081234567890)"
            error={errors.phone_number?.message}
            {...register('phone_number')}
          />

          <Input label="Nomor Passport" error={errors.passport_number?.message} {...register('passport_number')} />
          <Input
            label="Masa Berlaku Passport"
            type="date"
            error={errors.passport_expiry?.message}
            {...register('passport_expiry')}
          />

          <Select label="Select Blazer Size" error={errors.blazer_size?.message} {...register('blazer_size')}>
            <option value="">Select</option>
            {SIZE_OPTIONS.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </Select>
          <Input label="Nomor Meja" error={errors.nomor_meja?.message} {...register('nomor_meja')} />

          <div className="sm:col-span-2">
            <FileUpload
              label="Mohon upload copy KTP Anda"
              hint="JPEG, PNG, or WebP"
              error={ktpError}
              preview={ktpPreview}
              onFileSelect={(file) => {
                if (file.size > MAX_KTP_FILE_SIZE) {
                  setKtpError('Ukuran file maksimal 3MB')
                  return
                }
                setKtpFile(file)
                setKtpPreview(URL.createObjectURL(file))
                setKtpError('')
              }}
              onClear={() => {
                setKtpFile(null)
                setKtpPreview(null)
              }}
            />
            {!ktpError && <p className="mt-1.5 text-xs font-medium text-danger">Ukuran file maksimal 3MB</p>}
          </div>
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

export default PesertaFormModal
