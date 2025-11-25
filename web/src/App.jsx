import React from 'react';
import { Layout, Nav, Button, Typography, Spin } from '@douyinfe/semi-ui';
import { Routes, Route, Link } from 'react-router-dom';
import { useAuth } from './modules/auth/AuthContext.jsx';
import { LoginPage } from './modules/auth/LoginPage.jsx';
import { MePage } from './modules/me/MePage.jsx';
import { AdminChannelsPage } from './modules/admin/AdminChannelsPage.jsx';
import { AdminUsersPage } from './modules/admin/AdminUsersPage.jsx';
import { AdminQuotaRulesPage } from './modules/admin/AdminQuotaRulesPage.jsx';
import { AdminCreditRulesPage } from './modules/admin/AdminCreditRulesPage.jsx';
import { AdminLogsPage } from './modules/admin/AdminLogsPage.jsx';
import { AdminStatsPage } from './modules/admin/AdminStatsPage.jsx';
import { useSetupStatus } from './modules/setup/useSetupStatus.js';
import { SetupPage } from './modules/setup/SetupPage.jsx';

const { Header, Content, Footer } = Layout;
const { Title, Text } = Typography;

export default function App() {
  const { user, isAdmin, logout } = useAuth();
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
    <Layout style={{ minHeight: '100vh' }}>
      <Header>
        <Nav
          mode='horizontal'
          header={<Title heading={4}>LinuxDo Relay Console</Title>}
          footer={
            user ? (
              <>
                <Text style={{ marginRight: 16 }}>{user.linuxdo_username}</Text>
                <Button theme='borderless' type='primary' onClick={logout}>
                  退出登录
                </Button>
              </>
            ) : (
              <Link to='/login'>
                <Button type='primary'>LinuxDo 登录</Button>
              </Link>
            )
          }
        >
          <Nav.Item itemKey='home'>
            <Link to='/me'>我的账户</Link>
          </Nav.Item>
          {isAdmin && (
            <>
              <Nav.Item itemKey='users'>
                <Link to='/admin/users'>用户管理</Link>
              </Nav.Item>
              <Nav.Item itemKey='channels'>
                <Link to='/admin/channels'>渠道管理</Link>
              </Nav.Item>
              <Nav.Item itemKey='quota'>
                <Link to='/admin/quota_rules'>配额规则</Link>
              </Nav.Item>
              <Nav.Item itemKey='credit_rules'>
                <Link to='/admin/credit_rules'>积分规则</Link>
              </Nav.Item>
              <Nav.Item itemKey='logs'>
                <Link to='/admin/logs'>日志</Link>
              </Nav.Item>
              <Nav.Item itemKey='stats'>
                <Link to='/admin/stats'>统计</Link>
              </Nav.Item>
            </>
          )}
        </Nav>
      </Header>
      <Content style={{ padding: 24 }}>
        <Routes>
          <Route path='/login' element={<LoginPage />} />
          <Route path='/me' element={<MePage />} />
          <Route path='/admin/users' element={<AdminUsersPage />} />
          <Route path='/admin/channels' element={<AdminChannelsPage />} />
          <Route path='/admin/quota_rules' element={<AdminQuotaRulesPage />} />
          <Route path='/admin/credit_rules' element={<AdminCreditRulesPage />} />
          <Route path='/admin/logs' element={<AdminLogsPage />} />
          <Route path='/admin/stats' element={<AdminStatsPage />} />
          <Route path='*' element={<MePage />} />
        </Routes>
      </Content>
      <Footer style={{ textAlign: 'center' }}>
        LinuxDo Relay Example Console
      </Footer>
    </Layout>
  );
}

