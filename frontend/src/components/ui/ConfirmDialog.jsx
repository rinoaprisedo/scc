import Modal from './Modal'
import Button from './Button'

function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title = 'Are you sure?',
  message,
  loading,
  confirmLabel = 'Delete',
  loadingLabel = 'Deleting...',
  variant = 'danger',
}) {
  return (
    <Modal open={open} onClose={onClose} title={title} size="sm">
      <p className="text-sm text-text-secondary">
        {message || 'This action cannot be undone.'}
      </p>
      <div className="mt-5 flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button variant={variant} onClick={onConfirm} disabled={loading}>
          {loading ? loadingLabel : confirmLabel}
        </Button>
      </div>
    </Modal>
  )
}

export default ConfirmDialog
