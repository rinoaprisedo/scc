import { useEffect, useRef, useState } from 'react'
import { Html5Qrcode } from 'html5-qrcode'
import { X, Check, AlertCircle } from 'lucide-react'
import { scanQr } from '../api/qrGate'

const READER_ID = 'qr-gate-scanner-reader'

function ScannerModal({ onClose }) {
  const scannerRef = useRef(null)
  const busyRef = useRef(false)
  const [status, setStatus] = useState('scanning') // scanning | success | error
  const [message, setMessage] = useState('')

  useEffect(() => {
    const scanner = new Html5Qrcode(READER_ID)
    scannerRef.current = scanner
    let cancelled = false

    const startPromise = scanner.start(
      { facingMode: 'environment' },
      { fps: 10, qrbox: 240 },
      (decodedText) => {
        if (busyRef.current || cancelled) return
        busyRef.current = true
        try {
          scanner.pause(true)
        } catch {
          // already paused/stopped — ignore
        }
        scanQr(decodedText.trim())
          .then((res) => {
            setStatus('success')
            if (res.data.already_scanned) {
              setMessage(`Kamu sudah pernah scan QR ini — ${res.data.gate_name}`)
            } else {
              setMessage(`+${res.data.points_awarded} poin — ${res.data.gate_name}`)
            }
          })
          .catch((err) => {
            setStatus('error')
            setMessage(err.response?.data?.message || 'Gagal memproses QR')
          })
          .finally(() => {
            busyRef.current = false
          })
      },
      () => {
        // per-frame decode misses are expected while aiming — ignore
      },
    )

    startPromise.catch(() => {
      if (!cancelled) {
        setStatus('error')
        setMessage('Tidak bisa mengakses kamera. Periksa izin kamera browser kamu.')
      }
    })

    return () => {
      cancelled = true
      // React 18 StrictMode double-invokes effects in dev, so cleanup can
      // fire before start() has actually resolved — calling stop() on a
      // not-yet-running scanner throws. Chaining onto startPromise defers
      // stop()/clear() until start() has actually settled, whether that
      // happens before or after this cleanup runs; if start() rejected,
      // there's nothing to stop and the final .catch swallows it.
      startPromise
        .then(() => scanner.stop())
        .then(() => scanner.clear())
        .catch(() => {})
    }
  }, [])

  const handleScanAgain = () => {
    setStatus('scanning')
    setMessage('')
    try {
      scannerRef.current?.resume()
    } catch {
      // ignore — camera may have failed to start in the first place
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/80 p-4 backdrop-blur-sm">
      <div className="flex max-h-[90vh] w-full max-w-md flex-col overflow-hidden rounded-2xl bg-white shadow-2xl">
        <div className="flex shrink-0 items-center justify-between border-b border-surface-border px-6 py-4">
          <p className="text-base font-bold text-navy">Scanner QR</p>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="text-text-secondary transition hover:text-text-primary"
          >
            <X size={20} />
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-auto bg-surface-bg p-4">
          <div id={READER_ID} className={status === 'scanning' ? 'overflow-hidden rounded-xl' : 'hidden'} />

          {status === 'success' && (
            <div className="flex flex-col items-center gap-3 py-10 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100 text-green-600">
                <Check size={28} />
              </div>
              <p className="text-lg font-bold text-navy">Berhasil!</p>
              <p className="text-sm text-text-secondary">{message}</p>
            </div>
          )}

          {status === 'error' && (
            <div className="flex flex-col items-center gap-3 py-10 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100 text-red-600">
                <AlertCircle size={28} />
              </div>
              <p className="text-lg font-bold text-navy">Gagal</p>
              <p className="text-sm text-text-secondary">{message}</p>
            </div>
          )}
        </div>

        {status !== 'scanning' && (
          <div className="shrink-0 border-t border-surface-border bg-white px-6 py-4">
            <button
              type="button"
              onClick={handleScanAgain}
              className="w-full rounded-lg bg-gradient-to-r from-gold-light via-gold to-gold-dark py-3 text-sm font-bold uppercase tracking-wide text-navy-dark shadow-md transition hover:opacity-90"
            >
              Scan Lagi
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export default ScannerModal
