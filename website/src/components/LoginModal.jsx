import { useState } from 'react'
import { Eye, EyeOff, X } from 'lucide-react'
import { pesertaLogin } from '../api/auth'
import Input from './ui/Input'
import Button from './ui/Button'

function LoginModal({ onClose, onLoginSuccess }) {
  const [nik, setNik] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await pesertaLogin(nik, password)
      onLoginSuccess()
    } catch (err) {
      setError(err.response?.data?.message || 'Login failed, please try again')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 backdrop-blur-sm">
      <div className="relative w-full max-w-md rounded-2xl bg-white p-8 shadow-2xl">
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="absolute right-5 top-5 text-text-secondary transition hover:text-text-primary"
        >
          <X size={20} />
        </button>

        <h2 className="text-2xl font-bold text-navy">Login To Access</h2>
        <p className="mt-1 text-sm text-text-secondary">Welcome To SCC 2026</p>

        <form onSubmit={handleSubmit} className="mt-7 flex flex-col gap-5">
          <Input
            label="NIP"
            required
            value={nik}
            onChange={(e) => setNik(e.target.value)}
            autoComplete="username"
          />

          <Input
            label="Password"
            required
            type={showPassword ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            endAdornment={
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                className="text-text-secondary hover:text-text-primary"
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            }
          />

          {error && <p className="text-sm font-medium text-red-600">{error}</p>}

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? 'Logging in...' : 'Login'}
          </Button>
        </form>
      </div>
    </div>
  )
}

export default LoginModal
