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

delete from customer;

select * from 

# session old
/*
drop table if exists session;
create table session (
    id bigserial primary key,
    customer_nickname varchar(150) not null,
    session_date date not null,
    session_minutes int not null,
    session_service varchar(100) not null,
    session_status varchar(50) not null, -- realizada, cancelada_cobrar, cancelada_nao_cobrar 
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);
*/

## Adjusting table

alter table session rename to session_old;

insert into customer (name, vendor_id, nickname, document, email, whatsapp, created_at, updated_at, status) 
select distinct customer_nickname, 1, customer_nickname, null, null, null, now(), now(), 1
  from session_old;

# session new
drop table if exists session;
create table session (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    session_date date not null,
    session_minutes int not null,
    session_service varchar(100) not null,
    session_status varchar(50) not null, -- realizada, cancelada_cobrar, cancelada_nao_cobrar 
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

# finalize adjusting session (i need a status name)
insert into session (vendor_id, customer_id, session_date, session_minutes, session_service, session_status, comments, created_at, updated_at)
select 1, b.id, session_date, session_minutes, session_service, session_status, comments, now(), now()
  from session_old a
  join customer b on b.nickname = a.customer_nickname;


drop table session_old;