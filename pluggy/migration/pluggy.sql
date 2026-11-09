-- ==============================================================================
-- Pluggy Database Schema
-- Migration script: ./migration/pluggy.sql
-- Compatible with PostgreSQL 13+
-- ==============================================================================

-- 1. Table: pluggy_item (Instituições / Conexões do Pluggy)
CREATE TABLE IF NOT EXISTS pluggy_item (
    id VARCHAR(100) PRIMARY KEY,
    connector_id BIGINT,
    connector_name VARCHAR(150),
    connector_image_url TEXT,
    connector_primary_color VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',           -- UPDATING, UPDATED, WAITING_USER_INPUT, LOGIN_ERROR, OUTDATED, etc.
    execution_status VARCHAR(50) NOT NULL DEFAULT 'CREATED', -- SUCCESS, PARTIAL_SUCCESS, CREATED, ERROR
    client_user_id VARCHAR(100),
    error_code VARCHAR(100),
    error_message TEXT,
    last_updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_item_status ON pluggy_item(status);
CREATE INDEX IF NOT EXISTS idx_pluggy_item_client_user ON pluggy_item(client_user_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_item_last_updated ON pluggy_item(last_updated_at);

-- 2. Table: pluggy_account (Contas Bancárias e Cartões de Crédito)
CREATE TABLE IF NOT EXISTS pluggy_account (
    id VARCHAR(100) PRIMARY KEY,
    item_id VARCHAR(100) NOT NULL REFERENCES pluggy_item(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,                               -- BANK, CREDIT
    subtype VARCHAR(50) NOT NULL,                            -- CHECKING_ACCOUNT, SAVINGS_ACCOUNT, CREDIT_CARD
    name VARCHAR(200) NOT NULL,
    marketing_name VARCHAR(200),
    balance NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    currency_code VARCHAR(10) NOT NULL DEFAULT 'BRL',
    number VARCHAR(100),
    owner VARCHAR(200),
    tax_number VARCHAR(50),
    
    -- Dados de Conta Corrente / Poupança (bankData)
    bank_agency VARCHAR(50),
    bank_transfer_number VARCHAR(100),
    bank_closing_balance NUMERIC(15,2),
    bank_overdraft_contracted_limit NUMERIC(15,2),
    bank_overdraft_used_limit NUMERIC(15,2),
    bank_automatically_invested_balance NUMERIC(15,2),
    bank_has_reserved_balance BOOLEAN DEFAULT FALSE,

    -- Dados de Cartão de Crédito (creditData)
    credit_brand VARCHAR(50),                                -- VISA, MASTERCARD, AMEX, ELO, etc.
    credit_level VARCHAR(50),                                -- STANDARD, GOLD, PLATINUM, BLACK, INFINITE
    credit_limit NUMERIC(15,2),
    credit_available_limit NUMERIC(15,2),
    credit_balance_close_date DATE,
    credit_balance_due_date DATE,
    credit_minimum_payment NUMERIC(15,2),
    credit_status VARCHAR(50),                               -- ACTIVE, BLOCKED, etc.
    credit_holder_type VARCHAR(50),                          -- MAIN, ADDITIONAL

    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_account_item_id ON pluggy_account(item_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_account_type ON pluggy_account(type, subtype);
CREATE INDEX IF NOT EXISTS idx_pluggy_account_number ON pluggy_account(number);

-- 3. Table: pluggy_transaction (Transações de Extrato Bancário e Fatura de Cartão)
CREATE TABLE IF NOT EXISTS pluggy_transaction (
    id VARCHAR(100) PRIMARY KEY,
    account_id VARCHAR(100) NOT NULL REFERENCES pluggy_account(id) ON DELETE CASCADE,
    item_id VARCHAR(100) REFERENCES pluggy_item(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL,
    description_raw VARCHAR(500),
    amount NUMERIC(15,2) NOT NULL,
    currency_code VARCHAR(10) NOT NULL DEFAULT 'BRL',
    date TIMESTAMPTZ NOT NULL,
    category VARCHAR(100),
    category_id VARCHAR(50),
    type VARCHAR(20),                                        -- CREDIT, DEBIT
    status VARCHAR(20) DEFAULT 'POSTED',                     -- POSTED, PENDING
    balance NUMERIC(15,2),
    provider_code VARCHAR(100),
    merchant_name VARCHAR(200),
    merchant_cnpj VARCHAR(50),
    payment_data JSONB,
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_trans_account_id ON pluggy_transaction(account_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_trans_item_id ON pluggy_transaction(item_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_trans_date ON pluggy_transaction(date DESC);
CREATE INDEX IF NOT EXISTS idx_pluggy_trans_category ON pluggy_transaction(category);
CREATE INDEX IF NOT EXISTS idx_pluggy_trans_type ON pluggy_transaction(type);
CREATE INDEX IF NOT EXISTS idx_pluggy_trans_status ON pluggy_transaction(status);

-- 4. Table: pluggy_credit_card_bill (Faturas de Cartão de Crédito)
CREATE TABLE IF NOT EXISTS pluggy_credit_card_bill (
    id VARCHAR(100) PRIMARY KEY,
    account_id VARCHAR(100) NOT NULL REFERENCES pluggy_account(id) ON DELETE CASCADE,
    due_date DATE NOT NULL,
    bill_closing_date DATE,
    total_amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    currency_code VARCHAR(10) NOT NULL DEFAULT 'BRL',
    minimum_payment_amount NUMERIC(15,2) DEFAULT 0.00,
    allows_installments BOOLEAN DEFAULT FALSE,
    finance_charges JSONB,
    payments JSONB,
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_bill_account_id ON pluggy_credit_card_bill(account_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_bill_due_date ON pluggy_credit_card_bill(due_date DESC);

-- 5. Table: pluggy_investment (Aplicações e Investimentos Financeiros)
CREATE TABLE IF NOT EXISTS pluggy_investment (
    id VARCHAR(100) PRIMARY KEY,
    item_id VARCHAR(100) NOT NULL REFERENCES pluggy_item(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100),                                       -- CNPJ ou Código do Ativo
    isin VARCHAR(50),
    type VARCHAR(50) NOT NULL,                               -- MUTUAL_FUND, FIXED_INCOME, EQUITY, SECURITY, ETF, etc.
    subtype VARCHAR(50),
    balance NUMERIC(15,2) NOT NULL DEFAULT 0.00,             -- Saldo Líquido
    amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,              -- Valor Bruto
    currency_code VARCHAR(10) NOT NULL DEFAULT 'BRL',
    due_date DATE,                                           -- Vencimento (Renda Fixa, etc.)
    rate NUMERIC(10,4),
    rate_type VARCHAR(50),                                   -- CDI, PRE, IPCA, SELIC
    fixed_annual_rate NUMERIC(10,4),
    last_month_rate NUMERIC(10,4),
    annual_rate NUMERIC(10,4),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    institution VARCHAR(100),
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_inv_item_id ON pluggy_investment(item_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_inv_type ON pluggy_investment(type, subtype);
CREATE INDEX IF NOT EXISTS idx_pluggy_inv_code ON pluggy_investment(code);

-- 6. Table: pluggy_investment_transaction (Movimentações e Operações de Investimentos)
CREATE TABLE IF NOT EXISTS pluggy_investment_transaction (
    id VARCHAR(100) PRIMARY KEY,
    investment_id VARCHAR(100) NOT NULL REFERENCES pluggy_investment(id) ON DELETE CASCADE,
    trade_date TIMESTAMPTZ NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    type VARCHAR(50) NOT NULL,                               -- BUY, SELL, TAX, INTEREST, AMORTIZATION, TRANSFER
    movement_type VARCHAR(20),                               -- CREDIT, DEBIT
    description VARCHAR(500),
    quantity NUMERIC(15,6),
    value NUMERIC(15,6),
    amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    net_amount NUMERIC(15,2),
    agreed_rate NUMERIC(10,4),
    brokerage_number VARCHAR(100),
    expenses JSONB,
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_inv_trans_inv_id ON pluggy_investment_transaction(investment_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_inv_trans_date ON pluggy_investment_transaction(date DESC);
CREATE INDEX IF NOT EXISTS idx_pluggy_inv_trans_type ON pluggy_investment_transaction(type);

-- 7. Table: pluggy_webhook_event (Auditoria e Eventos Recebidos via Webhook)
CREATE TABLE IF NOT EXISTS pluggy_webhook_event (
    id BIGSERIAL PRIMARY KEY,
    event VARCHAR(100) NOT NULL,
    event_id VARCHAR(100),
    item_id VARCHAR(100),
    account_id VARCHAR(100),
    triggered_by VARCHAR(50),
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'RECEIVED',          -- RECEIVED, PROCESSED, ERROR
    error_message TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pluggy_wh_event ON pluggy_webhook_event(event);
CREATE INDEX IF NOT EXISTS idx_pluggy_wh_item_id ON pluggy_webhook_event(item_id);
CREATE INDEX IF NOT EXISTS idx_pluggy_wh_received ON pluggy_webhook_event(received_at DESC);

-- ==============================================================================
-- Views para Consulta Rápida e Relatórios
-- ==============================================================================

-- Visão: Extrato Bancário Consolidado (Conta Corrente / Poupança)
CREATE OR REPLACE VIEW vw_bank_statements AS
SELECT 
    t.id AS transaction_id,
    t.date AS transaction_date,
    t.description,
    t.amount,
    t.type AS transaction_type,
    t.category,
    t.status AS transaction_status,
    t.balance AS resulting_balance,
    a.id AS account_id,
    a.name AS account_name,
    a.subtype AS account_subtype,
    a.number AS account_number,
    i.id AS item_id,
    i.connector_name AS bank_name
FROM pluggy_transaction t
JOIN pluggy_account a ON t.account_id = a.id
JOIN pluggy_item i ON a.item_id = i.id
WHERE a.type = 'BANK';

-- Visão: Lançamentos de Cartão de Crédito
CREATE OR REPLACE VIEW vw_credit_card_statements AS
SELECT 
    t.id AS transaction_id,
    t.date AS transaction_date,
    t.description,
    t.amount,
    t.category,
    t.status AS transaction_status,
    a.id AS account_id,
    a.name AS card_name,
    a.credit_brand,
    a.credit_level,
    a.number AS card_masked_number,
    i.id AS item_id,
    i.connector_name AS issuer_name
FROM pluggy_transaction t
JOIN pluggy_account a ON t.account_id = a.id
JOIN pluggy_item i ON a.item_id = i.id
WHERE a.type = 'CREDIT';

-- Visão: Resumo Patrimonial Geral
CREATE OR REPLACE VIEW vw_portfolio_summary AS
SELECT
    COALESCE((SELECT SUM(balance) FROM pluggy_account WHERE type = 'BANK'), 0) AS total_bank_balance,
    COALESCE((SELECT SUM(balance) FROM pluggy_account WHERE type = 'CREDIT'), 0) AS total_credit_spent,
    COALESCE((SELECT SUM(balance) FROM pluggy_investment), 0) AS total_investments_balance,
    (SELECT COUNT(*) FROM pluggy_item) AS total_connected_items,
    (SELECT COUNT(*) FROM pluggy_account WHERE type = 'BANK') AS total_bank_accounts,
    (SELECT COUNT(*) FROM pluggy_account WHERE type = 'CREDIT') AS total_credit_cards,
    (SELECT COUNT(*) FROM pluggy_investment) AS total_investments;

