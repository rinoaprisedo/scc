import client from './client'

export const pesertaLogin = (nik, password) =>
  client.post('/auth/peserta-login', { nik, password }).then((r) => r.data)
export const getMe = () => client.get('/auth/me').then((r) => r.data)
export const logout = () => client.post('/auth/logout').then((r) => r.data)
