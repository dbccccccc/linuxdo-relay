import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Descriptions, Divider, Input, Typography, Table, Tabs, Space, Tag } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

export function MePage() {
  const { token, user, reloadUser } = useAuth();
  const [apiKey, setApiKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [quotaUsage, setQuotaUsage] = useState([]);
  const [quotaLoading, setQuotaLoading] = useState(false);
  const [apiLogs, setApiLogs] = useState([]);
  const [apiLogsLoading, setApiLogsLoading] = useState(false);
  const [operationLogs, setOperationLogs] = useState([]);
  const [operationLoading, setOperationLoading] = useState(false);
  const [creditTxns, setCreditTxns] = useState([]);
  const [creditLoading, setCreditLoading] = useState(false);
  const [profileLoading, setProfileLoading] = useState(false);
  const [checkInStatus, setCheckInStatus] = useState(null);
  const [checkInLoading, setCheckInLoading] = useState(false);
  const [checkInActionLoading, setCheckInActionLoading] = useState(false);

  const authHeaders = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : undefined),
    [token],
  );

  const handleRegenerate = useCallback(async () => {
    setLoading(true);
    try {
      const res = await axios.post(
        '/me/api_key/regenerate',
        {},
        { headers: authHeaders },
      );
      setApiKey(res.data.api_key || '');
    } finally {
      setLoading(false);
    }
  }, [authHeaders]);

  const fetchQuotaUsage = useCallback(async () => {
    if (!token) return;
    setQuotaLoading(true);
    try {
      const res = await axios.get('/me/quota_usage', { headers: authHeaders });
      setQuotaUsage(res.data?.items || []);
    } catch (err) {
      console.error('fetch quota usage failed', err);
    } finally {
      setQuotaLoading(false);
    }
  }, [authHeaders, token]);

  const fetchApiLogs = useCallback(async () => {
    if (!token) return;
    setApiLogsLoading(true);
    try {
      const res = await axios.get('/me/api_logs', {
        headers: authHeaders,
        params: { page: 1, page_size: 20 },
      });
      setApiLogs(res.data?.items || []);
    } catch (err) {
      console.error('fetch api logs failed', err);
    } finally {
      setApiLogsLoading(false);
    }
  }, [authHeaders, token]);

  const fetchOperationLogs = useCallback(async () => {
    if (!token) return;
    setOperationLoading(true);
    try {
      const res = await axios.get('/me/operation_logs', {
        headers: authHeaders,
        params: { page: 1, page_size: 20 },
      });
      setOperationLogs(res.data?.items || []);
    } catch (err) {
      console.error('fetch operation logs failed', err);
    } finally {
      setOperationLoading(false);
    }
  }, [authHeaders, token]);

  const fetchCreditTransactions = useCallback(async () => {
    if (!token) return;
    setCreditLoading(true);
    try {
      const res = await axios.get('/me/credit_transactions', {
        headers: authHeaders,
        params: { page: 1, page_size: 20 },
      });
      setCreditTxns(res.data?.items || []);
    } catch (err) {
      console.error('fetch credit transactions failed', err);
    } finally {
      setCreditLoading(false);
    }
  }, [authHeaders, token]);

  const fetchCheckInStatus = useCallback(async () => {
    if (!token) return;
    setCheckInLoading(true);
    try {
      const res = await axios.get('/me/check_in/status', { headers: authHeaders });
      setCheckInStatus(res.data);
    } catch (err) {
      console.error('fetch check-in status failed', err);
    } finally {
      setCheckInLoading(false);
    }
  }, [authHeaders, token]);

  const handleCheckIn = useCallback(async () => {
    if (!token) return;
    setCheckInActionLoading(true);
    try {
      const res = await axios.post('/me/check_in', {}, { headers: authHeaders });
      setCheckInStatus((prev) => ({
        ...prev,
        checked_in_today: true,
        today_reward: res.data.reward,
        streak: res.data.streak,
        credits: res.data.credits,
        recent_logs: res.data.recent_logs,
        config: res.data.config || prev?.config || null,
      }));
      await reloadUser();
      await fetchCreditTransactions();
    } catch (err) {
      if (err?.response?.data?.error === 'already_checked_in') {
        setCheckInStatus((prev) => prev ? { ...prev, checked_in_today: true } : prev);
      } else {
        console.error('check-in failed', err);
      }
    } finally {
      setCheckInActionLoading(false);
    }
  }, [authHeaders, fetchCreditTransactions, reloadUser, token]);

  const refreshProfile = useCallback(async () => {
    if (!token) return;
    setProfileLoading(true);
    try {
      await reloadUser();
      await fetchCreditTransactions();
    } catch (err) {
      console.error('refresh profile failed', err);
    } finally {
      setProfileLoading(false);
    }
  }, [fetchCreditTransactions, reloadUser, token]);

  useEffect(() => {
    if (!token) return;
    fetchQuotaUsage();
    fetchApiLogs();
    fetchOperationLogs();
    fetchCreditTransactions();
    fetchCheckInStatus();
  }, [token, fetchQuotaUsage, fetchApiLogs, fetchOperationLogs, fetchCreditTransactions, fetchCheckInStatus]);

  if (!user) {
    return (
      <Card>
        <Text>请先登录。</Text>
      </Card>
    );
  }

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <Title heading={4} style={{ marginBottom: 16 }}>
        我的账户
      </Title>
      <Card style={{ marginBottom: 24 }}>
        <Descriptions
          data={[
            { key: 'LinuxDo 用户名', value: user.linuxdo_username },
            { key: '角色', value: user.role },
            { key: '等级', value: user.level },
            { key: '状态', value: user.status },
            { key: '积分余额', value: user.credits ?? 0 },
          ]}
        />
      </Card>

      <Card>
        <Title heading={5}>API Key</Title>
        <Text type='tertiary'>
          点击下方按钮为当前账户生成新的 API Key。密钥只会在这里显示一次，
          请务必妥善保存。
        </Text>
        <Divider margin='12px 0' />
        <Button loading={loading} onClick={handleRegenerate} type='primary'>
          生成 / 重置 API Key
        </Button>
        {apiKey && (
          <div style={{ marginTop: 16 }}>
            <Text strong>新生成的 API Key：</Text>
            <Input readOnly value={apiKey} style={{ marginTop: 8 }} />
          </div>
        )}
      </Card>

      <Card style={{ marginTop: 24 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 12,
          }}
        >
          <Title heading={5} style={{ margin: 0 }}>
            每日签到
          </Title>
          <Space>
            <Button size='small' onClick={fetchCheckInStatus} loading={checkInLoading}>
              刷新
            </Button>
            <Button
              type='primary'
              loading={checkInActionLoading}
              disabled={checkInStatus?.checked_in_today}
              onClick={handleCheckIn}
            >
              {checkInStatus?.checked_in_today ? '今日已签到' : '立即签到'}
            </Button>
          </Space>
        </div>
        <Space style={{ marginBottom: 12 }} wrap>
          <Tag color={checkInStatus?.checked_in_today ? 'green' : 'blue'}>
            {checkInStatus?.checked_in_today ? '已完成' : '未签到'}
          </Tag>
          <Text>今日积分：{checkInStatus?.today_reward ?? '-'}</Text>
          <Text>连续天数：{checkInStatus?.streak ?? 0}</Text>
        </Space>
        {checkInStatus?.config && (
          <Descriptions
            data={[
              { key: '适用等级', value: checkInStatus.config.level },
              { key: '基础积分', value: checkInStatus.config.base_reward },
              { key: '衰减阈值', value: `${checkInStatus.config.decay_threshold} 积分` },
              {
                key: '最低倍率',
                value: `${checkInStatus.config.min_multiplier_percent}%`,
              },
            ]}
          />
        )}
        <Divider margin='16px 0' />
        <Table
          rowKey={(row) => `${row.check_in_date}-${row.id}`}
          loading={checkInLoading}
          dataSource={checkInStatus?.recent_logs || []}
          pagination={false}
          columns={[
            {
              title: '日期',
              dataIndex: 'check_in_date',
              render: (v) => new Date(v).toLocaleDateString('zh-CN'),
            },
            { title: '积分', dataIndex: 'earned_credits', width: 120 },
            { title: '连续天数', dataIndex: 'streak', width: 140 },
          ]}
        />
      </Card>

      <Card style={{ marginTop: 24 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 12,
          }}
        >
          <Title heading={5} style={{ margin: 0 }}>
            积分与账单
          </Title>
          <Button size='small' onClick={refreshProfile} loading={profileLoading}>
            刷新积分
          </Button>
        </div>
        <Space style={{ marginBottom: 12 }}>
          <Text>当前积分：</Text>
          <Tag color='orange' type='solid'>
            {user.credits ?? 0}
          </Tag>
        </Space>
        <Table
          rowKey='id'
          loading={creditLoading}
          dataSource={creditTxns}
          pagination={false}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            {
              title: '变动',
              dataIndex: 'delta',
              width: 100,
              render: (v) => (
                <Text type={v >= 0 ? 'success' : 'danger'}>
                  {v >= 0 ? `+${v}` : v}
                </Text>
              ),
            },
            { title: '原因', dataIndex: 'reason', width: 140 },
            { title: '状态', dataIndex: 'status', width: 100 },
            { title: '模型', dataIndex: 'model_name', width: 160 },
            {
              title: '时间',
              dataIndex: 'created_at',
              render: (v) => new Date(v).toLocaleString('zh-CN'),
            },
          ]}
        />
      </Card>

      <Card style={{ marginTop: 24 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 12,
          }}
        >
          <Title heading={5} style={{ margin: 0 }}>
            配额使用情况
          </Title>
          <Button size='small' onClick={fetchQuotaUsage} loading={quotaLoading}>
            刷新
          </Button>
        </div>
        <Table
          rowKey={(row, idx) => `${row.model_pattern || 'all'}-${idx}`}
          loading={quotaLoading}
          dataSource={quotaUsage}
          pagination={false}
          columns={[
            { title: '模型前缀', dataIndex: 'model_pattern' },
            { title: '最大请求数', dataIndex: 'max_requests', width: 140 },
            { title: '时间窗口(秒)', dataIndex: 'window_seconds', width: 160 },
            { title: '已用', dataIndex: 'used', width: 100 },
            { title: '剩余', dataIndex: 'remaining', width: 100 },
          ]}
        />
      </Card>

      <Card style={{ marginTop: 24 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 12,
          }}
        >
          <Title heading={5} style={{ margin: 0 }}>
            最近日志
          </Title>
          <Space>
            <Button size='small' onClick={fetchApiLogs} loading={apiLogsLoading}>
              刷新调用日志
            </Button>
            <Button size='small' onClick={fetchOperationLogs} loading={operationLoading}>
              刷新操作日志
            </Button>
          </Space>
        </div>
        <Tabs type='line'>
          <Tabs.TabPane tab='API 调用' itemKey='api'>
            <Table
              rowKey='id'
              loading={apiLogsLoading}
              dataSource={apiLogs}
              pagination={false}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 80 },
                { title: '模型', dataIndex: 'model', width: 180 },
                { title: '状态', dataIndex: 'status', width: 100 },
                { title: '状态码', dataIndex: 'status_code', width: 100 },
                {
                  title: '时间',
                  dataIndex: 'created_at',
                  render: (v) => new Date(v).toLocaleString('zh-CN'),
                },
              ]}
            />
          </Tabs.TabPane>
          <Tabs.TabPane tab='操作日志' itemKey='operation'>
            <Table
              rowKey='id'
              loading={operationLoading}
              dataSource={operationLogs}
              pagination={false}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 80 },
                { title: '类型', dataIndex: 'operation_type', width: 160 },
                { title: '详情', dataIndex: 'details' },
                {
                  title: '时间',
                  dataIndex: 'created_at',
                  width: 200,
                  render: (v) => new Date(v).toLocaleString('zh-CN'),
                },
              ]}
            />
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  );
}

