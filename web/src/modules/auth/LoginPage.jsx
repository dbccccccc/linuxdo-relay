import React, { useCallback, useEffect } from 'react';
import { Button, Card, Typography } from '@douyinfe/semi-ui';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext.jsx';

const { Title, Text } = Typography;

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
    <div style={{ display: 'flex', justifyContent: 'center', marginTop: 80 }}>
      <Card style={{ width: 400 }}>
        <Title heading={4} style={{ marginBottom: 16 }}>
          使用 LinuxDo 账号登录
        </Title>
        <Text type='tertiary'>
          点击下方按钮，将在新窗口中跳转至 LinuxDo 授权页面。授权完成后，
          本页面会自动更新为登录状态。
        </Text>
        <div style={{ marginTop: 24, textAlign: 'right' }}>
          <Button type='primary' onClick={handleLogin}>
            前往 LinuxDo 登录
          </Button>
        </div>
      </Card>
    </div>
  );
}

