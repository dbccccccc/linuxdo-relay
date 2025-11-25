import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Modal, Space, Table, Tag, Toast, Typography, Select } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

export function AdminUsersPage() {
  const { token, isAdmin } = useAuth();
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState(null);
  const [visible, setVisible] = useState(false);
  const [creditVisible, setCreditVisible] = useState(false);
  const [creditTarget, setCreditTarget] = useState(null);
  const [creditSubmitting, setCreditSubmitting] = useState(false);

  const headers = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : undefined),
    [token],
  );

  const fetchList = useCallback(async () => {
    setLoading(true);
    try {
      const res = await axios.get('/admin/users', { headers });
      setList(res.data || []);
    } catch (err) {
      console.error('fetch users failed', err);
      Toast.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  }, [headers]);

  useEffect(() => {
    if (isAdmin && token) {
      fetchList();
    }
  }, [isAdmin, token, fetchList]);

  const handleSubmit = useCallback(
    async (values) => {
      try {
        await axios.put(`/admin/users/${editing.id}`, values, { headers });
        Toast.success('更新用户成功');
        setVisible(false);
        setEditing(null);
        fetchList();
      } catch (err) {
        console.error('update user failed', err);
        const msg = err.response?.data?.error || '更新用户失败';
        Toast.error(msg);
      }
    },
    [editing, headers, fetchList],
  );

  const handleCreditSubmit = useCallback(
    async (values) => {
      if (!creditTarget) return;
      setCreditSubmitting(true);
      try {
        await axios.post(`/admin/users/${creditTarget.id}/credits`, values, { headers });
        Toast.success('调整积分成功');
        setCreditVisible(false);
        setCreditTarget(null);
        fetchList();
      } catch (err) {
        console.error('adjust credits failed', err);
        const msg = err.response?.data?.error || '调整积分失败';
        Toast.error(msg);
      } finally {
        setCreditSubmitting(false);
      }
    },
    [creditTarget, headers, fetchList],
  );

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
        用户管理
      </Title>
      <Card>
        <Table
          rowKey='id'
          loading={loading}
          dataSource={list}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: 'LinuxDo 用户名', dataIndex: 'linuxdo_username', width: 200 },
            {
              title: '角色',
              dataIndex: 'role',
              width: 100,
              render: (v) => (
                <Tag color={v === 'admin' ? 'red' : 'blue'}>
                  {v === 'admin' ? '管理员' : '普通用户'}
                </Tag>
              ),
            },
            { title: '等级', dataIndex: 'level', width: 80 },
            {
              title: '状态',
              dataIndex: 'status',
              width: 100,
              render: (v) => (
                <Tag color={v === 'normal' ? 'green' : 'grey'}>
                  {v === 'normal' ? '正常' : '已禁用'}
                </Tag>
              ),
            },
            { title: '积分', dataIndex: 'credits', width: 100 },
            {
              title: '注册时间',
              dataIndex: 'created_at',
              width: 180,
              render: (v) => new Date(v).toLocaleString('zh-CN'),
            },
            {
              title: '操作',
              width: 220,
              render: (_, row) => (
                <Space>
                  <Button
                    size='small'
                    onClick={() => {
                      setEditing(row);
                      setVisible(true);
                    }}
                  >
                    编辑
                  </Button>
                  <Button
                    size='small'
                    theme='light'
                    onClick={() => {
                      setCreditTarget(row);
                      setCreditVisible(true);
                    }}
                  >
                    调整积分
                  </Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        visible={visible}
        onCancel={() => setVisible(false)}
        footer={null}
        title='编辑用户'
      >
        <Form
          initValues={{
            role: editing?.role || 'user',
            level: editing?.level || 1,
            status: editing?.status || 'normal',
          }}
          onSubmit={handleSubmit}
        >
          <Form.Select field='role' label='角色' style={{ width: '100%' }}>
            <Select.Option value='user'>普通用户</Select.Option>
            <Select.Option value='admin'>管理员</Select.Option>
          </Form.Select>
          <Form.InputNumber
            field='level'
            label='等级'
            min={1}
            style={{ width: '100%' }}
            placeholder='用户等级，影响配额规则'
          />
          <Form.Select field='status' label='状态' style={{ width: '100%' }}>
            <Select.Option value='normal'>正常</Select.Option>
            <Select.Option value='disabled'>已禁用</Select.Option>
          </Form.Select>
          <div style={{ textAlign: 'right', marginTop: 16 }}>
            <Space>
              <Button onClick={() => setVisible(false)}>取消</Button>
              <Button htmlType='submit' type='primary'>
                保存
              </Button>
            </Space>
          </div>
        </Form>
      </Modal>

      <Modal
        visible={creditVisible}
        onCancel={() => setCreditVisible(false)}
        footer={null}
        title={creditTarget ? `调整 ${creditTarget.linuxdo_username} 的积分` : '调整积分'}
      >
        <Form
          initValues={{ delta: 0, reason: '' }}
          onSubmit={handleCreditSubmit}
        >
          <Form.InputNumber
            field='delta'
            label='积分变动'
            required
            min={-100000}
            max={100000}
            precision={0}
            style={{ width: '100%' }}
            placeholder='正数充值，负数扣除'
          />
          <Form.Input
            field='reason'
            label='备注'
            placeholder='请输入原因，如人工调账'
          />
          <div style={{ textAlign: 'right', marginTop: 16 }}>
            <Space>
              <Button onClick={() => setCreditVisible(false)}>取消</Button>
              <Button htmlType='submit' type='primary' loading={creditSubmitting}>
                保存
              </Button>
            </Space>
          </div>
        </Form>
      </Modal>
    </div>
  );
}

