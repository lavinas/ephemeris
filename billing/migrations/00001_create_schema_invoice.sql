-- Active: 1778275768971@@127.0.0.1@5432@ephemeris

create SCHEMA if not exists billing;
set search_path to billing;

drop Table if exists invoice cascade;
create table invoice (
    id serial primary key,
    customer_name varchar(150) not null,
    customer_email varchar(150) not null,
    customer_whatsapp varchar(20) not null,
    customer_document varchar(50),
    amount numeric(15, 2) not null,
    notes text,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

drop Table if exists invoice_item cascade;
create table invoice_item (
    id serial primary key,
    invoice_id int references invoice(id) on delete cascade,
    description varchar(255) not null,
    quantity int not null,
    unit_price numeric(15, 2) not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);