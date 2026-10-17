-- Active: 1790014326402@@127.0.0.1@5432@billing@billing

create SCHEMA if not exists billing;


create table tmp_invoices as (select * from invoice);
create table tmp_invoice_item as (select * from invoice_item);
create table tmp_emission as (select * from emission);
create table tmp_emission_item as (select * from emission_item);


drop table emission_item cascade;

drop table emission cascade;

drop table invoice_item cascade;

drop table invoice cascade;

create table invoice (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    amount numeric(15, 2) not null,
    invoice_date date not null,
    due_date date not null,
    payment_date date,
    sent_date date,
    receipt_date date,
    cancellation_date date,
    tax_date date,
    notes text null,
    external_code varchar(50) null,
    status int not null default 1,
    created_at timestamp not null,
    updated_at timestamp not null,
    constraint invoice_uk1 unique(vendor_id, external_code)
);
create index idx_invoice_customer_id on invoice(customer_id);

create table invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references invoice(id) on delete cascade,
    price numeric(15, 2) not null,
    quantity int not null,
    description varchar(255) not null,
    created_at timestamp not null,
    updated_at timestamp not null
);
create index idx_invoice_item_invoice_id on invoice_item(invoice_id);

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
create index idx_emission_vendor_id on emission(vendor_id);
create index idx_emission_date on emission(emission_date);

create table emission_item (
    id bigserial primary key,
    emission_id bigint not null references emission(id) on delete cascade,
    invoice_id bigint not null references invoice(id) on delete cascade,
    rps_number bigint not null,
    nfe_number bigint,
    nfe_datetime timestamp,
    nfe_verification varchar(100),
    nfe_amount numeric(15, 2)
);
create index idx_emission_item_emission_id on emission_item(emission_id);
create index idx_emission_item_invoice_id on emission_item(invoice_id);


select * from invoice

insert into invoice 
select id, 1, customer_id, amount, invoice_date, due_date, payment_date, email_sent_date, email_receipt_date, cancellation_date, tax_date, notes, null, status, created_at, updated_at    
  from tmp_invoices;

insert into invoice_item 
select id, invoice_id, price, quantity, description, created_at, updated_at
  from tmp_invoice_item;

insert into emission 
select id, vendor_id, emission_date, period_start, period_end, rps_start, rps_end, nfe_start, nfe_end, nfe_datetime, amount, quantity, created_at, updated_at
  from tmp_emission;

insert into emission_item 
select id, emission_id, invoice_id, rps_number, nfe_number, nfe_datetime, nfe_verification, nfe_amount
  from tmp_emission_item;


select count(*) from invoice;
select count(*) from invoice_item;
select count(*) from emission;
select count(*) from emission_item;

drop table tmp_invoices;
drop table tmp_invoice_item;
drop table tmp_emission;
drop table tmp_emission_item;