import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Modal, Popconfirm, Space, Table, Toast, Typography } from '@douyinfe/semi-ui';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

export function AdminCheckInConfigsPage() {
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
      const res = await axios.get('/admin/check_in_configs', { headers });
      setList(res.data || []);
    } catch (err) {
      console.error('fetch check-in configs failed', err);
      Toast.error('获取签到配置失败');
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
          await axios.put(`/admin/check_in_configs/${editing.id}`, values, { headers });
          Toast.success('更新签到配置成功');
        } else {
          await axios.post('/admin/check_in_configs', values, { headers });
          Toast.success('创建签到配置成功');
        }
        setVisible(false);
        setEditing(null);
        fetchList();
      } catch (err) {
        console.error('save check-in config failed', err);
        const msg = err.response?.data?.error || '保存签到配置失败';
        Toast.error(msg);
      }
    },
    [editing, headers, fetchList],
  );

  const handleDelete = useCallback(
    async (row) => {
      try {
        await axios.delete(`/admin/check_in_configs/${row.id}`, { headers });
        Toast.success('删除签到配置成功');
        fetchList();
      } catch (err) {
        console.error('delete check-in config failed', err);
        const msg = err.response?.data?.error || '删除签到配置失败';
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
    <div style={{ maxWidth: 1000, margin: '0 auto' }}>
      <Title heading={4} style={{ marginBottom: 16 }}>
        签到配置
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
            新建配置
          </Button>
        </div>
        <Table
          rowKey='id'
          loading={loading}
          dataSource={list}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: '用户等级', dataIndex: 'level', width: 100 },
            { title: '基础奖励', dataIndex: 'base_reward', width: 100 },
            { title: '衰减阈值', dataIndex: 'decay_threshold', width: 100 },
            { 
              title: '最低倍率(%)', 
              dataIndex: 'min_multiplier_percent', 
              width: 120,
              render: (v) => `${v}%`
            },
            {
              title: '更新时间',
              dataIndex: 'updated_at',
              width: 180,
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
                    content={`确定要删除等级 ${row.level} 的签到配置吗？`}
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
        title={editing ? '编辑签到配置' : '新建签到配置'}
      >
        <Form
          initValues={
            editing || {
              level: 1,
              base_reward: 100,
              decay_threshold: 1000,
              min_multiplier_percent: 10,
            }
          }
          onSubmit={handleSubmit}
        >
          <Form.InputNumber
            field='level'
            label='用户等级'
            min={1}
            precision={0}
            required
            style={{ width: '100%' }}
            extraText='签到奖励将根据用户等级匹配对应配置'
          />
          <Form.InputNumber
            field='base_reward'
            label='基础奖励'
            min={1}
            precision={0}
            required
            style={{ width: '100%' }}
            extraText='每次签到获得的基础积分数量'
          />
          <Form.InputNumber
            field='decay_threshold'
            label='衰减阈值'
            min={1}
            precision={0}
            required
            style={{ width: '100%' }}
            extraText='用户余额超过此值后，签到奖励会按比例衰减'
          />
          <Form.InputNumber
            field='min_multiplier_percent'
            label='最低倍率(%)'
            min={1}
            max={100}
            precision={0}
            required
            style={{ width: '100%' }}
            extraText='奖励衰减的最低百分比，例如10表示最低获得基础奖励的10%'
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
