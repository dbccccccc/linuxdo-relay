import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Button, Card, Table, Toast, Typography, Tag, Descriptions, Space, Progress } from '@douyinfe/semi-ui';
import { IconGift, IconRefresh, IconHistory, IconStar } from '@douyinfe/semi-icons';
import axios from 'axios';
import { useAuth } from '../auth/AuthContext.jsx';

const { Title, Text } = Typography;

// Wheel component
function SpinWheel({ options, spinning, targetIndex, onSpinEnd }) {
  const canvasRef = useRef(null);
  const [rotation, setRotation] = useState(0);
  const animationRef = useRef(null);

  const segmentAngle = options.length > 0 ? 360 / options.length : 360;

  // Draw wheel
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || options.length === 0) return;

    const ctx = canvas.getContext('2d');
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(centerX, centerY) - 10;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Draw segments
    options.forEach((opt, idx) => {
      const startAngle = (idx * segmentAngle - 90) * (Math.PI / 180);
      const endAngle = ((idx + 1) * segmentAngle - 90) * (Math.PI / 180);

      ctx.beginPath();
      ctx.moveTo(centerX, centerY);
      ctx.arc(centerX, centerY, radius, startAngle, endAngle);
      ctx.closePath();
      ctx.fillStyle = opt.color || '#FFD93D';
      ctx.fill();
      ctx.strokeStyle = '#fff';
      ctx.lineWidth = 2;
      ctx.stroke();

      // Draw text
      ctx.save();
      ctx.translate(centerX, centerY);
      ctx.rotate((startAngle + endAngle) / 2);
      ctx.textAlign = 'right';
      ctx.fillStyle = '#333';
      ctx.font = 'bold 14px sans-serif';
      ctx.fillText(opt.label, radius - 20, 5);
      ctx.restore();
    });

    // Draw center circle
    ctx.beginPath();
    ctx.arc(centerX, centerY, 30, 0, 2 * Math.PI);
    ctx.fillStyle = '#fff';
    ctx.fill();
    ctx.strokeStyle = '#ddd';
    ctx.lineWidth = 2;
    ctx.stroke();

    // Draw center text
    ctx.fillStyle = '#333';
    ctx.font = 'bold 12px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('签到', centerX, centerY + 4);
  }, [options, segmentAngle]);

  // Spin animation
  useEffect(() => {
    if (!spinning) return;

    const totalSpins = 5; // Number of full rotations
    const targetDeg = 360 * totalSpins + (360 - targetIndex * segmentAngle - segmentAngle / 2);
    const duration = 4000; // 4 seconds
    const startTime = Date.now();
    const startRotation = rotation;

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);

      // Ease out cubic
      const easeOut = 1 - Math.pow(1 - progress, 3);
      const currentRotation = startRotation + targetDeg * easeOut;

      setRotation(currentRotation);

      if (progress < 1) {
        animationRef.current = requestAnimationFrame(animate);
      } else {
        onSpinEnd?.();
      }
    };

    animationRef.current = requestAnimationFrame(animate);

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [spinning, targetIndex, segmentAngle, onSpinEnd]);

  return (
    <div style={{ position: 'relative', display: 'inline-block' }}>
      {/* Pointer */}
      <div
        style={{
          position: 'absolute',
          top: -5,
          left: '50%',
          transform: 'translateX(-50%)',
          width: 0,
          height: 0,
          borderLeft: '15px solid transparent',
          borderRight: '15px solid transparent',
          borderTop: '25px solid #e74c3c',
          zIndex: 10,
        }}
      />
      <canvas
        ref={canvasRef}
        width={300}
        height={300}
        style={{
          transform: `rotate(${rotation}deg)`,
          transition: spinning ? 'none' : 'transform 0.1s',
        }}
      />
    </div>
  );
}

