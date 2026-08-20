import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import Navbar from './Navbar'

function AppLayout() {
  return (
    <div className="min-h-screen bg-surface-bg">
      <Sidebar />
      <div className="lg:pl-[240px]">
        <Navbar />
        <main className="p-4 md:p-[28px]">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export default AppLayout
