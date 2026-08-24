import { X, Mail, Phone, CreditCard, IdCard, MapPin, Plane, Utensils, Shirt, Cake, Armchair } from 'lucide-react'
import { fileURL } from '../utils/url'

function Field({ icon: Icon, label, value }) {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-navy/5 text-navy">
        <Icon size={15} />
      </div>
      <div className="min-w-0">
        <p className="text-xs font-medium text-text-secondary">{label}</p>
        <p className="truncate text-sm font-semibold text-text-primary">{value || '-'}</p>
      </div>
    </div>
  )
}

function formatDate(value) {
  if (!value) return null
  return new Date(value).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })
}

function fullName(u) {
  return [u.title, u.first_name, u.middle_name, u.last_name].filter(Boolean).join(' ') || u.name
}

function ProfileModal({ user, onClose, onEdit }) {
  if (!user) return null
  const originCity = user.origin_city
    ? `${user.origin_city.name}${user.origin_city.province ? `, ${user.origin_city.province}` : ''}`
    : user.origin_city_other
  const dietary = user.dietary_restriction
  const blazerSize = user.blazer_size ? `Blazer ${user.blazer_size}` : null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 py-8 backdrop-blur-sm">
      <div className="flex max-h-[88vh] w-full max-w-lg flex-col overflow-hidden rounded-2xl bg-white shadow-2xl">
        <div className="relative shrink-0 bg-gradient-to-br from-navy via-navy to-navy-light px-8 pb-8 pt-8 text-center">
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="absolute right-5 top-5 text-white/70 transition hover:text-white"
          >
            <X size={20} />
          </button>
          <div className="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-gold-light via-gold to-gold-dark text-3xl font-bold text-navy-dark shadow-md ring-4 ring-white/20">
            {user.name?.charAt(0).toUpperCase() || 'P'}
          </div>
          <p className="mt-3 text-lg font-bold text-white">{fullName(user)}</p>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-8 py-6">
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
            <Field icon={Mail} label="Email" value={user.email} />
            <Field icon={Cake} label="Tanggal Lahir" value={formatDate(user.birth_date)} />
            <Field icon={Phone} label="Nomor HP" value={user.phone_number} />
            <Field icon={CreditCard} label="Nomor KTP" value={user.nomor_ktp} />
            <Field icon={IdCard} label="Nomor Passport" value={user.passport_number} />
            <Field icon={IdCard} label="Masa Berlaku Passport" value={formatDate(user.passport_expiry)} />
            <Field icon={MapPin} label="Kota Asal" value={originCity} />
            <Field icon={Plane} label="Bandara Terdekat" value={user.nearest_airport?.name} />
            <Field icon={Utensils} label="Pantangan Makanan" value={dietary} />
            <Field icon={Shirt} label="Ukuran Baju" value={blazerSize} />
            <Field icon={Armchair} label="Nomor Meja" value={user.nomor_meja} />
          </div>

          {user.ktp_file && (
            <div className="mt-6">
              <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-text-secondary">KTP</p>
              <img
                src={fileURL(user.ktp_file)}
                alt="KTP"
                className="h-32 w-auto rounded-lg border border-surface-border object-cover"
              />
            </div>
          )}
        </div>

        <div className="shrink-0 border-t border-surface-border bg-white px-8 py-5">
          <button
            type="button"
            onClick={onEdit}
            className="w-full rounded-lg bg-gradient-to-r from-gold-light via-gold to-gold-dark py-3 text-sm font-bold uppercase tracking-wide text-navy-dark shadow-md transition hover:opacity-90"
          >
            Edit Formulir
          </button>
        </div>
      </div>
    </div>
  )
}

export default ProfileModal
