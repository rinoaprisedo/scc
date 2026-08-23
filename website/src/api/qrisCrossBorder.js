import client from './client'

export const getMyQrisCrossBorder = (params) => client.get('/qris-cross-border/my', { params }).then((r) => r.data)
export const createMyQrisCrossBorder = (formData) =>
  client.post('/qris-cross-border/my', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const deleteMyQrisCrossBorder = (uuid) => client.delete(`/qris-cross-border/my/${uuid}`).then((r) => r.data)
