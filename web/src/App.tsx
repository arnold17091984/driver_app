import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './stores/authStore';
import { ProtectedRoute } from './components/common/ProtectedRoute';
import { AppShell } from './components/layout/AppShell';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { ConflictPage } from './pages/ConflictPage';
import { SettingsPage } from './pages/SettingsPage';
import { AuditLogPage } from './pages/AuditLogPage';
import { DriverPage } from './pages/DriverPage';

export default function App() {
  const { isAuthenticated, loadUser } = useAuthStore();

  useEffect(() => {
    loadUser();
  }, [loadUser]);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={
          isAuthenticated ? <Navigate to="/" replace /> : <LoginPage />
        } />

        <Route element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/booking" element={<Navigate to="/" replace />} />
          <Route path="/dispatches" element={<Navigate to="/" replace />} />
          <Route path="/reservations" element={<Navigate to="/" replace />} />
          <Route path="/conflicts" element={
            <ProtectedRoute roles={['admin', 'dispatcher']}>
              <ConflictPage />
            </ProtectedRoute>
          } />
          <Route path="/settings" element={
            <ProtectedRoute roles={['admin']}>
              <SettingsPage />
            </ProtectedRoute>
          } />
          <Route path="/audit-logs" element={
            <ProtectedRoute roles={['admin']}>
              <AuditLogPage />
            </ProtectedRoute>
          } />
          <Route path="/driver" element={
            <ProtectedRoute roles={['driver']}>
              <DriverPage />
            </ProtectedRoute>
          } />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
