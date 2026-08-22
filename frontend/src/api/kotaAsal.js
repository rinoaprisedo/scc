import client from './client'

export const getKotaAsal = (params) => client.get('/kota-asal', { params }).then((r) => r.data)
export const getKotaAsalOne = (uuid) => client.get(`/kota-asal/${uuid}`).then((r) => r.data)
export const createKotaAsal = (payload) => client.post('/kota-asal', payload).then((r) => r.data)
export const updateKotaAsal = (uuid, payload) => client.put(`/kota-asal/${uuid}`, payload).then((r) => r.data)
export const deleteKotaAsal = (uuid) => client.delete(`/kota-asal/${uuid}`).then((r) => r.data)
