import { Outlet } from 'react-router-dom'
import useSettingsStore from '../../store/settingsStore'

function AuthLayout() {
  const appName = useSettingsStore((s) => s.appName)
  const logo = useSettingsStore((s) => s.logo)

  return (
    <div className="flex min-h-screen items-center justify-center bg-surface-bg px-4">
      <div className="w-full max-w-sm">
        <div className="mb-6 flex flex-col items-center gap-2">
          {logo ? (
            <img src={logo} alt={appName} className="h-12 w-12 object-contain" />
          ) : (
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary text-lg font-bold text-white">
              {appName.charAt(0)}
            </div>
          )}
          <h1 className="text-[22px] font-semibold tracking-[-.02em] text-text-primary">{appName}</h1>
        </div>
        <div className="rounded-xl border border-surface-border bg-surface-card p-[44px_48px] shadow-md">
          <Outlet />
        </div>
      </div>
    </div>
  )
}

export default AuthLayout
