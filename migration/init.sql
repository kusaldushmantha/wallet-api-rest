CREATE SCHEMA IF NOT EXISTS wallet;

CREATE TABLE users
(
    id         UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE wallets
(
    id         UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id    UUID           NOT NULL REFERENCES users (id),
    balance    NUMERIC(20, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions
(
    id             UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    from_wallet_id UUID           NOT NULL REFERENCES wallets (id),
    to_wallet_id   UUID REFERENCES wallets (id), -- NULL unless it's a transfer
    type           VARCHAR(20)    NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'transfer')),
    amount         NUMERIC(20, 2) NOT NULL CHECK (amount > 0),
    created_at     TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- ADD SAMPLE USERS
INSERT INTO users(id)
VALUES ('0a644be3-cdf9-4491-b4ba-1cd8974c0278');
INSERT INTO users(id)
VALUES ('abe1f04a-68df-4e13-bd0d-5365ca9fdb0e');
-- ADD SAMPLE WALLETS
INSERT INTO wallets(id, user_id)
VALUES ('7dbacf5d-3099-4a66-ad3d-2fee93970017', '0a644be3-cdf9-4491-b4ba-1cd8974c0278');
INSERT INTO wallets(id, user_id)
VALUES ('2cbcd158-56d2-4d45-8113-d51adf9ef57a', 'abe1f04a-68df-4e13-bd0d-5365ca9fdb0e');

