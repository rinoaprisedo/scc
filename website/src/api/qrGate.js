import client from './client'

export const scanQr = (code) => client.post('/qr-gate/scan', { code }).then((r) => r.data)
export const getMyScans = () => client.get('/qr-gate/my-scans').then((r) => r.data)
