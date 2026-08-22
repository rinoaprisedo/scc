import { useState } from 'react'
import Button from '../ui/Button'
import { setMyAttendance } from '../../api/peserta'

function DeclinedScreen({ onLogout, onUpdated, registrationClosed }) {
  const [dismissed, setDismissed] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleChangeToHadir = async () => {
    setError('')
    setLoading(true)
    try {
      await setMyAttendance(true)
      await onUpdated()
    } catch (err) {
      setError(err.response?.data?.message || 'Gagal menyimpan, silakan coba lagi')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 text-center shadow-2xl">
        <h2 className="text-2xl font-bold text-navy">Kamu sudah memilih tidak hadir</h2>

        {!registrationClosed && !dismissed ? (
          <div className="mt-4 rounded-lg bg-surface-bg p-4 text-left">
            <p className="text-sm font-medium text-text-primary">Apakah kamu ingin mengubahnya menjadi hadir?</p>
            {error && <p className="mt-2 text-sm font-medium text-red-600">{error}</p>}
            <div className="mt-3 flex gap-2">
              <Button className="flex-1" disabled={loading} onClick={handleChangeToHadir}>
                {loading ? 'Menyimpan...' : 'Ya, Hadir'}
              </Button>
              <Button
                variant="secondary"
                className="flex-1"
                disabled={loading}
                onClick={() => setDismissed(true)}
              >
                Tidak, Saya Tidak Hadir
              </Button>
            </div>
          </div>
        ) : (
          <p className="mt-4 text-sm text-text-secondary">Sampai jumpa di kesempatan berikutnya!</p>
        )}

        <Button variant="secondary" className="mt-7 w-full" onClick={onLogout}>
          Logout
        </Button>
      </div>
    </div>
  )
}

export default DeclinedScreen
