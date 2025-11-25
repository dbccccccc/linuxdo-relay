import React, { useState } from 'react';
import { Layout, Nav, Button, Avatar, Dropdown, Typography } from '@douyinfe/semi-ui';
import {
  IconHome,
  IconUser,
  IconSetting,
  IconHistogram,
  IconFile,
  IconKey,
  IconCreditCard,
  IconServer,
  IconMenu,
  IconExit
} from '@douyinfe/semi-icons';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../modules/auth/AuthContext.jsx';

const { Header, Footer, Sider, Content } = Layout;
const { Text } = Typography;

export default function MainLayout({ children }) {
  const { user, isAdmin, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [isCollapsed, setIsCollapsed] = useState(false);

  // Map paths to menu keys
  const getSelectedKey = () => {
    const path = location.pathname;
    if (path.startsWith('/admin/users')) return 'users';
    if (path.startsWith('/admin/channels')) return 'channels';
    if (path.startsWith('/admin/quota_rules')) return 'quota';
    if (path.startsWith('/admin/credit_rules')) return 'credit_rules';
    if (path.startsWith('/admin/check_in_configs')) return 'check_in_configs';
    if (path.startsWith('/admin/logs')) return 'logs';
    if (path.startsWith('/admin/stats')) return 'stats';
    if (path === '/me') return 'home';
    return 'home';
  };

  const navItems = [
    { itemKey: 'home', text: '我的账户', icon: <IconHome />, path: '/me' },
    ...(isAdmin ? [
      { text: '管理面板', itemKey: 'admin-group', items: [
        { itemKey: 'stats', text: '数据统计', icon: <IconHistogram />, path: '/admin/stats' },
        { itemKey: 'users', text: '用户管理', icon: <IconUser />, path: '/admin/users' },
        { itemKey: 'channels', text: '渠道管理', icon: <IconServer />, path: '/admin/channels' },
        { itemKey: 'quota', text: '配额规则', icon: <IconKey />, path: '/admin/quota_rules' },
        { itemKey: 'credit_rules', text: '积分规则', icon: <IconCreditCard />, path: '/admin/credit_rules' },
        { itemKey: 'check_in_configs', text: '签到配置', icon: <IconSetting />, path: '/admin/check_in_configs' },
        { itemKey: 'logs', text: '系统日志', icon: <IconFile />, path: '/admin/logs' },
      ]}
    ] : [])
  ];

  const renderNavItems = (items) => {
    return items.map(item => {
      if (item.items) {
        return (
          <Nav.Sub itemKey={item.itemKey} text={item.text} icon={item.icon} key={item.itemKey}>
            {item.items.map(subItem => (
              <Nav.Item 
                itemKey={subItem.itemKey} 
                text={subItem.text} 
                icon={subItem.icon} 
                key={subItem.itemKey}
                onClick={() => navigate(subItem.path)}
              />
            ))}
          </Nav.Sub>
        );
      }
      return (
        <Nav.Item 
          itemKey={item.itemKey} 
          text={item.text} 
          icon={item.icon} 
          key={item.itemKey}
          onClick={() => navigate(item.path)}
        />
      );
    });
  };

  return (
    <Layout style={{ height: '100vh', display: 'flex', flexDirection: 'row' }}>
      <Sider style={{ backgroundColor: 'var(--semi-color-bg-1)' }}>
        <Nav
          defaultOpenKeys={['admin-group']}
          selectedKeys={[getSelectedKey()]}
          style={{ maxWidth: 220, height: '100%' }}
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
          {renderNavItems(navItems)}
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
