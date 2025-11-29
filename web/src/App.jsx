import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './modules/auth/AuthContext.jsx';
import { LoginPage } from './modules/auth/LoginPage.jsx';
import { MePage } from './modules/me/MePage.jsx';
import { CheckInPage } from './modules/checkin/CheckInPage.jsx';
import { AdminChannelsPage } from './modules/admin/AdminChannelsPage.jsx';
import { AdminUsersPage } from './modules/admin/AdminUsersPage.jsx';
import { AdminQuotaRulesPage } from './modules/admin/AdminQuotaRulesPage.jsx';
import { AdminCreditRulesPage } from './modules/admin/AdminCreditRulesPage.jsx';
import { AdminLogsPage } from './modules/admin/AdminLogsPage.jsx';
import { AdminStatsPage } from './modules/admin/AdminStatsPage.jsx';
import MainLayout from './layouts/MainLayout.jsx';
import { useVersionCheck } from './hooks/useVersionCheck.js';

export default function App() {
  const { user, isAdmin } = useAuth();
  
  // Check for version updates on every route change
  useVersionCheck();

  return (
    <Routes>
      <Route path='/login' element={user ? <Navigate to="/me" replace /> : <LoginPage />} />
      <Route
        path='/*'
        element={
          user ? (
            <MainLayout>
              <Routes>
                <Route path='/me' element={<MePage />} />
                <Route path='/check-in' element={<CheckInPage />} />
                <Route
                  path='/admin/users'
                  element={isAdmin ? <AdminUsersPage /> : <Navigate to="/me" replace />}
                />
                <Route
                  path='/admin/channels'
                  element={isAdmin ? <AdminChannelsPage /> : <Navigate to="/me" replace />}
                />
                <Route
                  path='/admin/quota_rules'
                  element={isAdmin ? <AdminQuotaRulesPage /> : <Navigate to="/me" replace />}
                />
                <Route
                  path='/admin/credit_rules'
                  element={isAdmin ? <AdminCreditRulesPage /> : <Navigate to="/me" replace />}
                />
                <Route
                  path='/admin/logs'
                  element={isAdmin ? <AdminLogsPage /> : <Navigate to="/me" replace />}
                />
                <Route
                  path='/admin/stats'
                  element={isAdmin ? <AdminStatsPage /> : <Navigate to="/me" replace />}
                />
                <Route path='*' element={<Navigate to="/me" replace />} />
              </Routes>
            </MainLayout>
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
    </Routes>
  );
}

