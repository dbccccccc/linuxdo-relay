import React, { useCallback, useEffect } from 'react';
import { Button, Card, Typography, Layout } from '@douyinfe/semi-ui';
import { IconGithubLogo } from '@douyinfe/semi-icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext.jsx';

const { Title, Text } = Typography;
const { Content, Footer } = Layout;

export function LoginPage() {
  const { saveAuth } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    function handleMessage(event) {
      try {
        const data =
          typeof event.data === 'string' ? JSON.parse(event.data) : event.data;
        if (data?.type !== 'linuxdo-login-success') return;
        saveAuth(data.token, data.user);
        // 通过客户端路由跳转，避免重新请求后端 /me API 路径
        navigate('/me', { replace: true });
      } catch {
        // ignore
      }
    }

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, [saveAuth, navigate]);

  const handleLogin = useCallback(() => {
    const w = 600;
    const h = 700;
    const left = window.screenX + (window.outerWidth - w) / 2;
    const top = window.screenY + (window.outerHeight - h) / 2;

    // 通过后端的 /auth/linuxdo/web_login 端点标记为弹窗模式
    const url = '/auth/linuxdo/web_login';
    window.open(
      url,
      'linuxdo-login',
      `width=${w},height=${h},left=${left},top=${top}`,
    );
  }, []);

  return (
    <Layout style={{ height: '100vh', backgroundColor: 'var(--semi-color-bg-0)' }}>
      <Content
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          flexDirection: 'column',
          background: 'linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)'
        }}
      >
        <Card
          style={{ width: 400, borderRadius: 16, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
          bodyStyle={{ padding: 32 }}
        >
          <div style={{ textAlign: 'center', marginBottom: 24 }}>
            <Title heading={3} style={{ marginBottom: 8 }}>
              LinuxDo Relay
            </Title>
            <Text type='secondary'>
              请登录以继续访问控制台
            </Text>
          </div>
          
          <div style={{ textAlign: 'center' }}>
            <Button 
              theme='solid' 
              type='primary' 
              size='large' 
              block 
              onClick={handleLogin}
              icon={<IconGithubLogo />} // Assuming LinuxDo is related or just using a generic icon for now
              style={{ height: 48, fontSize: 16 }}
            >
              使用 LinuxDo 账号登录
            </Button>
            <div style={{ marginTop: 16 }}>
              <Text type='tertiary' size='small'>
                点击按钮将在新窗口中打开授权页面
              </Text>
            </div>
          </div>
        </Card>
        <Footer style={{ marginTop: 24 }}>
          <Text type="tertiary">© 2025 LinuxDo Relay. All rights reserved.</Text>
        </Footer>
      </Content>
    </Layout>
  );
}

