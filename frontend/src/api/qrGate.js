import client from './client'

export const getQrGates = (params) => client.get('/qr-gate', { params }).then((r) => r.data)
export const getQrGateOne = (uuid) => client.get(`/qr-gate/${uuid}`).then((r) => r.data)
export const createQrGate = (payload) => client.post('/qr-gate', payload).then((r) => r.data)
export const updateQrGate = (uuid, payload) => client.put(`/qr-gate/${uuid}`, payload).then((r) => r.data)
export const deleteQrGate = (uuid) => client.delete(`/qr-gate/${uuid}`).then((r) => r.data)
export const getQrGateScans = (uuid) => client.get(`/qr-gate/${uuid}/scans`).then((r) => r.data)
