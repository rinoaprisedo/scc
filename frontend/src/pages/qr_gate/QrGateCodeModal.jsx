import { useRef } from 'react'
import { QRCodeCanvas } from 'qrcode.react'
import Modal from '../../components/ui/Modal'
import Button from '../../components/ui/Button'

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
            <QRCodeCanvas value={gate.code} size={220} />
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
