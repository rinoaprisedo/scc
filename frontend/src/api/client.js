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

// Login-free pages (the monitor leaderboard display) must keep polling
// through a 401 from App's boot-time /auth/me or a transient 503, not get
// bounced to /login or /maintenance.
const NO_REDIRECT_PATHS = ['/leaderboard/display']

client.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    if (NO_REDIRECT_PATHS.includes(window.location.pathname)) {
      return Promise.reject(error)
    }
    if (status === 401 && window.location.pathname !== '/login') {
      window.location.href = '/login'
    } else if (status === 503) {
      window.location.href = '/maintenance'
    }
    return Promise.reject(error)
  },
)

export default client
