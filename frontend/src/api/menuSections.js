import client from './client'

export const getMenuSections = (params) => client.get('/menu-sections', { params }).then((r) => r.data)
export const getMenuSection = (uuid) => client.get(`/menu-sections/${uuid}`).then((r) => r.data)
export const createMenuSection = (payload) => client.post('/menu-sections', payload).then((r) => r.data)
export const updateMenuSection = (uuid, payload) => client.put(`/menu-sections/${uuid}`, payload).then((r) => r.data)
export const deleteMenuSection = (uuid) => client.delete(`/menu-sections/${uuid}`).then((r) => r.data)
export const reorderMenuSections = (payload) => client.put('/menu-sections/reorder', payload).then((r) => r.data)
