import Button from './ui/Button'

function AlertModal({ message, onClose }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-2xl bg-white p-8 text-center shadow-2xl">
        <p className="text-base font-medium text-navy">{message}</p>
        <Button onClick={onClose} className="mt-6 w-full">
          OK
        </Button>
      </div>
    </div>
  )
}

export default AlertModal
