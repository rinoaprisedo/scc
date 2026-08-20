import axios from 'axios'

const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8500/api/v1'

const client = axios.create({
  baseURL,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

function getCookie(name) {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? match[2] : null
}

client.interceptors.request.use((config) => {
  const token = getCookie('csrf_token')
  if (token) {
    config.headers['X-CSRF-Token'] = token
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    if (status === 401 && window.location.pathname !== '/login') {
      window.location.href = '/login'
    } else if (status === 503) {
      window.location.href = '/maintenance'
    }
    return Promise.reject(error)
  },
)

export default client
