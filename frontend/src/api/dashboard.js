import client from './client'

export const getDashboardSummary = () => client.get('/dashboard/summary').then((r) => r.data)
