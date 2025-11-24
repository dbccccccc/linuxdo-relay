-- users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    linuxdo_user_id VARCHAR(64) NOT NULL UNIQUE,
    linuxdo_username VARCHAR(255) NOT NULL,
    role VARCHAR(16) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(16) NOT NULL DEFAULT 'normal',
    api_key_hash TEXT,
    api_key_created_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- channels table
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

-- quota rules table
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

