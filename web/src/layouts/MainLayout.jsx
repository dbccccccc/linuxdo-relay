import React, { useMemo, useState } from 'react';
import { Layout, Nav, Button, Avatar, Dropdown, Typography, Tabs, Divider } from '@douyinfe/semi-ui';
import {
  IconHome,
  IconUser,
  IconSetting,
  IconHistogram,
  IconFile,
  IconKey,
  IconCreditCard,
  IconServer,
  IconExit
} from '@douyinfe/semi-icons';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../modules/auth/AuthContext.jsx';

const { Header, Footer, Sider, Content } = Layout;
const { Text } = Typography;

export default function MainLayout({ children }) {
  const { user, isAdmin, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [isCollapsed, setIsCollapsed] = useState(false);

  const sectionFromPath = location.pathname.startsWith('/admin') && isAdmin ? 'admin' : 'user';
  const sections = useMemo(() => (
    isAdmin
      ? [
          { key: 'user', label: '普通用户' },
          { key: 'admin', label: '管理员配置' },
        ]
      : [
          { key: 'user', label: '普通用户' },
        ]
  ), [isAdmin]);

  const navItemsBySection = {
    user: [
      { itemKey: 'home', text: '我的账户', icon: <IconHome />, path: '/me' },
    ],
    admin: [
      { itemKey: 'stats', text: '数据统计', icon: <IconHistogram />, path: '/admin/stats' },
      { itemKey: 'users', text: '用户管理', icon: <IconUser />, path: '/admin/users' },
      { itemKey: 'channels', text: '渠道管理', icon: <IconServer />, path: '/admin/channels' },
      { itemKey: 'quota', text: '配额规则', icon: <IconKey />, path: '/admin/quota_rules' },
      { itemKey: 'credit_rules', text: '积分规则', icon: <IconCreditCard />, path: '/admin/credit_rules' },
      { itemKey: 'check_in_configs', text: '签到配置', icon: <IconSetting />, path: '/admin/check_in_configs' },
      { itemKey: 'logs', text: '系统日志', icon: <IconFile />, path: '/admin/logs' },
    ],
  };

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path.startsWith('/admin/users')) return 'users';
    if (path.startsWith('/admin/channels')) return 'channels';
    if (path.startsWith('/admin/quota_rules')) return 'quota';
    if (path.startsWith('/admin/credit_rules')) return 'credit_rules';
    if (path.startsWith('/admin/check_in_configs')) return 'check_in_configs';
    if (path.startsWith('/admin/logs')) return 'logs';
    if (path.startsWith('/admin/stats')) return 'stats';
    return 'home';
  };

  const handleSectionChange = (key) => {
    if (key === sectionFromPath) return;
    const fallbackRoute = key === 'admin' ? '/admin/stats' : '/me';
    navigate(fallbackRoute);
  };

  const currentNavItems = navItemsBySection[sectionFromPath] || navItemsBySection.user;

  const renderNavItems = (items) => (
    items.map(item => (
      <Nav.Item
        itemKey={item.itemKey}
        text={item.text}
        icon={item.icon}
        key={item.itemKey}
        onClick={() => navigate(item.path)}
      />
    ))
  );

  return (
    <Layout style={{ height: '100vh', display: 'flex', flexDirection: 'row' }}>
      <Sider style={{ backgroundColor: 'var(--semi-color-bg-1)', display: 'flex', flexDirection: 'column' }}>
        {sections.length > 1 && (
          <div style={{ padding: '12px 16px 0' }}>
            <Tabs
              type='button'
              size='small'
              activeKey={sectionFromPath}
              onChange={handleSectionChange}
            >
              {sections.map((section) => (
                <Tabs.TabPane tab={section.label} itemKey={section.key} key={section.key} />
              ))}
            </Tabs>
            <Divider style={{ margin: '12px 0' }} />
          </div>
        )}
        <Nav
          selectedKeys={[getSelectedKey()]}
          style={{ maxWidth: 220, flex: 1 }}
          isCollapsed={isCollapsed}
          header={{
            logo: <IconServer style={{ fontSize: 36, color: 'var(--semi-color-primary)' }} />,
            text: 'LinuxDo Relay'
          }}
          footer={{
            collapseButton: true,
          }}
          onCollapseChange={setIsCollapsed}
        >
          {renderNavItems(currentNavItems)}
        </Nav>
      </Sider>
      <Layout style={{ flex: 1, overflow: 'hidden' }}>
        <Header style={{ backgroundColor: 'var(--semi-color-bg-1)', height: 60, padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid var(--semi-color-border)' }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {/* Breadcrumbs or Title could go here */}
          </div>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {user && (
              <Dropdown
                trigger={'click'}
                position={'bottomRight'}
                render={
                  <Dropdown.Menu>
                    <Dropdown.Item icon={<IconUser />}>个人中心</Dropdown.Item>
                    <Dropdown.Divider />
                    <Dropdown.Item icon={<IconExit />} onClick={logout} type="danger">退出登录</Dropdown.Item>
                  </Dropdown.Menu>
                }
              >
                <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                  <Avatar color="orange" size="small" style={{ marginRight: 8 }}>
                    {user.linuxdo_username?.charAt(0)?.toUpperCase()}
                  </Avatar>
                  <Text>{user.linuxdo_username}</Text>
                </div>
              </Dropdown>
            )}
          </div>
        </Header>
        <Content
          style={{
            padding: '24px',
            backgroundColor: 'var(--semi-color-bg-0)',
            overflowY: 'auto'
          }}
        >
          <div
            style={{
              borderRadius: '10px',
              minHeight: '100%',
            }}
          >
            {children}
          </div>
        </Content>
        <Footer style={{ textAlign: 'center', padding: '12px 0', color: 'var(--semi-color-text-2)' }}>
          LinuxDo Relay Console © 2025
        </Footer>
      </Layout>
    </Layout>
  );
}
