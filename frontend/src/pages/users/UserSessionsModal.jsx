import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { format } from 'date-fns'
import { LogOut } from 'lucide-react'
import Modal from '../../components/ui/Modal'
import EmptyState from '../../components/ui/EmptyState'
import Skeleton from '../../components/ui/Skeleton'
import { getUserSessions, deleteUserSession } from '../../api/users'

function UserSessionsModal({ open, onClose, user }) {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['users', user?.uuid, 'sessions'],
    queryFn: () => getUserSessions(user.uuid),
    enabled: open && !!user,
  })

  const killMutation = useMutation({
    mutationFn: (sessionUuid) => deleteUserSession(user.uuid, sessionUuid),
    onSuccess: () => {
      toast.success('Session revoked')
      queryClient.invalidateQueries({ queryKey: ['users', user?.uuid, 'sessions'] })
    },
    onError: (err) => toast.error(err.response?.data?.message || 'Failed to revoke session'),
  })

  const sessions = data?.data || []

  return (
    <Modal open={open} onClose={onClose} title={`Sessions - ${user?.name || ''}`}>
      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      ) : sessions.length === 0 ? (
        <EmptyState title="No active sessions" />
      ) : (
        <ul className="divide-y divide-surface-border">
          {sessions.map((s) => (
            <li key={s.uuid} className="flex items-center justify-between py-3">
              <div className="text-sm">
                <p className="font-medium text-text-primary">{s.ip_address}</p>
                <p className="text-text-secondary">{s.user_agent}</p>
                <p className="text-xs text-text-secondary">
                  Expires {s.expired_at ? format(new Date(s.expired_at), 'PP p') : '-'}
                </p>
              </div>
              <button
                onClick={() => killMutation.mutate(s.uuid)}
                className="rounded p-2 text-danger hover:bg-danger/10"
                title="Revoke session"
              >
                <LogOut size={16} />
              </button>
            </li>
          ))}
        </ul>
      )}
    </Modal>
  )
}

export default UserSessionsModal
