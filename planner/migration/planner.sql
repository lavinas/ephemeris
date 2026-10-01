-- Active: 1790015123887@@127.0.0.1@5433@planner@planner
create schema if not exists planner;

set search_path to planner;

# vendor table
drop table if exists vendor cascade;
create table vendor (
    id bigserial primary key,
    nickname varchar(150),
    legal_name varchar(150) not null,
    trading_name varchar(150),
    document varchar(50) not null,
    email VARCHAR(150) not null,
    whatsapp varchar(20) not null,
    tax_document varchar(50) not null,
    account_bank varchar(100) not null,
    account_agency varchar(20) not null,
    account_number varchar(50) not null,
    pix_token varchar(255)not null,
    pix_name varchar(255) not null,
    pix_city varchar(255) not null,
    logo_name varchar(255),
    created_at timestamp not null,
    updated_at timestamp not null,
    last_rps bigint not null,
    smtp_host varchar(255),
    smtp_port int,
    smtp_user varchar(255),
    smtp_password varchar(255),
    constraint unique_vendor_document unique(document),
    constraint unique_vendor_nickname unique(nickname)
);

# main vendor
insert into vendor (nickname, legal_name, trading_name, document, tax_document, account_bank, account_agency, account_number, pix_token, pix_name, pix_city, logo_name, email, whatsapp, last_rps, smtp_host, smtp_port, smtp_user, smtp_password, created_at, updated_at) values
('estudio_amelia', 'Cardoso e Barbosa Serviços Musicais e Tecnologia LTDA', 'Estúdio Amélia Cardoso', '27.928.875/0001-04', '5.727.888-1', '033 - Santander', '0985', '13001001-4', '27.928.875/0001-04', 'Estúdio Vocal Amélia Cardoso', 'São Paulo', 'logo_amelia.png', 'financeiro@ameliacardoso.com.br', '(11) 98088-8399', 2435, 'smtp.zoho.com', 465, 'financeiro@ameliacardoso.com.br', 'pwd22Adm**', now(), now());

# customer table
drop table if exists customer cascade;
create table customer (
    id bigserial primary key,
    name varchar(150) not null,
    vendor_id bigint not null references vendor(id) on delete cascade,
    nickname varchar(150) not null,
    document varchar(50),
    email varchar(150),
    whatsapp varchar(20),
    created_at timestamp not null,
    updated_at timestamp not null,
    status int not null default 1,
    constraint unique_customer_document unique(vendor_id, document),
    constraint unique_customer_nickname unique(vendor_id, nickname)
);

# session new
drop table if exists session;
create table session (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    session_date date not null,
    session_minutes int not null,
    session_service varchar(100) not null,
    session_status varchar(50) not null, -- realizada, cancelada_cobrar, cancelada_nao_cobrar 
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    constraint fk_customer_id foreign key (customer_id) references customer(id) on delete cascade
);

---------------------------------------
-- plans and services
---------------------------------------
drop table if exists service cascade;
create table service (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    name varchar(150) not null,
    created_at timestamp not null,
    updated_at timestamp not null
);

drop table if exists offer;
create table offer (
    id bigserial primary key,
    name varchar(150) not null,
    invoice_description varchar(255) not null,
    service_id bigint not null references service(id) on delete cascade,
    session_periodicity int not null, -- 0 - once, 1 - Weekly, 2 - Biweekly, 3 - Monthly, 4 - separate (payment_type just post-paid)
    session_minutes int not null,
    session_price numeric(15, 2) not null,
    sessions_month_limit int null, -- null - limited
    status int not null default 1,
    created_at timestamp not null,
    updated_at timestamp not null
);

drop table if exists subscription;
create table subscription (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    payment_day int not null, -- 1 - 31
    payment_type int not null, -- 1 - pre-paid, 2 - post-paid, 3 - per-session
    fixed_price numeric(15, 2) null, -- null = price per session
    status int not null default 1,
    created_at timestamp not null,
    updated_at timestamp not null
);

