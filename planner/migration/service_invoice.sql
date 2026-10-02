-------------------------------------------------------
-- 5. Fechamento de Faturas de Serviços (Lado Planner)
-------------------------------------------------------
drop table if exists service_invoice cascade;
create table service_invoice (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    subscription_id bigint null references subscription(id) on delete set null,
    reference_month date not null,       -- Ex: '2026-10-01' (mês de competência)
    period_start date not null,          -- Ex: '2026-10-01'
    period_end date not null,            -- Ex: '2026-10-31'
    due_date date not null,              -- Vencimento calculado com base no payment_day
    payment_type int not null,           -- 1: Pré-pago, 2: Pós-pago
    total_amount numeric(15, 2) not null default 0.00,
    total_minutes int not null default 0,
    total_sessions int not null default 0,
    status int not null default 1,       -- 1: Rascunho/Em confronto, 2: Fechada/Aprovada, 3: Sincronizada com Billing, 4: Cancelada
    billing_invoice_id bigint null,      -- ID retornado pelo microservice billing após sincronização
    notes text null,
    closed_at timestamp null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

drop table if exists service_invoice_item cascade;
create table service_invoice_item (
    id bigserial primary key,
    service_invoice_id bigint not null references service_invoice(id) on delete cascade,
    offer_id bigint null references offer(id) on delete set null,
    item_type varchar(50) not null,      -- 'plano_recorrente', 'sessao_avulsa', 'sessao_excedente', 'credito_banco_horas', 'ajuste_manual'
    description varchar(255) not null,   -- Texto que será repassado para o billing (ex: "Aulas de Canto - 4 sessões (Outubro/2026)")
    -- Dados de confronto operacional
    expected_sessions int null,          -- O que estava no contrato
    performed_sessions int null,         -- O que foi executado na tabela session
    billed_sessions int not null,        -- Quantidade efetivamente faturada
    billed_minutes int not null,         -- Minutos faturados
    unit_price numeric(15, 2) not null,
    total_price numeric(15, 2) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);



