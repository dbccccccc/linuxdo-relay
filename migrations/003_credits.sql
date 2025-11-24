-- Add credits column to users
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS credits INTEGER NOT NULL DEFAULT 0;

-- Table storing per-model credit costs
CREATE TABLE IF NOT EXISTS model_credit_rules (
    id SERIAL PRIMARY KEY,
    model_pattern VARCHAR(128) NOT NULL,
    credit_cost INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_model_credit_rules_pattern
    ON model_credit_rules(model_pattern);

-- Table for auditing credit balance changes
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

CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_id
    ON credit_transactions(user_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_credit_transactions_request
    ON credit_transactions(request_id);
