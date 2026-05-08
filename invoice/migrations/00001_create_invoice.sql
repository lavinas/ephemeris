-- Active: 1778275768971@@127.0.0.1@5432@ephemeris

create SCHEMA if not exists invoice;
set search_path to invoice;

drop Table if exists customer cascade;
create table customer (
    id bigserial primary key,
    customer_type varchar(20) not null, 
    nickname varchar(100) unique not null,
    name varchar(150) not null,
    email varchar(150) not null,
    whatsapp varchar(20) not null,
    document varchar(50),
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

drop Table if exists product cascade;
create table product (
    id serial primary key,
    name varchar(255) not null,
    price numeric(10, 2) not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

drop Table if exists invoice cascade;
create table invoice (
    id serial primary key,
    customer_id bigint references customer(id) on delete set null,
    customer_name varchar(150) not null,
    customer_email varchar(150) not null,
    customer_whatsapp varchar(20) not null,
    customer_document varchar(50),
    amount numeric(10, 2) not null,
    notes text,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

drop Table if exists invoice_item cascade;
create table invoice_item (
    id serial primary key,
    invoice_id int references invoice(id) on delete cascade,
    product_id int references product(id) on delete set null,
    product_name varchar(255) not null,
    quantity int not null,
    price numeric(10, 2) not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);