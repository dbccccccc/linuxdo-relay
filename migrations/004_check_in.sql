-- Daily check-in configuration
CREATE TABLE IF NOT EXISTS check_in_configs (
    id SERIAL PRIMARY KEY,
    level INTEGER NOT NULL UNIQUE,
    base_reward INTEGER NOT NULL,                     -- 基础签到奖励积分
    decay_threshold INTEGER NOT NULL,                 -- 积分余额阈值（超过此值开始衰减）
    min_multiplier_percent INTEGER NOT NULL DEFAULT 10, -- 最低奖励倍率百分比
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Daily check-in logs per user and date (UTC+8 calendar)
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

-- 默认配置：当用户积分余额超过阈值时，签到奖励开始衰减
-- 规则：每超出阈值 100 积分，衰减 5%，最低不低于 min_multiplier_percent
INSERT INTO check_in_configs (level, base_reward, decay_threshold, min_multiplier_percent)
VALUES
    (1, 10, 1000, 50),   -- 等级1：基础10分，余额≥1000开始衰减，最低50%
    (2, 15, 2000, 60),   -- 等级2：基础15分，余额≥2000开始衰减，最低60%
    (3, 20, 5000, 70)    -- 等级3：基础20分，余额≥5000开始衰减，最低70%
ON CONFLICT (level) DO NOTHING;
