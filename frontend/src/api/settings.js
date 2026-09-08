import client from './client'

export const getSettings = () => client.get('/settings').then((r) => r.data)
// No session required — safe to call before login (e.g. on the login page)
// since the backend only returns whitelisted branding fields.
export const getPublicSettings = () => client.get('/settings/public').then((r) => r.data)
export const updateSettings = (payload) => client.put('/settings', payload).then((r) => r.data)
export const uploadSetting = (formData) =>
  client.post('/settings/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)

// event_information_file/about_malaysia_file are multi-image settings —
// UploadImages appends to the existing list rather than replacing it, so
// admins build a gallery one batch at a time.
export const uploadSettingImages = (target, files) => {
  const formData = new FormData()
  formData.append('target', target)
  files.forEach((file) => formData.append('files', file))
  return client
    .post('/settings/upload-images', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    .then((r) => r.data)
}
export const removeSettingImage = (target, path) =>
  client.delete('/settings/images', { data: { target, path } }).then((r) => r.data)
