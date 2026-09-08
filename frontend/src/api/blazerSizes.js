import client from './client'

export const getBlazerSizes = (params) => client.get('/blazer-sizes', { params }).then((r) => r.data)
export const getBlazerSizeOptions = () => client.get('/blazer-sizes/options').then((r) => r.data)
export const getBlazerSizeOne = (uuid) => client.get(`/blazer-sizes/${uuid}`).then((r) => r.data)
export const createBlazerSize = (payload) => client.post('/blazer-sizes', payload).then((r) => r.data)
export const updateBlazerSize = (uuid, payload) => client.put(`/blazer-sizes/${uuid}`, payload).then((r) => r.data)
export const deleteBlazerSize = (uuid) => client.delete(`/blazer-sizes/${uuid}`).then((r) => r.data)
export const reorderBlazerSizes = (payload) => client.put('/blazer-sizes/reorder', payload).then((r) => r.data)
