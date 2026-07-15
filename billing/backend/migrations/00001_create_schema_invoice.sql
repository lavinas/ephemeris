-- Active: 1778275768971@@localhost@5432@ephemeris@billing

create database ephemeris;

create SCHEMA if not exists billing;
set search_path to billing;

# vendor table
drop table if exists vendor cascade;
create table vendor (
    id bigserial primary key,
    legal_name varchar(150) not null,
    nickname varchar(150),
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
    constraint unique_vendor_document unique(document)
);

# main vendor
insert into vendor (legal_name, nickname, document, tax_document, account_bank, account_agency, account_number, pix_token, pix_name, pix_city, logo_name, email, whatsapp, created_at, updated_at, last_rps) values
('Cardoso e Barbosa Servicos Musicais e Tecnologia LTDA', 'estudio_amelia', '27.928.875/0001-04', '27.928.875/0001-04', 'Santander (033)', '0985', '13001001-4', '27.928.875/0001-04', 'Estudio Vocal Amelia Cardoso', 'Sao Paulo', 'logo_amelia.png', 'financeiro@ameilacardoso.com.br', '(11) 98088-8399', now(), now(), 2435);


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

drop Table if exists invoice cascade;
create table invoice (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    amount numeric(15, 2) not null,
    invoice_date date not null,
    due_date date not null,
    payment_date date,
    email_sent_date date,
    whatsapp_sent_date date,
    cancellation_date date,
    tax_date date,
    notes text null,
    status int not null default 1,
    created_at timestamp not null,
    updated_at timestamp not null
);

drop Table if exists invoice_item cascade;
create table invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references invoice(id) on delete cascade,
    price numeric(15, 2) not null,
    quantity int not null,
    description varchar(255) not null,
    created_at timestamp not null,
    updated_at timestamp not null
);


drop Table if exists emission cascade;
create table emission (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    emission_date date not null ,
    period_start date not null,
    period_end date not null,
    rps_start bigint not null,
    rps_end bigint not null,
    nfe_start bigint,
    nfe_end bigint,
    nfe_datetime timestamp,
    amount numeric(15, 2) not null,
    quantity int not null,
    created_at timestamp not null,
    updated_at timestamp not null
);

drop Table if exists emission_item cascade;
create table emission_item (
    id bigserial primary key,
    emission_id bigint not null references emission(id) on delete cascade,
    invoice_id bigint not null references invoice(id) on delete cascade,
    rps_number bigint not null,
    nfe_number bigint,
    nfe_datetime timestamp,
    nfe_verification varchar(100)
);



