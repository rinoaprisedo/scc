import { useState } from 'react'
import Button from '../ui/Button'
import { setMyAttendance } from '../../api/peserta'

function OptionButton({ selected, onClick, children }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full rounded-lg border px-4 py-3 text-left text-sm font-semibold transition ${
        selected ? 'border-gold bg-gold/10 text-navy' : 'border-surface-border text-text-primary hover:border-gold/50'
      }`}
    >
      {children}
    </button>
  )
}

function AttendanceModal({ userName, onUpdated }) {
  const [choice, setChoice] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async () => {
    if (choice === null) return
    setError('')
    setLoading(true)
    try {
      await setMyAttendance(choice)
      await onUpdated()
    } catch (err) {
      setError(err.response?.data?.message || 'Gagal menyimpan, silakan coba lagi')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 shadow-2xl">
        <h2 className="text-2xl font-bold text-navy">Hei {userName}</h2>
        <p className="mt-4 text-sm font-medium text-text-primary">Apakah Anda bersedia hadir di SCC 2026?</p>

        <div className="mt-4 flex flex-col gap-3">
          <OptionButton selected={choice === true} onClick={() => setChoice(true)}>
            Ya, saya akan hadir
          </OptionButton>
          <OptionButton selected={choice === false} onClick={() => setChoice(false)}>
            Tidak, saya tidak hadir
          </OptionButton>
        </div>

        {error && <p className="mt-4 text-sm font-medium text-red-600">{error}</p>}

        <Button className="mt-7 w-full" disabled={choice === null || loading} onClick={handleSubmit}>
          {loading ? 'Menyimpan...' : 'Lanjutkan'}
        </Button>
      </div>
    </div>
  )
}

export default AttendanceModal
