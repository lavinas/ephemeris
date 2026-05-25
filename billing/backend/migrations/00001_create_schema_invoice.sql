-- Active: 1778275768971@@127.0.0.1@5432@ephemeris

create database ephemeris;

create SCHEMA if not exists billing;
set search_path to billing;

# business table
drop table if exists business cascade;
create table business (
    id bigserial primary key,
    legal_name varchar(150) not null,
    trade_name varchar(150),
    document varchar(50) not null,
    account_bank varchar(100) not null,
    account_agency varchar(20) not null,
    account_number varchar(50) not null,
    pix_token varchar(255)not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    constraint unique_business_document unique(document)
);
# main business
insert into business (legal_name, trade_name, document, account_bank, account_agency, account_number, pix_token, created_at, updated_at) values
('Cardoso e Barbosa Servicos Musicais e Tecnologia LTDA', 'Estudio Amelia Cardoso', '27.928.875/0001-04', 'Santander (033)', '0985', '13001001-4', 'CNPJ: 27.928.875/0001-04', now(), now());

# customer table
drop table if exists customer cascade;
create table customer (
    id bigserial primary key,
    name varchar(150) not null,
    nickname varchar(150) not null,
    status int not null default 1,
    document varchar(50),
    email varchar(150),
    whatsapp varchar(20),
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now(),
    constraint unique_customer_document unique(document),
    constraint unique_customer_nickname unique(nickname)
);


drop Table if exists invoice cascade;
create table invoice (
    id bigserial primary key,
    business_id bigint not null references business(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    amount numeric(15, 2) not null,
    notes text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

drop Table if exists invoice_item cascade;
create table invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references invoice(id) on delete cascade,
    description varchar(255) not null,
    quantity int not null,
    unit_price numeric(15, 2) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
