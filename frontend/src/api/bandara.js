import client from './client'

export const getBandara = (params) => client.get('/bandara', { params }).then((r) => r.data)
export const getBandaraOne = (uuid) => client.get(`/bandara/${uuid}`).then((r) => r.data)
export const createBandara = (payload) => client.post('/bandara', payload).then((r) => r.data)
export const updateBandara = (uuid, payload) => client.put(`/bandara/${uuid}`, payload).then((r) => r.data)
export const deleteBandara = (uuid) => client.delete(`/bandara/${uuid}`).then((r) => r.data)
