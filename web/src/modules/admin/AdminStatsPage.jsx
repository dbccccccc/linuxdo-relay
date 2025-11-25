import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Typography, Row, Col, Button, Toast } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

function StatCard({ title, value }) {
  return (
    <Card
      bordered={false}
      style={{ textAlign: 'center', height: '100%' }}
    >
      <Text type='tertiary'>{title}</Text>
      <Title heading={3} style={{ marginTop: 8 }}>
        {value ?? '--'}
      </Title>
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
    <div style={{ maxWidth: 900, margin: '0 auto' }}>
      <Title heading={4} style={{ marginBottom: 16 }}>
        全局统计
      </Title>
      <Card loading={loading}>
        <Row gutter={16}>
          <Col span={8}>
            <StatCard
              title='用户总数'
              value={data?.total_users}
            />
          </Col>
          <Col span={8}>
            <StatCard
              title='累计请求数'
              value={data?.total_requests}
            />
          </Col>
          <Col span={8}>
            <StatCard
              title='24h 活跃用户'
              value={data?.active_users_24h}
            />
          </Col>
        </Row>
        <div style={{ marginTop: 24 }}>
          <Button onClick={fetchStats}>刷新</Button>
        </div>
      </Card>
    </div>
  );
}
