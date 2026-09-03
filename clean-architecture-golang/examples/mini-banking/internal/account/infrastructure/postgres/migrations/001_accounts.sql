CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    balance_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    overdraft_limit_minor BIGINT NOT NULL CHECK (overdraft_limit_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('active', 'frozen')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (balance_minor >= -overdraft_limit_minor)
);
