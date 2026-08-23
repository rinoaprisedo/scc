import client from './client'

export const getPeserta = (params) => client.get('/peserta', { params }).then((r) => r.data)
// Shared loader for AsyncSelect-driven peserta pickers (qris_cross_border's
// form and list filter) — keeps the option shape in one place.
export const searchPesertaOptions = (search) =>
  getPeserta({ search, limit: 10 }).then((res) => (res.data || []).map((p) => ({ uuid: p.uuid, name: p.name, ktp_number: p.ktp_number })))
export const getPesertaOne = (uuid) => client.get(`/peserta/${uuid}`).then((r) => r.data)
export const createPeserta = (payload) => client.post('/peserta', payload).then((r) => r.data)
export const updatePeserta = (uuid, payload) => client.put(`/peserta/${uuid}`, payload).then((r) => r.data)
export const deletePeserta = (uuid) => client.delete(`/peserta/${uuid}`).then((r) => r.data)
export const updatePesertaStatus = (uuid, status) => client.put(`/peserta/${uuid}/status`, { status }).then((r) => r.data)
export const getPesertaPointHistory = (uuid) => client.get(`/peserta/${uuid}/points`).then((r) => r.data)
export const uploadPesertaKtp = (uuid, formData) =>
  client.post(`/peserta/${uuid}/ktp`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)

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

export const exportPesertaCsv = (params) =>
  client.get('/peserta/export/csv', { params, responseType: 'blob' }).then((r) => downloadFile(r.data, 'peserta.csv'))
export const exportPesertaExcel = (params) =>
  client.get('/peserta/export/excel', { params, responseType: 'blob' }).then((r) => downloadFile(r.data, 'peserta.xlsx'))
export const exportPesertaKtpZip = (params) =>
  client.get('/peserta/export/ktp-zip', { params, responseType: 'blob' }).then((r) => downloadFile(r.data, 'peserta_ktp.zip'))
export const downloadPesertaImportTemplate = () =>
  client
    .get('/peserta/import/template', { responseType: 'blob' })
    .then((r) => downloadFile(r.data, 'peserta_import_template.xlsx'))
export const validatePesertaImport = (formData) =>
  client.post('/peserta/import/validate', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const importPesertaExcel = (formData) =>
  client.post('/peserta/import', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
