// Mirrors backend/modules/peserta's AttendanceStatus constants — the single
// field combining the attendance answer and form/shirt completeness.
export const ATTENDANCE_LABELS = {
  belum_konfirmasi: { text: 'Belum Konfirmasi', variant: 'neutral' },
  hadir_belum_lengkap: { text: 'Hadir, Belum Isi Form', variant: 'warning' },
  tidak_hadir: { text: 'Tidak Hadir', variant: 'danger' },
  hadir_lengkap: { text: 'Hadir, Data Lengkap', variant: 'success' },
}

export function attendanceLabel(status) {
  return ATTENDANCE_LABELS[status] || ATTENDANCE_LABELS.belum_konfirmasi
}
