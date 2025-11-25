import React from 'react';
import { Spin } from '@douyinfe/semi-ui';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './modules/auth/AuthContext.jsx';
import { LoginPage } from './modules/auth/LoginPage.jsx';
import { MePage } from './modules/me/MePage.jsx';
import { AdminChannelsPage } from './modules/admin/AdminChannelsPage.jsx';
import { AdminUsersPage } from './modules/admin/AdminUsersPage.jsx';
import { AdminQuotaRulesPage } from './modules/admin/AdminQuotaRulesPage.jsx';
import { AdminCreditRulesPage } from './modules/admin/AdminCreditRulesPage.jsx';
import { AdminCheckInConfigsPage } from './modules/admin/AdminCheckInConfigsPage.jsx';
import { AdminLogsPage } from './modules/admin/AdminLogsPage.jsx';
import { AdminStatsPage } from './modules/admin/AdminStatsPage.jsx';
import { useSetupStatus } from './modules/setup/useSetupStatus.js';
import { SetupPage } from './modules/setup/SetupPage.jsx';
import MainLayout from './layouts/MainLayout.jsx';

export default function App() {
  const { user } = useAuth();
  const { status: setupStatus, loading: setupLoading, refresh: refreshSetup } = useSetupStatus();

  if (setupLoading) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin tip='正在检测服务器状态...' size='large' />
      </div>
    );
  }

  if (setupStatus?.mode && setupStatus.mode !== 'ready') {
    return <SetupPage status={setupStatus} refresh={refreshSetup} />;
  }

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
                <Route path='/admin/users' element={<AdminUsersPage />} />
                <Route path='/admin/channels' element={<AdminChannelsPage />} />
                <Route path='/admin/quota_rules' element={<AdminQuotaRulesPage />} />
                <Route path='/admin/credit_rules' element={<AdminCreditRulesPage />} />
                <Route path='/admin/check_in_configs' element={<AdminCheckInConfigsPage />} />
                <Route path='/admin/logs' element={<AdminLogsPage />} />
                <Route path='/admin/stats' element={<AdminStatsPage />} />
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