export function CheckInPage() {
  const { token, user, reloadUser } = useAuth();
  const [config, setConfig] = useState(null);
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [spinning, setSpinning] = useState(false);
  const [targetIndex, setTargetIndex] = useState(0);
  const [result, setResult] = useState(null);

  const authHeaders = useMemo(
    () => (token ? { Authorization: `Bearer ${token}` } : undefined),
    [token],
  );

  const fetchConfig = useCallback(async () => {
    if (!token) return;
    try {
      const res = await axios.get('/me/check_in/config', { headers: authHeaders });
      setConfig(res.data);
    } catch (err) {
      console.error('fetch check-in config failed', err);
      Toast.error('获取签到配置失败');
    }
  }, [authHeaders, token]);

  const fetchStatus = useCallback(async () => {
    if (!token) return;
    try {
      const res = await axios.get('/me/check_in/status', { headers: authHeaders });
      setStatus(res.data);
    } catch (err) {
      console.error('fetch check-in status failed', err);
      Toast.error('获取签到状态失败');
    }
  }, [authHeaders, token]);

  const loadData = useCallback(async () => {
    setLoading(true);
    await Promise.all([fetchConfig(), fetchStatus()]);
    setLoading(false);
  }, [fetchConfig, fetchStatus]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSpin = useCallback(async () => {
    if (!token || spinning || status?.checked_in_today) return;

    setSpinning(true);
    setResult(null);

    try {
      const res = await axios.post('/me/check_in/spin', {}, { headers: authHeaders });
      const reward = res.data.reward;
      setTargetIndex(reward.wheel_index);
      setResult({
        label: reward.label,
        baseCredits: reward.base_credits,
        multiplier: reward.multiplier_percent,
        finalCredits: reward.final_credits,
        color: reward.color,
      });
      setStatus((prev) => ({
        ...prev,
        checked_in_today: true,
        today_reward: reward.final_credits,
        streak: res.data.streak,
        credits: res.data.credits,
        recent_logs: res.data.recent_logs,
      }));
      await reloadUser();
    } catch (err) {
      setSpinning(false);
      if (err?.response?.data?.error === 'already_checked_in') {
        setStatus((prev) => (prev ? { ...prev, checked_in_today: true } : prev));
        Toast.warning('今日已签到');
      } else {
        console.error('spin failed', err);
        Toast.error('签到失败');
      }
    }
  }, [authHeaders, reloadUser, spinning, status?.checked_in_today, token]);

  const handleSpinEnd = useCallback(() => {
    setSpinning(false);
    if (result) {
      Toast.success({
        content: (
          <span>
            恭喜获得 <strong>{result.label}</strong>！
            <br />
            基础 {result.baseCredits} × {result.multiplier}% = <strong>{result.finalCredits}</strong> 积分
          </span>
        ),
        duration: 5,
      });
    }
  }, [result]);

  const rewardOptions = config?.reward_options || [];
  const decayRules = config?.decay_rules || [];

  if (!user) {
    return (
      <Card>
        <Text>请先登录。</Text>
      </Card>
    );
  }

  return (
    <div style={{ maxWidth: 900, margin: '0 auto' }}>
      <Title heading={3} style={{ marginBottom: 24 }}>
        <IconGift style={{ marginRight: 8 }} />
        每日签到
      </Title>

      <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
        {/* Left: Wheel */}
        <Card
          style={{ flex: '1 1 340px', minWidth: 340, textAlign: 'center' }}
          loading={loading}
        >
          <div style={{ marginBottom: 16 }}>
            {rewardOptions.length > 0 ? (
              <SpinWheel
                options={rewardOptions}
                spinning={spinning}
                targetIndex={targetIndex}
                onSpinEnd={handleSpinEnd}
              />
            ) : (
              <div
                style={{
                  width: 300,
                  height: 300,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: '#f5f5f5',
                  borderRadius: '50%',
                  margin: '0 auto',
                }}
              >
                <Text type="tertiary">暂无奖励配置</Text>
              </div>
            )}
          </div>

          <Button
            type="primary"
            theme="solid"
            size="large"
            loading={spinning}
            disabled={status?.checked_in_today || rewardOptions.length === 0}
            onClick={handleSpin}
            style={{ width: '100%', height: 50, fontSize: 18 }}
          >
            {status?.checked_in_today ? '今日已签到' : spinning ? '抽奖中...' : '开始签到'}
          </Button>

          {result && !spinning && (
            <div
              style={{
                marginTop: 16,
                padding: 16,
                background: result.color || '#FFD93D',
                borderRadius: 8,
              }}
            >
              <Title heading={4} style={{ color: '#333', margin: 0 }}>
                🎉 {result.label}
              </Title>
              <Text style={{ color: '#333' }}>
                {result.baseCredits} × {result.multiplier}% ={' '}
                <strong>{result.finalCredits}</strong> 积分
              </Text>
            </div>
          )}
        </Card>

        {/* Right: Stats & Info */}
        <div style={{ flex: '1 1 300px', minWidth: 300 }}>
          <Card
            title={
              <Space>
                <IconStar /> 签到统计
              </Space>
            }
            headerExtraContent={
              <Button
                icon={<IconRefresh />}
                size="small"
                theme="borderless"
                onClick={loadData}
                loading={loading}
              />
            }
            style={{ marginBottom: 16 }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-around', padding: '16px 0' }}>
              <div style={{ textAlign: 'center' }}>
                <Text type="secondary">今日积分</Text>
                <div
                  style={{
                    fontSize: 28,
                    fontWeight: 'bold',
                    color: 'var(--semi-color-warning)',
                  }}
                >
                  {status?.today_reward ?? '-'}
                </div>
              </div>
              <div style={{ textAlign: 'center' }}>
                <Text type="secondary">连续天数</Text>
                <div
                  style={{
                    fontSize: 28,
                    fontWeight: 'bold',
                    color: 'var(--semi-color-primary)',
                  }}
                >
                  {status?.streak ?? 0}
                </div>
              </div>
              <div style={{ textAlign: 'center' }}>
                <Text type="secondary">当前积分</Text>
                <div
                  style={{
                    fontSize: 28,
                    fontWeight: 'bold',
                    color: 'var(--semi-color-success)',
                  }}
                >
                  {status?.credits ?? user?.credits ?? 0}
                </div>
              </div>
            </div>

            <Descriptions
              align="left"
              data={[
                {
                  key: '当前倍率',
                  value: (
                    <Tag color={config?.current_multiplier === 100 ? 'green' : 'orange'}>
                      {config?.current_multiplier ?? 100}%
                    </Tag>
                  ),
                },
              ]}
            />

            {decayRules.length > 0 && (
              <div style={{ marginTop: 12 }}>
                <Text type="secondary" size="small">
                  衰减规则：
                </Text>
                <div style={{ marginTop: 4 }}>
                  {decayRules.map((rule, idx) => (
                    <Tag key={idx} color="grey" style={{ marginRight: 4, marginBottom: 4 }}>
                      ≥{rule.threshold} → {rule.multiplier_percent}%
                    </Tag>
                  ))}
                </div>
              </div>
            )}
          </Card>

          <Card
            title={
              <Space>
                <IconHistory /> 最近签到
              </Space>
            }
          >
            <Table
              rowKey={(row) => `${row.check_in_date}-${row.id}`}
              loading={loading}
              dataSource={status?.recent_logs || []}
              pagination={false}
              size="small"
              columns={[
                {
                  title: '日期',
                  dataIndex: 'check_in_date',
                  render: (v) => new Date(v).toLocaleDateString('zh-CN'),
                },
                { title: '积分', dataIndex: 'earned_credits' },
                { title: '连续', dataIndex: 'streak' },
              ]}
            />
          </Card>
        </div>
      </div>

      {/* Reward Options Legend */}
      {rewardOptions.length > 0 && (
        <Card title="奖励说明" style={{ marginTop: 24 }}>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 16 }}>
            {rewardOptions.map((opt) => (
              <div
                key={opt.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  padding: '8px 16px',
                  background: opt.color || '#FFD93D',
                  borderRadius: 8,
                }}
              >
                <div
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    background: '#333',
                    marginRight: 8,
                  }}
                />
                <Text style={{ color: '#333' }}>
                  <strong>{opt.label}</strong>: {opt.credits} 积分
                  {opt.probability && (
                    <Text type="tertiary" size="small" style={{ marginLeft: 8 }}>
                      (权重: {opt.probability})
                    </Text>
                  )}
                </Text>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}
