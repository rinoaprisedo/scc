import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { setPendingQrCode } from '../utils/qr'

// Landing target for a QR scanned with the phone's native camera
// (2026scc.com/<code>) — stash the code and let Dashboard redeem it once the
// participant is logged in and past onboarding.
function QrRedirect() {
  const { code } = useParams()
  const navigate = useNavigate()

  useEffect(() => {
    if (code) setPendingQrCode(code)
    navigate('/dashboard', { replace: true })
  }, [code, navigate])

  return null
}

export default QrRedirect
