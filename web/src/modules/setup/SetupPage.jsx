import React, { useState } from 'react';
import { Banner, Button, Card, Form, Typography, Space, Tag, Steps, Divider } from '@douyinfe/semi-ui';
import { IconCheckCircleStroked, IconLink, IconServer, IconTick } from '@douyinfe/semi-icons';
import axios from 'axios';

const { Title, Text, Paragraph } = Typography;

export function SetupPage({ status, refresh }) {
  const [currentStep, setCurrentStep] = useState(0);
  const [formValues, setFormValues] = useState({
    dsn: '',
    redis_addr: '',
    redis_password: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [testing, setTesting] = useState(false);
  const [running, setRunning] = useState(false);
  const [message, setMessage] = useState(null);
  const [testResult, setTestResult] = useState(null);
  const [dbSaved, setDbSaved] = useState(false);

  // Determine initial step based on status
  React.useEffect(() => {
    if (status?.mode === 'pending' || status?.mode === 'partial') {
      setCurrentStep(1);
      setDbSaved(true);
    } else if (status?.mode === 'ready') {
      setCurrentStep(2);
      setDbSaved(true);
    }
  }, [status?.mode]);

  const onTestConnection = async () => {
    if (!formValues.dsn) {
      setMessage({ type: 'warning', text: '请填写数据库 DSN' });
      return;
    }
    setTesting(true);
    setMessage(null);
    setTestResult(null);
    try {
      const res = await axios.post('/setup/database/test', {
        dsn: formValues.dsn,
      });
      setTestResult(res.data);
      if (res.data.success) {
        setMessage({ type: 'success', text: `连接成功！数据库版本: ${res.data.version}` });
      }
    } catch (err) {
      setTestResult({ success: false, error: err.response?.data?.error || err.message });
      setMessage({ type: 'error', text: err.response?.data?.error || err.message });
    } finally {
      setTesting(false);
    }
  };

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
      setMessage({ type: 'success', text: '数据库配置已保存' });
      setDbSaved(true);
      
      // Check if migrations were applied
      if (res.data.mode === 'ready') {
        setCurrentStep(2);
      } else {
        setCurrentStep(1);
      }
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
      setCurrentStep(2);
      refresh();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || err.message });
    } finally {
      setRunning(false);
    }
  };

  const onComplete = async () => {
    setSubmitting(true);
    setMessage(null);
    try {
      await axios.post('/setup/complete');
      setMessage({ type: 'success', text: '配置完成！请重启服务器以应用更改。' });
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || err.message });
    } finally {
      setSubmitting(false);
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

  const renderStep0 = () => (
    <div>
      <Paragraph style={{ marginBottom: 16 }}>
        请配置 PostgreSQL 数据库连接信息。系统将使用此数据库存储用户、渠道和日志等数据。
      </Paragraph>
      
      <Form
        layout='vertical'
        onValueChange={(vals) => setFormValues((prev) => ({ ...prev, ...vals }))}
        initValues={formValues}
      >
        <Form.Input
          field='dsn'
          label='PostgreSQL DSN'
          placeholder='例如：postgres://user:pass@host:5432/dbname?sslmode=disable'
          rules={[{ required: true, message: '必须填写 DSN' }]}
          extraText='格式：postgres://用户名:密码@主机:端口/数据库名?sslmode=disable'
        />
        
        <Divider margin={16}>可选配置</Divider>
        
        <Form.Input 
          field='redis_addr' 
          label='Redis 地址' 
          placeholder='redis:6379 (可选，用于配额限制)'
        />
        <Form.Input 
          field='redis_password' 
          label='Redis 密码' 
          placeholder='(可选)' 
          mode='password'
        />
      </Form>

      <Space style={{ marginTop: 16, width: '100%' }} align='start'>
        <Button 
          icon={<IconLink />}
          loading={testing} 
          onClick={onTestConnection}
        >
          测试连接
        </Button>
        <Button 
          type='primary' 
          icon={<IconServer />}
          loading={submitting} 
          onClick={onSave}
          disabled={!testResult?.success}
        >
          保存并继续
        </Button>
      </Space>

      {testResult && (
        <Banner
          type={testResult.success ? 'success' : 'danger'}
          closeIcon={null}
          style={{ marginTop: 16 }}
          description={testResult.success ? `数据库连接成功` : `连接失败: ${testResult.error}`}
        />
      )}
    </div>
  );

  const renderStep1 = () => (
    <div>
      <Paragraph style={{ marginBottom: 16 }}>
        数据库已配置，现在需要运行数据库迁移来初始化表结构。
      </Paragraph>

      {status?.mode === 'pending' && (
        <>
          <Banner type='info' closeIcon={null} style={{ marginBottom: 16 }}>
            检测到 {status?.result?.pending?.length || 0} 个待执行的迁移
          </Banner>
          {renderPending()}
          <Button 
            style={{ marginTop: 16 }} 
            type='primary' 
            loading={running} 
            onClick={() => onRunMigrations(false)}
          >
            执行迁移
          </Button>
        </>
      )}

      {status?.mode === 'partial' && (
        <>
          <Banner type='warning' closeIcon={null} style={{ marginBottom: 16 }}>
            数据库中已有部分表但缺少迁移记录，可能是从旧版本升级。请确认后强制执行迁移。
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

      {status?.mode === 'unconfigured' && dbSaved && (
        <>
          <Banner type='info' closeIcon={null} style={{ marginBottom: 16 }}>
            数据库为空，需要执行初始化迁移
          </Banner>
          <Button 
            style={{ marginTop: 16 }} 
            type='primary' 
            loading={running} 
            onClick={() => onRunMigrations(false)}
          >
            执行迁移
          </Button>
        </>
      )}
    </div>
  );

  const renderStep2 = () => (
    <div>
      <Banner type='success' closeIcon={null} icon={<IconCheckCircleStroked />} style={{ marginBottom: 16 }}>
        数据库初始化已完成！
      </Banner>
      
      <Paragraph style={{ marginBottom: 16 }}>
        所有配置已完成。请确保以下环境变量已正确设置，然后重启服务器：
      </Paragraph>

      <Card style={{ marginBottom: 16, background: '#f5f5f5' }}>
        <Text code style={{ display: 'block', whiteSpace: 'pre-wrap' }}>
{`# 必需的环境变量
APP_JWT_SECRET=your-secret-key
APP_LINUXDO_CLIENT_ID=your-client-id
APP_LINUXDO_CLIENT_SECRET=your-client-secret
APP_LINUXDO_REDIRECT_URL=https://your-domain/auth/callback`}
        </Text>
      </Card>

      <Button type='primary' onClick={onComplete} loading={submitting}>
        完成配置
      </Button>
    </div>
  );

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}>
      <Card style={{ width: 600, maxWidth: '100%' }}>
        <Title heading={3} style={{ marginBottom: 8 }}>服务器初始化向导</Title>
        <Text type='tertiary' style={{ display: 'block', marginBottom: 24 }}>
          当前状态：<Tag color={status?.mode === 'ready' ? 'green' : 'orange'}>{status?.mode || 'unknown'}</Tag>
        </Text>

        <Steps current={currentStep} style={{ marginBottom: 24 }}>
          <Steps.Step title="数据库配置" description="配置 PostgreSQL 连接" />
          <Steps.Step title="数据库初始化" description="运行数据库迁移" />
          <Steps.Step title="完成" description="启动服务" />
        </Steps>

        {message && (
          <Banner
            type={message.type}
            closeIcon={null}
            style={{ marginBottom: 16 }}
          >
            {message.text}
          </Banner>
        )}

        {currentStep === 0 && renderStep0()}
        {currentStep === 1 && renderStep1()}
        {currentStep === 2 && renderStep2()}

        {currentStep > 0 && currentStep < 2 && (
          <Button 
            style={{ marginTop: 16 }} 
            type='tertiary'
            onClick={() => setCurrentStep(currentStep - 1)}
          >
            上一步
          </Button>
        )}
      </Card>
    </div>
  );
}
