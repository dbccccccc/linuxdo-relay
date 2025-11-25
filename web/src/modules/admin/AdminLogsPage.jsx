import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Tabs, Table, Toast, Typography } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;
const PAGE_SIZE = 20;

export function AdminLogsPage() {
  const { token, isAdmin } = useAuth();
  const [apiLogs, setApiLogs] = useState([]);
  const [apiTotal, setApiTotal] = useState(0);
  const [apiPage, setApiPage] = useState(1);
  const [apiLoading, setApiLoading] = useState(false);

  const [loginLogs, setLoginLogs] = useState([]);
  const [loginTotal, setLoginTotal] = useState(0);
  const [loginPage, setLoginPage] = useState(1);
  const [loginLoading, setLoginLoading] = useState(false);

  const [creditLogs, setCreditLogs] = useState([]);
  const [creditTotal, setCreditTotal] = useState(0);
  const [creditPage, setCreditPage] = useState(1);
  const [creditLoading, setCreditLoading] = useState(false);

  const headers = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : undefined),
    [token],
  );

  const fetchApiLogs = useCallback(async () => {
    if (!token || !isAdmin) return;
    setApiLoading(true);
    try {
      const res = await axios.get('/admin/api_logs', {
        headers,
        params: { page: apiPage, page_size: PAGE_SIZE },
      });
      setApiLogs(res.data?.items || []);
      setApiTotal(res.data?.total || 0);
    } catch (err) {
      console.error('fetch api logs failed', err);
      Toast.error('获取 API 日志失败');
    } finally {
      setApiLoading(false);
    }
  }, [apiPage, headers, isAdmin, token]);

  const fetchLoginLogs = useCallback(async () => {
    if (!token || !isAdmin) return;
    setLoginLoading(true);
    try {
      const res = await axios.get('/admin/login_logs', {
        headers,
        params: { page: loginPage, page_size: PAGE_SIZE },
      });
      setLoginLogs(res.data?.items || []);
      setLoginTotal(res.data?.total || 0);
    } catch (err) {
      console.error('fetch login logs failed', err);
      Toast.error('获取登录日志失败');
    } finally {
      setLoginLoading(false);
    }
  }, [headers, isAdmin, loginPage, token]);

  const fetchCreditLogs = useCallback(async () => {
    if (!token || !isAdmin) return;
    setCreditLoading(true);
    try {
      const res = await axios.get('/admin/credit_transactions', {
        headers,
        params: { page: creditPage, page_size: PAGE_SIZE },
      });
      setCreditLogs(res.data?.items || []);
      setCreditTotal(res.data?.total || 0);
    } catch (err) {
      console.error('fetch credit transactions failed', err);
      Toast.error('获取积分流水失败');
    } finally {
      setCreditLoading(false);
    }
  }, [creditPage, headers, isAdmin, token]);

  useEffect(() => {
    fetchApiLogs();
  }, [fetchApiLogs]);

  useEffect(() => {
    fetchLoginLogs();
  }, [fetchLoginLogs]);

  useEffect(() => {
    fetchCreditLogs();
  }, [fetchCreditLogs]);

  if (!isAdmin) {
    return (
      <Card>
        <Text>仅管理员可访问此页面。</Text>
      </Card>
    );
  }

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto' }}>
      <Title heading={4} style={{ marginBottom: 16 }}>
        日志中心
      </Title>
      <Card>
        <Tabs type='line'>
          <Tabs.TabPane tab='API 调用日志' itemKey='api'>
            <Table
              rowKey='id'
              loading={apiLoading}
              dataSource={apiLogs}
              pagination={{
                currentPage: apiPage,
                pageSize: PAGE_SIZE,
                total: apiTotal,
                onPageChange: setApiPage,
              }}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 80 },
                { title: '用户 ID', dataIndex: 'user_id', width: 100 },
                { title: '模型', dataIndex: 'model', width: 180 },
                { title: '状态', dataIndex: 'status', width: 100 },
                { title: '状态码', dataIndex: 'status_code', width: 100 },
                { title: 'IP', dataIndex: 'ip_address', width: 140 },
                {
                  title: '时间',
                  dataIndex: 'created_at',
                  width: 200,
                  render: (v) => new Date(v).toLocaleString('zh-CN'),
                },
              ]}
            />
          </Tabs.TabPane>
          <Tabs.TabPane tab='登录日志' itemKey='login'>
            <Table
              rowKey='id'
              loading={loginLoading}
              dataSource={loginLogs}
              pagination={{
                currentPage: loginPage,
                pageSize: PAGE_SIZE,
                total: loginTotal,
                onPageChange: setLoginPage,
              }}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 80 },
                { title: '用户 ID', dataIndex: 'user_id', width: 100 },
                { title: 'IP', dataIndex: 'ip_address', width: 160 },
                { title: 'UA', dataIndex: 'user_agent' },
                {
                  title: '时间',
                  dataIndex: 'created_at',
                  width: 200,
                  render: (v) => new Date(v).toLocaleString('zh-CN'),
                },
              ]}
            />
          </Tabs.TabPane>
          <Tabs.TabPane tab='积分流水' itemKey='credits'>
            <Table
              rowKey='id'
              loading={creditLoading}
              dataSource={creditLogs}
              pagination={{
                currentPage: creditPage,
                pageSize: PAGE_SIZE,
                total: creditTotal,
                onPageChange: setCreditPage,
              }}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 80 },
                { title: '用户 ID', dataIndex: 'user_id', width: 100 },
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
                { title: '状态', dataIndex: 'status', width: 120 },
                { title: '模型', dataIndex: 'model_name', width: 180 },
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
