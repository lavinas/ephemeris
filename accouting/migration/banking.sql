-- Active: 1793746833660@@127.0.0.1@5435@accounting

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


create table account (
    id bigserial primary key,
    vendor_id bigint not null,
    bank_code varchar(10) not null,
    account_number varchar(50) not null,
    account_branch varchar(20) not null,
    status int not null default 1, -- 0 - inactive, 1 - active
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    CONSTRAINT fk_account_vendor foreign key (vendor_id) references vendor(id),
    CONSTRAINT uk_account_vendor_agency_number unique(vendor_id, account_branch, account_number)
);

insert into account (vendor_id, bank_code, account_number, account_branch, created_at, updated_at) values
(1, '033', '13001001-4', '0985', now(), now());

create index idx_account_vendor_id on account(vendor_id);

create table account_santander_api(
    id bigint primary key references account(id),
    client_id varchar(50) not null,
    client_secret varchar(255) not null
);

insert into account_santander_api (id, client_id, client_secret) values
(1, 'aSsooLX631tn49AdNXxAfUBsQZ2g0OWSV', 'cm4zpTdWhD9JbfGa');


create table statement (
    id bigserial primary key,
    account_id bigint not null references account(id),
    sdate date,
    amount numeric(10,2),
    memo varchar(200),
    partie varchar(50),
    checksum varchar(100),
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_statement_account_id on statement(account_id);

create unique index unique_statement_checksum on statement(checksum);

create table statement_santander_api (
    id bigserial primary key references statement(id),
    transaction_id varchar(100),
    transaction_status varchar(100),
    transaction_sign varchar(1),
    transaction_name varchar(200),
    transaction_type varchar(100),
    transaction_amount numeric(10,2),
    transaction_currency varchar(3),
    transaction_date date,
    partie_document varchar(50),
    partie_type varchar(100),
    partie_bank_code varchar(10),
    partie_branch_code varchar(10),
    partie_account varchar(50),
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

