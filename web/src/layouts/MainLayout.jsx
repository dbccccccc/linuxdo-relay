import React, { useMemo, useState } from 'react';
import { Layout, Nav, Button, Avatar, Dropdown, Typography, Select } from '@douyinfe/semi-ui';
import {
  IconHome,
  IconUser,
  IconSetting,
  IconHistogram,
  IconFile,
  IconKey,
  IconCreditCard,
  IconServer,
  IconExit,
  IconChevronDown,
  IconDoubleChevronLeft,
  IconDoubleChevronRight,
  IconGift
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
          { value: 'user', label: '控制台' },
          { value: 'admin', label: '管理员配置' },
        ]
      : [
          { value: 'user', label: '控制台' },
        ]
  ), [isAdmin]);

  const navItemsBySection = {
    user: [
      { itemKey: 'home', text: '我的账户', icon: <IconHome />, path: '/me' },
      { itemKey: 'checkin', text: '每日签到', icon: <IconGift />, path: '/check-in' },
    ],
    admin: [
      { itemKey: 'stats', text: '数据统计', icon: <IconHistogram />, path: '/admin/stats' },
      { itemKey: 'users', text: '用户管理', icon: <IconUser />, path: '/admin/users' },
      { itemKey: 'channels', text: '渠道管理', icon: <IconServer />, path: '/admin/channels' },
      { itemKey: 'quota', text: '配额规则', icon: <IconKey />, path: '/admin/quota_rules' },
      { itemKey: 'credit_rules', text: '积分规则', icon: <IconCreditCard />, path: '/admin/credit_rules' },
      { itemKey: 'logs', text: '系统日志', icon: <IconFile />, path: '/admin/logs' },
    ],
  };

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path.startsWith('/check-in')) return 'checkin';
    if (path.startsWith('/admin/users')) return 'users';
    if (path.startsWith('/admin/channels')) return 'channels';
    if (path.startsWith('/admin/quota_rules')) return 'quota';
    if (path.startsWith('/admin/credit_rules')) return 'credit_rules';
    if (path.startsWith('/admin/logs')) return 'logs';
    if (path.startsWith('/admin/stats')) return 'stats';
    return 'home';
  };

  const handleSectionChange = (value) => {
    if (value === sectionFromPath) return;
    const fallbackRoute = value === 'admin' ? '/admin/stats' : '/me';
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

  const sectionMenus = useMemo(() => {
    const menus = [];
    
    // User section as a submenu
    menus.push({
      itemKey: 'user-section',
      text: '控制台',
      icon: <IconHome />,
      items: navItemsBySection.user,
    });

    // Admin section as a submenu (only for admins)
    if (isAdmin) {
      menus.push({
        itemKey: 'admin-section',
        text: '管理员配置',
        icon: <IconSetting />,
        items: navItemsBySection.admin,
      });
    }

    return menus;
  }, [isAdmin]);

  const renderSectionMenus = (menus) => (
    menus.map(section => (
      <Nav.Sub
        itemKey={section.itemKey}
        text={section.text}
        icon={section.icon}
        key={section.itemKey}
      >
        {section.items.map(item => (
          <Nav.Item
            itemKey={item.itemKey}
            text={item.text}
            icon={item.icon}
            key={item.itemKey}
            onClick={() => navigate(item.path)}
          />
        ))}
      </Nav.Sub>
    ))
  );

  return (
    <Layout style={{ height: '100vh', display: 'flex', flexDirection: 'row' }}>
      <Sider style={{ backgroundColor: 'var(--semi-color-bg-1)', display: 'flex', flexDirection: 'column' }}>
        <div style={{ 
          padding: '16px', 
          borderBottom: '1px solid var(--semi-color-border)',
          display: 'flex',
          alignItems: 'center',
          gap: '12px'
        }}>
          <IconServer style={{ fontSize: 32, color: 'var(--semi-color-primary)' }} />
          {!isCollapsed && (
            <Text strong style={{ fontSize: 16 }}>LinuxDo Relay</Text>
          )}
        </div>
        
        <Nav
          selectedKeys={[getSelectedKey()]}
          defaultOpenKeys={['user-section', 'admin-section']}
          style={{ flex: 1, maxWidth: 220 }}
          isCollapsed={isCollapsed}
        >
          {renderSectionMenus(sectionMenus)}
        </Nav>

        <div style={{ 
          padding: '12px 16px', 
          borderTop: '1px solid var(--semi-color-border)',
          display: 'flex',
          justifyContent: isCollapsed ? 'center' : 'flex-end'
        }}>
          <Button
            icon={isCollapsed ? <IconDoubleChevronRight /> : <IconDoubleChevronLeft />}
            theme='borderless'
            size='small'
            onClick={() => setIsCollapsed(!isCollapsed)}
            style={{ color: 'var(--semi-color-text-2)' }}
          />
        </div>
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
