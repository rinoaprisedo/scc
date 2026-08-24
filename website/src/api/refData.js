import client from './client'

export const getKotaAsalOptions = () => client.get('/kota-asal/options').then((r) => r.data)
export const getBandaraOptions = () => client.get('/bandara/options').then((r) => r.data)
export const getBlazerSizeOptions = () => client.get('/blazer-sizes/options').then((r) => r.data)
