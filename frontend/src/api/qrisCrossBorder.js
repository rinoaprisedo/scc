import client from './client'

export const getQrisCrossBorder = (params) => client.get('/qris-cross-border', { params }).then((r) => r.data)
export const getQrisCrossBorderOne = (uuid) => client.get(`/qris-cross-border/${uuid}`).then((r) => r.data)
export const createQrisCrossBorder = (formData) =>
  client.post('/qris-cross-border', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const updateQrisCrossBorder = (uuid, formData) =>
  client.put(`/qris-cross-border/${uuid}`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const deleteQrisCrossBorder = (uuid) => client.delete(`/qris-cross-border/${uuid}`).then((r) => r.data)
export const updateQrisCrossBorderStatus = (uuid, status) =>
  client.put(`/qris-cross-border/${uuid}/status`, { status }).then((r) => r.data)
