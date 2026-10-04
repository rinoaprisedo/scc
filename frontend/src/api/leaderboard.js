import client from './client'

export const getLeaderboard = () => client.get('/leaderboard').then((r) => r.data)
export const getPublicLeaderboard = (limit) => client.get('/leaderboard/public', { params: { limit } }).then((r) => r.data)
