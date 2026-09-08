import client from './client'

export const scanQr = (code) => client.post('/qr-gate/scan', { code }).then((r) => r.data)
export const getMyScans = () => client.get('/qr-gate/my-scans').then((r) => r.data)
export const getLeaderboard = (limit) => client.get('/qr-gate/leaderboard', { params: { limit } }).then((r) => r.data)
