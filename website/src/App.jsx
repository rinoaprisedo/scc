import { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import Landing from './pages/Landing.jsx'
import Dashboard from './pages/Dashboard.jsx'
import { getPublicSettings } from './api/settings.js'

function App() {
  useEffect(() => {
    // Seeds the csrf_token cookie before the user ever submits the login
    // form — the backend's double-submit-cookie CSRF check would otherwise
    // reject a cold POST that has no cookie to echo back yet.
    getPublicSettings().catch(() => {})
  }, [])

  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/dashboard" element={<Dashboard />} />
    </Routes>
  )
}

export default App
