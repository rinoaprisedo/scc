import { useEffect, useState } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Input from '../ui/Input'
import Select from '../ui/Select'
import Combobox from '../ui/Combobox'
import FileUpload from '../ui/FileUpload'
import { formSchema } from '../../utils/validation'
import { updateMyProfile, uploadMyKtp, uploadMyPassport } from '../../api/peserta'
import { fileURL } from '../../utils/url'
import { isFormComplete } from '../../utils/onboarding'

const digitsOnly = (e) => {
  e.target.value = e.target.value.replace(/\D/g, '')
}

const TITLE_OPTIONS = ['Mr', 'Mrs', 'Ms']
const DIETARY_OPTIONS = [
  'tidak ada pantangan',
  'tidak makan daging',
  'tidak makan ayam',
  'tidak makan seafood',
  'vegetarian',
]
const OTHER_CITY_VALUE = 'other'

function defaultsFrom(user) {
  return {
    title: user.title || '',
    first_name: user.first_name || '',
    middle_name: user.middle_name || '',
    last_name: user.last_name || '',
    birth_date: user.birth_date ? user.birth_date.slice(0, 10) : '',
    origin_city_uuid: user.origin_city?.uuid || (user.origin_city_other ? OTHER_CITY_VALUE : ''),
    origin_city_other: user.origin_city_other || '',
    nearest_airport_uuid: user.nearest_airport?.uuid || '',
    dietary_restriction: user.dietary_restriction || '',
    phone_number: user.phone_number || '',
    nomor_ktp: user.nomor_ktp || '',
    passport_number: user.passport_number || '',
    passport_expiry: user.passport_expiry ? user.passport_expiry.slice(0, 10) : '',
    passport_single_name: user.passport_single_name || false,
  }
}

