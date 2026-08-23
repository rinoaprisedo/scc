import { Routes, Route } from 'react-router-dom'
import AppLayout from '../components/layout/AppLayout'
import AuthLayout from '../components/layout/AuthLayout'
import ProtectedRoute from './ProtectedRoute'

import Login from '../pages/auth/Login'
import Dashboard from '../pages/dashboard/Dashboard'
import Users from '../pages/users/Users'
import Peserta from '../pages/peserta/Peserta'
import KotaAsal from '../pages/kota_asal/KotaAsal'
import Bandara from '../pages/bandara/Bandara'
import QrGate from '../pages/qr_gate/QrGate'
import QrisCrossBorder from '../pages/qris_cross_border/QrisCrossBorder'
import Menus from '../pages/menus/Menus'
import MenuSections from '../pages/menu_sections/MenuSections'
import Roles from '../pages/roles/Roles'
import RolePermissions from '../pages/roles/RolePermissions'
import Settings from '../pages/settings/Settings'
import ActivityLogs from '../pages/activity_logs/ActivityLogs'
import Forbidden from '../pages/Forbidden'
import NotFound from '../pages/NotFound'
import Maintenance from '../pages/Maintenance'

function AppRouter() {
  return (
    <Routes>
      <Route element={<AuthLayout />}>
        <Route path="/login" element={<Login />} />
      </Route>

      <Route path="/maintenance" element={<Maintenance />} />
      <Route path="/403" element={<Forbidden />} />
      <Route path="/404" element={<NotFound />} />

      <Route
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Dashboard />} />
        <Route
          path="/users"
          element={
            <ProtectedRoute menuKey="users">
              <Users />
            </ProtectedRoute>
          }
        />
        <Route
          path="/peserta"
          element={
            <ProtectedRoute menuKey="peserta">
              <Peserta />
            </ProtectedRoute>
          }
        />
        <Route
          path="/kota-asal"
          element={
            <ProtectedRoute menuKey="kota-asal">
              <KotaAsal />
            </ProtectedRoute>
          }
        />
        <Route
          path="/bandara"
          element={
            <ProtectedRoute menuKey="bandara">
              <Bandara />
            </ProtectedRoute>
          }
        />
        <Route
          path="/qr-gate"
          element={
            <ProtectedRoute menuKey="qr-gate">
              <QrGate />
            </ProtectedRoute>
          }
        />
        <Route
          path="/qris-cross-border"
          element={
            <ProtectedRoute menuKey="qris-cross-border">
              <QrisCrossBorder />
            </ProtectedRoute>
          }
        />
        <Route
          path="/menus"
          element={
            <ProtectedRoute menuKey="menus">
              <Menus />
            </ProtectedRoute>
          }
        />
        <Route
          path="/menu-sections"
          element={
            <ProtectedRoute menuKey="menu-sections">
              <MenuSections />
            </ProtectedRoute>
          }
        />
        <Route
          path="/roles"
          element={
            <ProtectedRoute menuKey="roles">
              <Roles />
            </ProtectedRoute>
          }
        />
        <Route
          path="/roles/:uuid/permissions"
          element={
            <ProtectedRoute menuKey="roles">
              <RolePermissions />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute menuKey="settings">
              <Settings />
            </ProtectedRoute>
          }
        />
        <Route
          path="/activity-logs"
          element={
            <ProtectedRoute menuKey="activity-logs">
              <ActivityLogs />
            </ProtectedRoute>
          }
        />
      </Route>

      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

export default AppRouter
