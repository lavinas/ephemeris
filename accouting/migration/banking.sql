-- Active: 1793746833660@@127.0.0.1@5435@accounting@accounting

create schema if not exists accounting;

set search_path to accounting;

-- Active: 1793746833660@@127.0.0.1@5435@accounting
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

insert into vendor (nickname, legal_name, trading_name, document, tax_document, account_bank, account_agency, account_number, pix_token, pix_name, pix_city, logo_name, email, whatsapp, last_rps, smtp_host, smtp_port, smtp_user, smtp_password, created_at, updated_at) values
('estudio_amelia', 'Cardoso e Barbosa Serviços Musicais e Tecnologia LTDA', 'Estúdio Amélia Cardoso', '27.928.875/0001-04', '5.727.888-1', '033 - Santander', '0985', '13001001-4', '27.928.875/0001-04', 'Estúdio Vocal Amélia Cardoso', 'São Paulo', 'logo_amelia.png', 'financeiro@ameliacardoso.com.br', '(11) 98088-8399', 2435, 'smtp.zoho.com', 465, 'financeiro@ameliacardoso.com.br', 'pwd22Adm**', now(), now());


create table bank_account (
    id bigserial primary key,
    vendor_id bigint not null,
    bank_code varchar(10) not null,
    account_number varchar(50) not null,
    account_branch varchar(20) not null,
    status int not null default 1, -- 0 - inactive, 1 - active
    initial_balance numeric(10,2) not null,
    initial_date date not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    CONSTRAINT fk_bank_account_vendor foreign key (vendor_id) references vendor(id),
    CONSTRAINT uk_bank_account_vendor_agency_number unique(vendor_id, account_branch, account_number)
);

insert into bank_account (vendor_id, bank_code, account_number, account_branch, status, initial_balance, initial_date, created_at, updated_at) values
(1, '033', '13001001-4', '0985', 1, 1135.51, '2026-10-01', now(), now());

create index idx_bank_account_vendor_id on bank_account(vendor_id);

create table bank_account_santander_api(
    id bigint primary key references bank_account(id),
    client_id varchar(50) not null,
    client_secret varchar(255) not null
);


insert into bank_account_santander_api (id, client_id, client_secret) values
(1, 'Xi6jcH4HxPCV2JAxshFOnG0SG22oiG7Z', 'C3mUGaA3EQyOFDl1');


create table bank_statement (
    id bigserial primary key,
    account_id bigint not null references bank_account(id),
    sdate date not null,
    transaction varchar(200) not null,
    complement varchar(200),
    document varchar(100),
    amount numeric(10,2) not null,
    balance numeric(10,2) not null,
    key varchar(100) not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_bank_statement_account_id on bank_statement(account_id);

create unique index uk_bank_statement_key on bank_statement(key);

create index idx_bank_statement_date_account_id on bank_statement(account_id, sdate);

create table bank_statement_santander_api (
    id bigserial primary key references bank_statement(id),
    credit_debit_type varchar(10),
    transaction_name varchar(100),
    historic_complement varchar(100),
    amount numeric(10,2),
    transaction_date date,
    document_number varchar(50),
    key varchar(100)
);


create table bank_account_action (
    id bigserial PRIMARY KEY,
    account_id bigint not null references bank_account(id),
    action int not null, -- 1 - ignore line
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_bank_account_action_account_id on bank_account_action(account_id);

create table bank_account_action_rule (
    id bigserial PRIMARY KEY,
    action_id bigint not null references bank_account_action(id),
    field int not null, -- 1 - date, 2 - transaction, 3 - complement, 4 - document, 5 - amount
    operator int not null, -- 1 - equals, 2 - contains, 3 - starts_with, 4 - ends_with, 5 - regex, 6 - greater_than, 7 - less_than
    value varchar(255) not null,
    case_sensitive boolean default false,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_bank_account_action_rule_action_id on bank_account_action_rule(action_id);


