import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Modal, Space, Table, Tag, Typography, Select } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

export function AdminChannelsPage() {
  const { token, isAdmin } = useAuth();
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState(null);
  const [visible, setVisible] = useState(false);

  const fetchList = useCallback(async () => {
    setLoading(true);
    try {
      const res = await axios.get('/admin/channels', {
        headers: { Authorization: `Bearer ${token}` },
      });
      setList(res.data || []);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (isAdmin && token) {
      fetchList();
    }
  }, [isAdmin, token, fetchList]);

  const handleSubmit = useCallback(
    async (values) => {
      const headers = { Authorization: `Bearer ${token}` };
      if (editing?.id) {
        await axios.put(`/admin/channels/${editing.id}`, values, { headers });
      } else {
        await axios.post('/admin/channels', values, { headers });
      }
      setVisible(false);
      setEditing(null);
      fetchList();
    },
    [editing, token, fetchList],
  );

  const handleDelete = useCallback(
    async (row) => {
      await axios.delete(`/admin/channels/${row.id}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      fetchList();
    },
    [token, fetchList],
  );

  if (!isAdmin) {
    return (
      <Card>
        <Text>仅管理员可访问此页面。</Text>
      </Card>
    );
  }

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto' }}>
      <Title heading={4} style={{ marginBottom: 16 }}>
        渠道管理
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
            新建渠道
          </Button>
        </div>
        <Table
          rowKey='id'
          loading={loading}
          dataSource={list}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: '名称', dataIndex: 'name' },
            { title: 'Base URL', dataIndex: 'base_url' },
            { title: '模型列表(JSON)', dataIndex: 'models' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (v) => (
                <Tag color={v === 'enabled' ? 'green' : 'grey'}>
                  {v === 'enabled' ? '启用' : '禁用'}
                </Tag>
              ),
            },
            {
              title: '操作',
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
        title={editing ? '编辑渠道' : '新建渠道'}
      >
        <Form
          initValues={
            editing || {
              name: '',
              base_url: '',
              api_key: '',
              models: '[]',
              status: 'enabled',
            }
          }
          onSubmit={handleSubmit}
        >
          <Form.Input field='name' label='名称' required />
          <Form.Input field='base_url' label='Base URL' required />
          <Form.Input field='api_key' label='上游 API Key' required />
          <Form.TextArea
            field='models'
            label='支持模型(JSON 数组)'
            rows={3}
            placeholder='["gpt-4", "gpt-3.5-turbo"]'
          />
          <Form.Select field='status' label='状态' style={{ width: '100%' }}>
            <Select.Option value='enabled'>启用</Select.Option>
            <Select.Option value='disabled'>禁用</Select.Option>
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
    </div>
  );
}