drop table if exists subscription_item;
create table subscription_item (
    id bigserial primary key,
    subscription_id bigint not null references subscription(id) on delete cascade,
    offer_id bigint not null references offer(id) on delete cascade,
    session_start_at date not null,
    session_end_at date null,
    session_day int not null, -- 1 - Monday, 2 - Tuesday, 3 - Wednesday, 4 - Thursday, 5 - Friday, 6 - Saturday, 7 - sunday, 0 - loose
    fixed_price numeric(15, 2) null, -- null - price per session
    status int not null default 1,
    created_at timestamp not null,
    updated_at timestamp not null
);

-------------------------------------------------------
-- 1. Ofertas e Serviços
-------------------------------------------------------
drop table if exists service cascade;
create table service (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    name varchar(150) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

drop table if exists offer cascade;
create table offer (
    id bigserial primary key,
    service_id bigint not null references service(id) on delete cascade,
    name varchar(150) not null,
    invoice_description varchar(255) not null,
    session_periodicity int not null, -- 0: Avulsa, 1: Semanal, 2: Quinzenal, 3: Mensal
    session_minutes int not null,
    session_price numeric(15, 2) not null,
    sessions_month_expected int null, -- Qtd estimada no mês (ex: 4 para semanal)
    status int not null default 1,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

-------------------------------------------------------
-- 2. Assinaturas / Contratos
-------------------------------------------------------
drop table if exists subscription cascade;
create table subscription (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    payment_day int not null check (payment_day between 1 and 31),
    payment_type int not null,         -- 1: Pré-pago, 2: Pós-pago
    monthly_fixed_price numeric(15, 2) null, -- Se o cliente paga um valor fixo fechado mensal
    status int not null default 1,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

drop table if exists subscription_item cascade;
create table subscription_item (
    id bigserial primary key,
    subscription_id bigint not null references subscription(id) on delete cascade,
    offer_id bigint not null references offer(id) on delete cascade,
    session_start_at date not null,
    session_end_at date null,
    session_day int not null,          -- 1 a 7 (Segunda a Domingo), 0: Avulso / sem dia fixo
    custom_price numeric(15, 2) null,  -- Preço negociado pelo serviço (sobrescreve offer.session_price)
    status int not null default 1,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

-------------------------------------------------------
-- 3. Sessões Realizadas (Registro Manual do Dia a Dia)
-------------------------------------------------------
drop table if exists session cascade;
create table session (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    offer_id bigint null references offer(id) on delete set null,
    session_date date not null,
    session_minutes int not null,
    session_status varchar(50) not null, -- 'realizada', 'cancelada_cobrar', 'cancelada_nao_cobrar'
    reconciled_at timestamp null,        -- Preenchido quando entra no fechamento do mês
    invoice_id bigint null,              -- Vínculo com a fatura gerada (se aplicável)
    comments text,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted_at timestamp null
);

-------------------------------------------------------
-- 4. Banco de Horas / Créditos de Sessões
-------------------------------------------------------
drop table if exists time_bank cascade;
create table time_bank (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    offer_id bigint null references offer(id) on delete set null,
    reference_month date not null,      -- Ex: '2026-10-01' (mês de competência da divergência)
    minutes_balance int not null,       -- Positivo: crédito (+60 min), Negativo: débito/consumo (-60 min)
    sessions_balance int not null,      -- Positivo: +1 aula, Negativo: -1 aula
    entry_type varchar(50) not null,    -- 'confronto_mensal', 'ajuste_manual', 'consumo_credito'
    notes text,                         -- Motivo do ajuste (ex: "Aluno faltou aula dia 15/10 com aviso prévio")
    created_at timestamp not null default now()
);

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

-------------------------------------------------------
-- Atualização na tabela session para rastreabilidade
-------------------------------------------------------
-- Cada sessão vincula-se diretamente ao fechamento que a faturou
alter table session 
    add column if not exists service_invoice_id bigint references service_invoice(id) on delete set null;


