import { useEffect, useState } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Input from '../ui/Input'
import Select from '../ui/Select'
import Combobox from '../ui/Combobox'
import FileUpload from '../ui/FileUpload'
import { formSchema } from '../../utils/validation'
import { updateMyProfile, uploadMyKtp } from '../../api/peserta'
import { fileURL } from '../../utils/url'

const digitsOnly = (e) => {
  e.target.value = e.target.value.replace(/\D/g, '')
}

const TITLE_OPTIONS = ['Mr', 'Mrs', 'Ms']
const DIETARY_OPTIONS = ['Tidak Ada', 'Vegetarian', 'Vegan', 'Alergi Seafood', 'Other']
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
    dietary_restriction_other: user.dietary_restriction_other || '',
    phone_number: user.phone_number || '',
    nomor_ktp: user.nomor_ktp || '',
    passport_number: user.passport_number || '',
    passport_expiry: user.passport_expiry ? user.passport_expiry.slice(0, 10) : '',
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
    ...kotaAsal.map((c) => ({ value: c.uuid, label: c.name + (c.province ? `, ${c.province}` : '') })),
    { value: OTHER_CITY_VALUE, label: 'Other' },
  ]
  const airportOptions = bandara.map((b) => ({ value: b.uuid, label: b.name }))

  const [ktpFile, setKtpFile] = useState(null)
  const [ktpPreview, setKtpPreview] = useState(fileURL(user.ktp_file))
  const [ktpError, setKtpError] = useState('')
  const [error, setError] = useState('')

  const originCityUuid = watch('origin_city_uuid')
  const dietaryRestriction = watch('dietary_restriction')

  useEffect(() => () => onSubmittingChange(false), [onSubmittingChange])

  const submit = async (values) => {
    if (!ktpPreview) {
      setKtpError('Mohon upload copy KTP Anda')
      return
    }
    setKtpError('')
    setError('')
    onSubmittingChange(true)
    try {
      const payload = { ...values }
      if (payload.origin_city_uuid === OTHER_CITY_VALUE) {
        payload.origin_city_uuid = ''
      } else {
        payload.origin_city_other = ''
      }
      if (payload.dietary_restriction !== 'Other') {
        payload.dietary_restriction_other = ''
      }

      await updateMyProfile(payload)
      if (ktpFile) {
        const formData = new FormData()
        formData.append('ktp_file', ktpFile)
        await uploadMyKtp(formData)
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

      <Input
        label="Nama Depan (Sesuai Passport)"
        required
        error={errors.first_name?.message}
        {...register('first_name')}
      />
      <Input label="Nama Tengah (Sesuai Passport)" error={errors.middle_name?.message} {...register('middle_name')} />
      <Input
        label="Nama Belakang (Sesuai Passport)"
        required
        className="sm:col-span-2"
        error={errors.last_name?.message}
        {...register('last_name')}
      />

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
      {dietaryRestriction === 'Other' && (
        <Input
          label="Masukan pantangan makanan Anda"
          required
          className="sm:col-span-2"
          error={errors.dietary_restriction_other?.message}
          {...register('dietary_restriction_other')}
        />
      )}

      <Input
        label="Nomor KTP"
        inputMode="numeric"
        error={errors.nomor_ktp?.message}
        {...register('nomor_ktp', { onChange: digitsOnly })}
      />
      <Input label="Nomor Passport" required error={errors.passport_number?.message} {...register('passport_number')} />

      <FileUpload
        className="sm:col-span-2"
        label="Mohon upload copy KTP Anda"
        required
        hint="JPEG, PNG, or WebP"
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

      <Input
        label="Masa Berlaku Passport"
        required
        className="sm:col-span-2"
        type="date"
        error={errors.passport_expiry?.message}
        {...register('passport_expiry')}
      />

      {error && <p className="text-sm font-medium text-red-600 sm:col-span-2">{error}</p>}
    </form>
  )
}

export default FormTab
