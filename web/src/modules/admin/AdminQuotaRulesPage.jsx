import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Modal,
  Space,
  Table,
  Typography,
} from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

export function AdminQuotaRulesPage() {
  const { token, isAdmin } = useAuth();
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [editing, setEditing] = useState(null);

  const headers = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : undefined),
    [token],
  );

  const fetchList = useCallback(async () => {
    if (!token || !isAdmin) return;
    setLoading(true);
    try {
      const res = await axios.get('/admin/quota_rules', { headers });
      setList(res.data || []);
    } catch (err) {
      console.error('fetch quota rules failed', err);
    } finally {
      setLoading(false);
    }
  }, [headers, isAdmin, token]);

  useEffect(() => {
    fetchList();
  }, [fetchList]);

  const handleSubmit = useCallback(
    async (values) => {
      try {
        if (editing) {
          await axios.put(`/admin/quota_rules/${editing.id}`, values, { headers });
        } else {
          await axios.post('/admin/quota_rules', values, { headers });
        }
        setVisible(false);
        setEditing(null);
        fetchList();
      } catch (err) {
        console.error('save quota rule failed', err);
      }
    },
    [editing, headers, fetchList],
  );

  const handleDelete = useCallback(
    async (row) => {
      try {
        await axios.delete(`/admin/quota_rules/${row.id}`, { headers });
        fetchList();
      } catch (err) {
        console.error('delete quota rule failed', err);
      }
    },
    [fetchList, headers],
  );

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
        配额规则
      </Title>
      <Card>
        <div style={{ marginBottom: 16, textAlign: 'right' }}>
          <Button
            type='primary'
            onClick={() => {
              setEditing(null);
              setVisible(true);
            }}
          >
            新建规则
          </Button>
        </div>
        <Table
          rowKey='id'
          loading={loading}
          dataSource={list}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: '用户等级', dataIndex: 'level', width: 120 },
            { title: '模型前缀', dataIndex: 'model_pattern' },
            { title: '最大请求数', dataIndex: 'max_requests', width: 160 },
            { title: '时间窗口(秒)', dataIndex: 'window_seconds', width: 180 },
            {
              title: '操作',
              width: 160,
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
                    theme='borderless'
                    type='danger'
                    onClick={() => handleDelete(row)}
                  >
                    删除
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
        title={editing ? '编辑规则' : '新建规则'}
      >
        <Form
          initValues={
            editing || {
              level: 1,
              model_pattern: '',
              max_requests: 20,
              window_seconds: 3600,
            }
          }
          onSubmit={handleSubmit}
        >
          <Form.InputNumber
            field='level'
            label='用户等级'
            min={1}
            required
            style={{ width: '100%' }}
          />
          <Form.Input
            field='model_pattern'
            label='模型前缀'
            placeholder='例如 gpt-4'
            required
          />
          <Form.InputNumber
            field='max_requests'
            label='最大请求次数'
            min={1}
            required
            style={{ width: '100%' }}
          />
          <Form.InputNumber
            field='window_seconds'
            label='时间窗口(秒)'
            min={1}
            required
            style={{ width: '100%' }}
          />
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
    </div>
  );
}
