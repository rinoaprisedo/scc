import Button from '../ui/Button'

// A dead-end screen for a deadline that has already closed off the whole
// flow (nothing left for the participant to do but log out) — distinct from
// AlertModal, which is dismissible and used when there's a normal dashboard
// to return to.
function NoticeScreen({ title, message, onLogout }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/70 px-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl bg-white p-8 text-center shadow-2xl">
        <h2 className="text-2xl font-bold text-navy">{title}</h2>
        <p className="mt-4 text-sm text-text-secondary">{message}</p>
        <Button variant="secondary" className="mt-7 w-full" onClick={onLogout}>
          Logout
        </Button>
      </div>
    </div>
  )
}

export default NoticeScreen
