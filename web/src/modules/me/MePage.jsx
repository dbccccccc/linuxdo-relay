import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Descriptions, Divider, Input, Typography, Table, Tabs, Space, Tag, Toast, Row, Col, Avatar } from '@douyinfe/semi-ui';
import { IconUser, IconKey, IconCreditCard, IconActivity, IconHistory, IconRefresh, IconCopy } from '@douyinfe/semi-icons';
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
      Toast.success('API Key 已重新生成');
    } catch (err) {
      console.error('regenerate api key failed', err);
      Toast.error('生成 API Key 失败');
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
      Toast.error('获取配额使用情况失败');
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
      Toast.error('获取 API 日志失败');
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
      Toast.error('获取操作日志失败');
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
      Toast.error('获取积分流水失败');
    } finally {
      setCreditLoading(false);
    }
  }, [authHeaders, token]);

  const refreshProfile = useCallback(async () => {
    if (!token) return;
    setProfileLoading(true);
    try {
      await reloadUser();
      await fetchCreditTransactions();
    } catch (err) {
      console.error('refresh profile failed', err);
      Toast.error('刷新失败');
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
  }, [token, fetchQuotaUsage, fetchApiLogs, fetchOperationLogs, fetchCreditTransactions]);

  if (!user) {
    return (
      <Card>
        <Text>请先登录。</Text>
      </Card>
    );
  }

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto' }}>
      <Title heading={3} style={{ marginBottom: 24 }}>
        我的账户
      </Title>
      
      <Row gutter={[24, 24]}>
        {/* Left Column */}
        <Col xs={24} lg={10}>
          <Card 
            title={<Space><IconUser /> 用户信息</Space>}
            headerExtraContent={
              <Button icon={<IconRefresh />} size='small' onClick={refreshProfile} loading={profileLoading} theme='borderless' />
            }
            style={{ marginBottom: 24, borderRadius: 10 }}
          >
            <div style={{ display: 'flex', alignItems: 'center', marginBottom: 24 }}>
              <Avatar color='orange' size='large' style={{ marginRight: 16 }}>
                {user.linuxdo_username?.charAt(0)?.toUpperCase()}
              </Avatar>
              <div>
                <Title heading={4}>{user.linuxdo_username}</Title>
                <Tag color='blue' style={{ marginTop: 4 }}>{user.role}</Tag>
              </div>
            </div>
            <Descriptions
              align="left"
              data={[
                { key: '等级', value: user.level },
                { key: '状态', value: <Tag color={user.status === 'active' ? 'green' : 'red'}>{user.status}</Tag> },
                { key: '积分余额', value: <Text strong type='warning' size='large'>{user.credits ?? 0}</Text> },
              ]}
            />
          </Card>

          <Card 
            title={<Space><IconKey /> API Key</Space>}
            style={{ marginBottom: 24, borderRadius: 10 }}
          >
            <Text type='tertiary'>
              点击下方按钮为当前账户生成新的 API Key。密钥只会在这里显示一次，
              请务必妥善保存。
            </Text>
            <Divider margin='12px 0' />
            <Button loading={loading} onClick={handleRegenerate} type='primary' theme='solid' block>
              生成 / 重置 API Key
            </Button>
            {apiKey && (
              <div style={{ marginTop: 16, padding: 12, background: 'var(--semi-color-fill-0)', borderRadius: 6 }}>
                <Text strong>新生成的 API Key：</Text>
                <div style={{ display: 'flex', marginTop: 8 }}>
                  <Input readOnly value={apiKey} />
                  <Button icon={<IconCopy />} style={{ marginLeft: 8 }} onClick={() => {
                    navigator.clipboard.writeText(apiKey);
                    Toast.success('已复制');
                  }} />
                </div>
              </div>
            )}
          </Card>
        </Col>

        {/* Right Column */}
        <Col xs={24} lg={14}>
          <Card 
            title={<Space><IconCreditCard /> 积分变动</Space>}
            headerExtraContent={
              <Button size='small' onClick={fetchCreditTransactions} loading={creditLoading} icon={<IconRefresh />} theme='borderless' />
            }
            style={{ marginBottom: 24, borderRadius: 10 }}
          >
            <Table
              rowKey='id'
              loading={creditLoading}
              dataSource={creditTxns}
              pagination={false}
              size="small"
              columns={[
                {
                  title: '变动',
                  dataIndex: 'delta',
                  width: 100,
                  render: (v) => (
                    <Text type={v >= 0 ? 'success' : 'danger'} strong>
                      {v >= 0 ? `+${v}` : v}
                    </Text>
                  ),
                },
                { title: '原因', dataIndex: 'reason' },
                { title: '模型', dataIndex: 'model_name' },
                {
                  title: '时间',
                  dataIndex: 'created_at',
                  render: (v) => new Date(v).toLocaleString('zh-CN'),
                },
              ]}
            />
          </Card>

          <Card 
            title={<Space><IconActivity /> 配额使用情况</Space>}
            headerExtraContent={
              <Button size='small' onClick={fetchQuotaUsage} loading={quotaLoading} icon={<IconRefresh />} theme='borderless' />
            }
            style={{ marginBottom: 24, borderRadius: 10 }}
          >
            <Table
              rowKey={(row, idx) => `${row.model_pattern || 'all'}-${idx}`}
              loading={quotaLoading}
              dataSource={quotaUsage}
              pagination={false}
              size="small"
              columns={[
                { title: '模型前缀', dataIndex: 'model_pattern' },
                { title: '限制', dataIndex: 'max_requests', render: (v) => v === -1 ? '无限制' : v },
                { title: '窗口(秒)', dataIndex: 'window_seconds' },
                { title: '已用', dataIndex: 'used' },
                { title: '剩余', dataIndex: 'remaining', render: (v) => v === -1 ? '∞' : v },
              ]}
            />
          </Card>

          <Card 
            title={<Space><IconHistory /> 系统日志</Space>}
            headerExtraContent={
              <Space>
                <Button size='small' onClick={fetchApiLogs} loading={apiLogsLoading} icon={<IconRefresh />}>API</Button>
                <Button size='small' onClick={fetchOperationLogs} loading={operationLoading} icon={<IconRefresh />}>操作</Button>
              </Space>
            }
            style={{ borderRadius: 10 }}
          >
            <Tabs type='line' size="small">
              <Tabs.TabPane tab='API 调用' itemKey='api'>
                <Table
                  rowKey='id'
                  loading={apiLogsLoading}
                  dataSource={apiLogs}
                  pagination={false}
                  size="small"
                  columns={[
                    { title: '模型', dataIndex: 'model' },
                    { title: '状态', dataIndex: 'status_code', render: (v) => <Tag color={v === 200 ? 'green' : 'red'}>{v}</Tag> },
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
                  size="small"
                  columns={[
                    { title: '类型', dataIndex: 'operation_type' },
                    { title: '详情', dataIndex: 'details', ellipsis: true },
                    {
                      title: '时间',
                      dataIndex: 'created_at',
                      render: (v) => new Date(v).toLocaleString('zh-CN'),
                    },
                  ]}
                />
              </Tabs.TabPane>
            </Tabs>
          </Card>
        </Col>
      </Row>
    </div>
  );
}

