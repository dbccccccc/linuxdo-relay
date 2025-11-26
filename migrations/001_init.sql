-- =============================================================================
-- LinuxDo Relay - 完整数据库初始化脚本
-- 合并自: 001_init ~ 007_drop_legacy_check_in_config
-- 生成日期: 2025-11-26
-- =============================================================================

-- =============================================================================
-- 1. 核心表
-- =============================================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    linuxdo_user_id BIGINT NOT NULL UNIQUE,
    linuxdo_username VARCHAR(255) NOT NULL,
    role VARCHAR(16) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(16) NOT NULL DEFAULT 'normal',
    credits INTEGER NOT NULL DEFAULT 0,
    api_key_hash TEXT,
    api_key_created_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 渠道表
CREATE TABLE IF NOT EXISTS channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    base_url TEXT NOT NULL,
    api_key TEXT NOT NULL,
    models JSONB NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'enabled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status);

-- 配额规则表
CREATE TABLE IF NOT EXISTS quota_rules (
    id SERIAL PRIMARY KEY,
    level INTEGER NOT NULL,
    model_pattern VARCHAR(64) NOT NULL,
    max_requests INTEGER NOT NULL,
    window_seconds INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quota_rules_level ON quota_rules(level);

-- =============================================================================
-- 2. 日志表
-- =============================================================================

-- API 调用日志
CREATE TABLE IF NOT EXISTS api_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    ip_address VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_logs_user_created_at ON api_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_created_at ON api_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_api_logs_status ON api_logs(status);

-- 操作日志
CREATE TABLE IF NOT EXISTS operation_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_type VARCHAR(64) NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operation_logs_user_created_at ON operation_logs(user_id, created_at);

-- 登录日志
CREATE TABLE IF NOT EXISTS login_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(64),
    user_agent VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_logs_user_created_at ON login_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_login_logs_created_at ON login_logs(created_at);

-- =============================================================================
-- 3. 积分系统
-- =============================================================================

-- 模型积分消耗规则
CREATE TABLE IF NOT EXISTS model_credit_rules (
    id SERIAL PRIMARY KEY,
    model_pattern VARCHAR(128) NOT NULL,
    credit_cost INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_model_credit_rules_pattern ON model_credit_rules(model_pattern);

-- 积分交易流水
CREATE TABLE IF NOT EXISTS credit_transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delta INTEGER NOT NULL,
    reason VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL,
    model_name VARCHAR(128),
    request_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_id ON credit_transactions(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_credit_transactions_request ON credit_transactions(request_id);

-- =============================================================================
-- 4. 签到系统 (转盘抽奖)
-- =============================================================================

-- 签到日志
CREATE TABLE IF NOT EXISTS check_in_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    check_in_date DATE NOT NULL,
    earned_credits INTEGER NOT NULL,
    streak INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, check_in_date)
);

CREATE INDEX IF NOT EXISTS idx_check_in_logs_user ON check_in_logs(user_id);

-- 转盘奖励选项
CREATE TABLE IF NOT EXISTS check_in_reward_options (
    id SERIAL PRIMARY KEY,
    label VARCHAR(64) NOT NULL,
    credits INTEGER NOT NULL,
    probability INTEGER NOT NULL,
    color VARCHAR(24) DEFAULT '#FFD93D',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 积分衰减规则
CREATE TABLE IF NOT EXISTS check_in_decay_rules (
    id SERIAL PRIMARY KEY,
    threshold INTEGER NOT NULL,
    multiplier_percent INTEGER NOT NULL CHECK (multiplier_percent > 0 AND multiplier_percent <= 100),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- 5. 默认数据
-- =============================================================================

-- 默认转盘奖励选项
INSERT INTO check_in_reward_options (label, credits, probability, color, sort_order)
VALUES
    ('小奖', 50, 600, '#FFD93D', 0),
    ('中奖', 100, 300, '#6BCB77', 1),
    ('大奖', 200, 100, '#FF6B6B', 2)
ON CONFLICT DO NOTHING;

-- 默认衰减规则
INSERT INTO check_in_decay_rules (threshold, multiplier_percent, sort_order)
VALUES
    (500, 80, 0),
    (1000, 50, 1)
ON CONFLICT DO NOTHING;

