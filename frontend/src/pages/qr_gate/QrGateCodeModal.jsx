import { useRef } from 'react'
import { QRCodeCanvas } from 'qrcode.react'
import Modal from '../../components/ui/Modal'
import Button from '../../components/ui/Button'

const websiteURL = (import.meta.env.VITE_WEBSITE_URL || 'https://2026scc.com').replace(/\/+$/, '')

// A URL rather than the bare code so a phone's native camera opens the
// website directly; the website's in-app scanner extracts the code back out.
const qrValue = (code) => `${websiteURL}/${encodeURIComponent(code)}`

function QrGateCodeModal({ gate, onClose }) {
  const canvasRef = useRef(null)

  const handleDownload = () => {
    const canvas = canvasRef.current?.querySelector('canvas')
    if (!canvas) return
    const link = document.createElement('a')
    link.download = `qr-gate-${gate.code}.png`
    link.href = canvas.toDataURL('image/png')
    link.click()
  }

  return (
    <Modal open={!!gate} onClose={onClose} title={gate ? `QR — ${gate.name}` : ''} size="sm">
      {gate && (
        <div className="flex flex-col items-center gap-4">
          <div ref={canvasRef} className="rounded-lg border border-surface-border bg-white p-4">
            <QRCodeCanvas value={qrValue(gate.code)} size={220} />
          </div>
          <p className="text-center text-sm text-text-secondary">
            Code: <span className="font-mono font-semibold text-text-primary">{gate.code}</span>
          </p>
          <Button onClick={handleDownload} className="w-full">
            Download PNG
          </Button>
        </div>
      )}
    </Modal>
  )
}

export default QrGateCodeModal
