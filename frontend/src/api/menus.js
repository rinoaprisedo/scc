import client from './client'

export const getMenus = (params) => client.get('/menus', { params }).then((r) => r.data)
export const getMenu = (uuid) => client.get(`/menus/${uuid}`).then((r) => r.data)
export const createMenu = (payload) => client.post('/menus', payload).then((r) => r.data)
export const updateMenu = (uuid, payload) => client.put(`/menus/${uuid}`, payload).then((r) => r.data)
export const deleteMenu = (uuid) => client.delete(`/menus/${uuid}`).then((r) => r.data)
export const reorderMenus = (payload) => client.put('/menus/reorder', payload).then((r) => r.data)
