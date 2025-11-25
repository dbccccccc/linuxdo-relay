import React, { useState } from 'react';
import { Banner, Button, Card, Form, Typography, Space, Tag } from '@douyinfe/semi-ui';
import axios from 'axios';

const { Title, Text } = Typography;

export function SetupPage({ status, refresh }) {
  const [formValues, setFormValues] = useState({
    dsn: '',
    redis_addr: '',
    redis_password: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [running, setRunning] = useState(false);
  const [message, setMessage] = useState(null);

  const onSave = async () => {
    if (!formValues.dsn) {
      setMessage({ type: 'warning', text: '请填写数据库 DSN' });
      return;
    }
    setSubmitting(true);
    setMessage(null);
    try {
      const res = await axios.post('/setup/database', {
        dsn: formValues.dsn,
        redis_addr: formValues.redis_addr,
        redis_password: formValues.redis_password,
      });
      setMessage({ type: 'success', text: '已保存数据库配置' });
      setFormValues((prev) => ({ ...prev, dsn: '', redis_password: '' }));
      refresh();
      return res.data;
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || err.message });
    } finally {
      setSubmitting(false);
    }
    return null;
  };

  const onRunMigrations = async (force = false) => {
    setRunning(true);
    setMessage(null);
    try {
      await axios.post(`/setup/migrate${force ? '?force=true' : ''}`);
      setMessage({ type: 'success', text: '迁移执行完成' });
      refresh();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || err.message });
    } finally {
      setRunning(false);
    }
  };

  const renderPending = () => {
    const pending = status?.result?.pending || [];
    if (!pending.length) return null;
    return (
      <Space wrap style={{ marginTop: 12 }}>
        {pending.map((item) => (
          <Tag key={item} color='blue'>
            {item}
          </Tag>
        ))}
      </Space>
    );
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <Card style={{ width: 520 }}>
        <Title heading={4}>服务器初始化</Title>
        <Text style={{ display: 'block', marginBottom: 16 }}>
          当前状态：<Tag color='orange'>{status?.mode || 'unknown'}</Tag>
        </Text>
        {message && (
          <Banner
            type={message.type}
            closeIcon={null}
            style={{ marginBottom: 16 }}
          >
            {message.text}
          </Banner>
        )}
        {(status?.mode === 'unconfigured' || status?.mode === 'invalid_credentials') && (
          <Form
            layout='vertical'
            onValueChange={(vals) => setFormValues((prev) => ({ ...prev, ...vals }))}
          >
            <Form.Input
              field='dsn'
              label='Postgres DSN'
              placeholder='例如：postgres://user:pass@host:5432/db?sslmode=disable'
              rules={[{ required: true, message: '必须填写 DSN' }]}
            />
            <Form.Input field='redis_addr' label='Redis 地址' placeholder='redis:6379 (可选)' />
            <Form.Input field='redis_password' label='Redis 密码' placeholder='(可选)' />
            <Button type='primary' block loading={submitting} onClick={onSave}>
              保存并运行迁移
            </Button>
          </Form>
        )}

        {status?.mode === 'pending' && (
          <>
            <Banner type='warning' closeIcon={null}>
              检测到待执行的迁移，请点击下方按钮运行。
            </Banner>
            {renderPending()}
            <Button style={{ marginTop: 16 }} type='primary' loading={running} onClick={() => onRunMigrations(false)}>
              执行迁移
            </Button>
          </>
        )}

        {status?.mode === 'partial' && (
          <>
            <Banner type='warning' closeIcon={null}>
              数据库中已有部分数据但缺少迁移记录，请确认后强制执行。
            </Banner>
            {renderPending()}
            <Button
              style={{ marginTop: 16 }}
              theme='solid'
              type='warning'
              loading={running}
              onClick={() => onRunMigrations(true)}
            >
              强制执行迁移
            </Button>
          </>
        )}

        {status?.mode === 'ready' && (
          <Banner type='success' closeIcon={null}>
            初始化已完成，请刷新页面进入正常界面。
          </Banner>
        )}
      </Card>
    </div>
  );
}
