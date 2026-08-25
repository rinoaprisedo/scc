import client from './client'

export const updateMyProfile = (payload) => client.put('/peserta/me', payload).then((r) => r.data)
export const uploadMyKtp = (formData) =>
  client.post('/peserta/me/ktp', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const uploadMyPassport = (formData) =>
  client.post('/peserta/me/passport', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const setMyAttendance = (confirmed) =>
  client.put('/peserta/me/attendance', { confirmed }).then((r) => r.data)
