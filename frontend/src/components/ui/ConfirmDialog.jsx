import Modal from './Modal'
import Button from './Button'

function ConfirmDialog({ open, onClose, onConfirm, title = 'Are you sure?', message, loading }) {
  return (
    <Modal open={open} onClose={onClose} title={title} size="sm">
      <p className="text-sm text-text-secondary">
        {message || 'This action cannot be undone.'}
      </p>
      <div className="mt-5 flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button variant="danger" onClick={onConfirm} disabled={loading}>
          {loading ? 'Deleting...' : 'Delete'}
        </Button>
      </div>
    </Modal>
  )
}

export default ConfirmDialog