function FormTab({ user, kotaAsal, bandara, formId, onSaved, onSubmittingChange }) {
  const {
    register,
    handleSubmit,
    watch,
    control,
    formState: { errors },
  } = useForm({ resolver: zodResolver(formSchema), defaultValues: defaultsFrom(user) })

  const cityOptions = [
    { value: OTHER_CITY_VALUE, label: 'Other' },
    ...kotaAsal.map((c) => ({
      value: c.uuid,
      label: c.name + (c.province ? `, ${c.province}` : ''),
      keywords: c.description,
    })),
  ]
  const airportOptions = bandara.map((b) => ({ value: b.uuid, label: b.name }))

  const [ktpFile, setKtpFile] = useState(null)
  const [ktpPreview, setKtpPreview] = useState(fileURL(user.ktp_file))
  const [ktpError, setKtpError] = useState('')
  const [passportFile, setPassportFile] = useState(null)
  const [passportPreview, setPassportPreview] = useState(fileURL(user.passport_file))
  const [passportError, setPassportError] = useState('')
  const [error, setError] = useState('')

  // Passport scan became a required upload after some participants had
  // already finished onboarding without it — grandfather anyone who was
  // already complete under the old rules instead of retroactively blocking
  // their edits on a file they were never asked for.
  const [passportGrandfathered] = useState(() => isFormComplete(user))

  const originCityUuid = watch('origin_city_uuid')
  const passportSingleName = watch('passport_single_name')

  useEffect(() => () => onSubmittingChange(false), [onSubmittingChange])

  const submit = async (values) => {
    if (!ktpPreview) {
      setKtpError('Mohon upload copy KTP Anda')
      return
    }
    if (!passportPreview && !passportGrandfathered) {
      setPassportError('Mohon upload copy paspor Anda')
      return
    }
    setKtpError('')
    setPassportError('')
    setError('')
    onSubmittingChange(true)
    try {
      const payload = { ...values }
      if (payload.origin_city_uuid === OTHER_CITY_VALUE) {
        payload.origin_city_uuid = ''
      } else {
        payload.origin_city_other = ''
      }

      await updateMyProfile(payload)
      if (ktpFile) {
        const formData = new FormData()
        formData.append('ktp_file', ktpFile)
        await uploadMyKtp(formData)
      }
      if (passportFile) {
        const formData = new FormData()
        formData.append('passport_file', passportFile)
        await uploadMyPassport(formData)
      }
      await onSaved()
    } catch (err) {
      setError(err.response?.data?.message || 'Gagal menyimpan, silakan coba lagi')
    } finally {
      onSubmittingChange(false)
    }
  }

  return (
    <form id={formId} className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2" onSubmit={handleSubmit(submit)}>
      <Select label="Title" required error={errors.title?.message} {...register('title')}>
        <option value="">Select</option>
        {TITLE_OPTIONS.map((t) => (
          <option key={t} value={t}>
            {t}
          </option>
        ))}
      </Select>
      <Input label="Tanggal Lahir" required type="date" error={errors.birth_date?.message} {...register('birth_date')} />

      <div className="sm:col-span-2">
        <Input
          label="Nama Depan (Sesuai Paspor)"
          required
          error={errors.first_name?.message}
          {...register('first_name')}
        />
        <label className="mt-2 flex items-center gap-2 text-xs font-medium text-text-secondary">
          <input type="checkbox" className="h-4 w-4 rounded border-surface-border" {...register('passport_single_name')} />
          Nama saya di paspor hanya terdiri 1 nama
        </label>
      </div>
      {!passportSingleName && (
        <>
          <Input label="Nama Tengah (Sesuai Paspor)" error={errors.middle_name?.message} {...register('middle_name')} />
          <Input
            label="Nama Belakang (Sesuai Paspor)"
            required
            error={errors.last_name?.message}
            {...register('last_name')}
          />
        </>
      )}

      <div>
        <Controller
          name="origin_city_uuid"
          control={control}
          render={({ field }) => (
            <Combobox
              label="Kota Asal Keberangkatan"
              required
              placeholder="Ketik untuk cari kota..."
              error={errors.origin_city_uuid?.message}
              options={cityOptions}
              value={field.value}
              onChange={field.onChange}
            />
          )}
        />
        <p className="mt-1.5 text-xs text-text-secondary">
          Isi dengan kota tempat tinggal saat ini. Jika pilihan tidak tersedia, mohon isi &apos;Others&apos; dengan
          format (Kota, Provinsi)
        </p>
      </div>
      <Controller
        name="nearest_airport_uuid"
        control={control}
        render={({ field }) => (
          <Combobox
            label="Bandara Terdekat"
            required
            placeholder="Ketik untuk cari bandara..."
            error={errors.nearest_airport_uuid?.message}
            options={airportOptions}
            value={field.value}
            onChange={field.onChange}
          />
        )}
      />
      {originCityUuid === OTHER_CITY_VALUE && (
        <Input
          label="Kota Asal Lainnya (Kota, Provinsi)"
          required
          className="sm:col-span-2"
          placeholder="Contoh: Kudus, Jawa Tengah"
          error={errors.origin_city_other?.message}
          {...register('origin_city_other')}
        />
      )}

      <Select
        label="Apakah ada pantangan makanan?"
        required
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
        required
        inputMode="numeric"
        error={errors.phone_number?.message}
        {...register('phone_number', { onChange: digitsOnly })}
      />

      <Input
        label="Nomor KTP"
        required
        inputMode="numeric"
        error={errors.nomor_ktp?.message}
        {...register('nomor_ktp', { onChange: digitsOnly })}
      />
      <Input label="Nomor Paspor" required error={errors.passport_number?.message} {...register('passport_number')} />

      <div className="sm:col-span-2">
        <FileUpload
          label="Mohon upload copy KTP Anda"
          required
          hint="JPEG, PNG, or WebP"
          sizeNote="Ukuran file maksimal 10MB"
          error={ktpError}
          preview={ktpPreview}
          onFileSelect={(file) => {
            setKtpFile(file)
            setKtpPreview(URL.createObjectURL(file))
            setKtpError('')
          }}
          onClear={() => {
            setKtpFile(null)
            setKtpPreview(null)
          }}
        />
      </div>

      <div className="sm:col-span-2">
        <FileUpload
          label="Mohon upload copy paspor Anda"
          required={!passportGrandfathered}
          hint="JPEG, PNG, or WebP"
          sizeNote="Ukuran file maksimal 10MB"
          error={passportError}
          preview={passportPreview}
          alt="Paspor preview"
          onFileSelect={(file) => {
            setPassportFile(file)
            setPassportPreview(URL.createObjectURL(file))
            setPassportError('')
          }}
          onClear={() => {
            setPassportFile(null)
            setPassportPreview(null)
          }}
        />
      </div>

      <div className="sm:col-span-2">
        <Input
          label="Masa Berlaku Paspor"
          required
          type="date"
          error={errors.passport_expiry?.message}
          {...register('passport_expiry')}
        />
        <p className="mt-1.5 text-xs text-text-secondary">
          Paspor harus berlaku minimal 6 bulan setelah 10 Oktober 2026 (berlaku hingga minimal 10 April 2027)
        </p>
      </div>

      {error && <p className="text-sm font-medium text-red-600 sm:col-span-2">{error}</p>}
    </form>
  )
}

export default FormTab
