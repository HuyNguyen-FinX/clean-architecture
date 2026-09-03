CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    balance_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    overdraft_minor BIGINT NOT NULL CHECK (overdraft_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('active', 'frozen')),
    CHECK (balance_minor >= -overdraft_minor)
);
