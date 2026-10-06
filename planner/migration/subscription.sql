-- Active: 1790015123887@@127.0.0.1@5433@planner@planner
-------------------------------------------------------
-- 1. Services, plans and offers
-------------------------------------------------------

set search_path to planner;

drop table if exists service cascade;
create table service (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    name varchar(150) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

insert into service values (1, 1, 'aula/canto', now(), now());
insert into service values (2, 1, 'aula/piano', now(), now());

create table plan (
    id bigserial primary key,
    service_id bigint not null references service(id) on delete cascade,
    name varchar(150) not null, -- eg: "aulas de canto semanais de 60 minutos", "aulas de piano semanais de 60 minutos", etc
    session_recurrence int not null, -- 0: Single, 1: Weekly, 2: Bi-Weekly, 3: Monthly
    session_minutes int not null, -- session duration in minutes
    sessions_monthly_limit int null, -- max sessions per month
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

drop table if exists offer cascade;
create table offer (
    id bigserial primary key,
    name varchar(150) not null, -- eg: "aulas de canto semanais de 60 minutos / R$ 120"
    start_at date not null, -- when the offer starts (usually same as vendor_id 'from')
    end_at date null, -- when the offer ends (usually same as vendor_id 'to')
    monthly_fixed_price numeric(15, 2) null, -- total fixed price of the offer, if null, it is valid price per session
    invoice_description varchar(150) null, -- short description to be used in the invoice, override offer_item.invoice_description if not null
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table offer_item (
    id bigserial primary key,
    offer_id bigint not null references offer(id) on delete cascade,
    plan_id bigint not null references plan(id) on delete cascade,
    session_price numeric(15, 2) null,   -- custom price for the item
    invoice_description varchar(150) null, -- short description to be used in the invoice    
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

-------------------------------------------------------
-- 2. Assinaturas / Contratos
-------------------------------------------------------
drop table if exists subscription cascade;
create table subscription (
    id bigserial primary key,
    offer_id bigint not null references offer(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    payment_day int not null check (payment_day between 1 and 31),
    payment_type int not null, -- 1: monthly pre-paid (just for recurrence), 2: monthly post-paid
    status int not null default 1,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

drop table if exists subscription_item cascade;
create table subscription_item (
    id bigserial primary key,
    subscription_id bigint not null references subscription(id) on delete cascade,
    offer_item_id bigint not null references offer_item(id) on delete cascade,
    day_of_week int not null, -- 1 to 7 (Monday to Sunday), 0: Avulso / sem dia fixo
    start_at date not null,
    end_at date null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
