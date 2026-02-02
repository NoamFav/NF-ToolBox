-- NF-ToolBox License Server Schema
-- Run with: psql -U postgres -d license_db < migrations/001_initial.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    plan TEXT NOT NULL DEFAULT 'monthly',
    current_period_end TIMESTAMP NOT NULL,
    max_devices INT DEFAULT 3,
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tools catalog
CREATE TABLE IF NOT EXISTS tools (
    name TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    description TEXT,
    price_monthly INT DEFAULT 0,
    price_yearly INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Initial tools
INSERT INTO tools (name, display_name, description, price_monthly, price_yearly) VALUES
    ('iskra', 'Iskra', 'Multi-repo git tooling', 800, 8000),
    ('zvezda', 'Zvezda', 'TUI git operator', 600, 6000),
    ('altair', 'Altair', 'Multi-repo operations', 800, 8000),
    ('derevo', 'Derevo', 'Syntax-aware AST parsing', 1000, 10000),
    ('puls', 'Puls', 'System monitor', 500, 5000),
    ('setka', 'Setka', 'Network monitor', 500, 5000),
    ('astrotop', 'AstroTop', 'Fancy system top', 500, 5000),
    ('briz', 'Briz', 'Auto-sort utility', 400, 4000),
    ('iris', 'Iris', 'Invisible AI assistant', 2000, 20000)
ON CONFLICT (name) DO NOTHING;

-- User entitlements (which tools they own)
CREATE TABLE IF NOT EXISTS entitlements (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tool_name TEXT REFERENCES tools(name),
    acquired_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, tool_name)
);

CREATE INDEX IF NOT EXISTS idx_entitlements_user ON entitlements(user_id);

-- Device activations
CREATE TABLE IF NOT EXISTS activations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tool_name TEXT,
    machine_fingerprint TEXT NOT NULL,
    nickname TEXT,
    platform TEXT,
    arch TEXT,
    activated_at TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    UNIQUE (user_id, machine_fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_activations_user ON activations(user_id);
CREATE INDEX IF NOT EXISTS idx_activations_fingerprint ON activations(machine_fingerprint);

-- Signing keys (only one row ever)
CREATE TABLE IF NOT EXISTS signing_keys (
    id INT PRIMARY KEY DEFAULT 1,
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CHECK (id = 1)
);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS subscriptions_updated_at ON subscriptions;
CREATE TRIGGER subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
