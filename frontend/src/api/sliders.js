import client from './client'

export const getSliders = (params) => client.get('/sliders', { params }).then((r) => r.data)
export const getSliderOne = (uuid) => client.get(`/sliders/${uuid}`).then((r) => r.data)
export const createSlider = (formData) =>
  client.post('/sliders', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const updateSlider = (uuid, formData) =>
  client.put(`/sliders/${uuid}`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((r) => r.data)
export const deleteSlider = (uuid) => client.delete(`/sliders/${uuid}`).then((r) => r.data)
export const reorderSliders = (payload) => client.put('/sliders/reorder', payload).then((r) => r.data)
