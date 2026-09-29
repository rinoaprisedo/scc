import client from './client'

function downloadFile(blob, filename) {
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  window.URL.revokeObjectURL(url)
}

export const getQrisCrossBorder = (params) => client.get('/qris-cross-border', { params }).then((r) => r.data)
export const exportQrisCrossBorderExcel = (params) =>
  client
    .get('/qris-cross-border/export/excel', { params, responseType: 'blob' })
    .then((r) => downloadFile(r.data, 'qris_cross_border.xlsx'))
export const getQrisCrossBorderOne = (uuid) => client.get(`/qris-cross-border/${uuid}`).then((r) => r.data)
export const createQrisCrossBorder = (formData) =>
  client.post('/qris-cross-border', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const updateQrisCrossBorder = (uuid, formData) =>
  client.put(`/qris-cross-border/${uuid}`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const deleteQrisCrossBorder = (uuid) => client.delete(`/qris-cross-border/${uuid}`).then((r) => r.data)
export const updateQrisCrossBorderStatus = (uuid, status) =>
  client.put(`/qris-cross-border/${uuid}/status`, { status }).then((r) => r.data)
