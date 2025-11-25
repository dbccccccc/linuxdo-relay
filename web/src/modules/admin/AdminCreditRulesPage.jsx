import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Modal, Popconfirm, Space, Table, Toast, Typography } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

export function AdminCreditRulesPage() {
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
      const res = await axios.get('/admin/model_credit_rules', { headers });
      setList(res.data || []);
    } catch (err) {
      console.error('fetch credit rules failed', err);
      Toast.error('获取积分规则失败');
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
          await axios.put(`/admin/model_credit_rules/${editing.id}`, values, { headers });
          Toast.success('更新积分规则成功');
        } else {
          await axios.post('/admin/model_credit_rules', values, { headers });
          Toast.success('创建积分规则成功');
        }
        setVisible(false);
        setEditing(null);
        fetchList();
      } catch (err) {
        console.error('save credit rule failed', err);
        const msg = err.response?.data?.error || '保存积分规则失败';
        Toast.error(msg);
      }
    },
    [editing, headers, fetchList],
  );

  const handleDelete = useCallback(
    async (row) => {
      try {
        await axios.delete(`/admin/model_credit_rules/${row.id}`, { headers });
        Toast.success('删除积分规则成功');
        fetchList();
      } catch (err) {
        console.error('delete credit rule failed', err);
        const msg = err.response?.data?.error || '删除积分规则失败';
        Toast.error(msg);
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
        模型积分价格
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
          pagination={{
            pageSize: 10,
            showTotal: true,
            showSizeChanger: true,
            pageSizeOpts: [10, 20, 50],
          }}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: '模型前缀', dataIndex: 'model_pattern' },
            { title: '积分单价', dataIndex: 'credit_cost', width: 140 },
            {
              title: '更新时间',
              dataIndex: 'updated_at',
              width: 200,
              render: (v) => new Date(v).toLocaleString('zh-CN'),
            },
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
                  <Popconfirm
                    title='确认删除'
                    content={`确定要删除模型前缀 "${row.model_pattern}" 的积分规则吗？`}
                    onConfirm={() => handleDelete(row)}
                  >
                    <Button
                      size='small'
                      theme='borderless'
                      type='danger'
                    >
                      删除
                    </Button>
                  </Popconfirm>
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
        title={editing ? '编辑积分规则' : '新建积分规则'}
      >
        <Form
          initValues={
            editing || {
              model_pattern: '',
              credit_cost: 1,
            }
          }
          onSubmit={handleSubmit}
        >
          <Form.Input
            field='model_pattern'
            label='模型前缀'
            required
            placeholder='例如 gpt-4'
          />
          <Form.InputNumber
            field='credit_cost'
            label='积分单价'
            min={1}
            precision={0}
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
