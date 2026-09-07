// Completeness is computed on read from the profile fields already on the
// user record rather than a separate "completed" flag, so there's nothing
// that can drift out of sync with the actual data.
export function isFormComplete(u) {
  if (!u) return false
  // A peserta whose passport genuinely has only one name never has a
  // last_name to give — mirrors backend/modules/peserta/service.go's
  // isFormComplete.
  const nameOK = Boolean(u.last_name) || Boolean(u.passport_single_name)
  return Boolean(
    u.title &&
      u.first_name &&
      nameOK &&
      u.birth_date &&
      (u.origin_city || u.origin_city_other) &&
      u.nearest_airport &&
      u.dietary_restriction &&
      u.phone_number &&
      u.ktp_number &&
      u.ktp_file &&
      u.passport_number &&
      u.passport_expiry,
  )
}

export function isShirtComplete(u) {
  if (!u) return false
  return Boolean(u.blazer_size)
}
