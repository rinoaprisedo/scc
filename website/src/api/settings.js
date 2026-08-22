import client from './client'

export const getPublicSettings = () => client.get('/settings/public').then((r) => r.data)
