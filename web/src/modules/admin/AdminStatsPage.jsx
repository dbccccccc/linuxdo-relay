import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Typography, Row, Col, Button, Toast, Skeleton } from '@douyinfe/semi-ui';
import { IconUser, IconActivity, IconGlobe, IconRefresh } from '@douyinfe/semi-icons';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

function StatCard({ title, value, icon, color }) {
  return (
    <Card
      bordered={false}
      style={{ height: '100%', borderRadius: 10, boxShadow: '0 2px 8px rgba(0,0,0,0.05)' }}
      bodyStyle={{ display: 'flex', alignItems: 'center', padding: 24 }}
    >
      <div
        style={{
          width: 48,
          height: 48,
          borderRadius: '50%',
          backgroundColor: color || 'var(--semi-color-primary-light-default)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginRight: 16,
          color: 'white'
        }}
      >
        {icon}
      </div>
      <div>
        <Text type='secondary'>{title}</Text>
        <Title heading={3} style={{ marginTop: 4, margin: 0 }}>
          {value ?? '--'}
        </Title>
      </div>
    </Card>
  );
}

export function AdminStatsPage() {
  const { token, isAdmin } = useAuth();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);

  const headers = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : undefined),
    [token],
  );

  const fetchStats = useCallback(async () => {
    if (!token || !isAdmin) return;
    setLoading(true);
    try {
      const res = await axios.get('/admin/stats', { headers });
      setData(res.data || null);
    } catch (err) {
      console.error('fetch stats failed', err);
      Toast.error('获取统计数据失败');
    } finally {
      setLoading(false);
    }
  }, [headers, isAdmin, token]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  if (!isAdmin) {
    return (
      <Card>
        <Text>仅管理员可访问此页面。</Text>
      </Card>
    );
  }

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title heading={3} style={{ margin: 0 }}>
          全局统计
        </Title>
        <Button icon={<IconRefresh />} onClick={fetchStats} loading={loading}>刷新</Button>
      </div>
      
      <Row gutter={[24, 24]}>
        <Col span={8}>
          <Skeleton placeholder={<Skeleton.Image />} loading={loading} active>
            <StatCard
              title='用户总数'
              value={data?.total_users}
              icon={<IconUser size="large" />}
              color="#5e81f4"
            />
          </Skeleton>
        </Col>
        <Col span={8}>
          <Skeleton placeholder={<Skeleton.Image />} loading={loading} active>
            <StatCard
              title='累计请求数'
              value={data?.total_requests}
              icon={<IconGlobe size="large" />}
              color="#ff8042"
            />
          </Skeleton>
        </Col>
        <Col span={8}>
          <Skeleton placeholder={<Skeleton.Image />} loading={loading} active>
            <StatCard
              title='24h 活跃用户'
              value={data?.active_users_24h}
              icon={<IconActivity size="large" />}
              color="#00c49f"
            />
          </Skeleton>
        </Col>
      </Row>
    </div>
  );
}
