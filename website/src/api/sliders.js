import client from './client'

export const getActiveSliders = () => client.get('/sliders/active').then((r) => r.data)
