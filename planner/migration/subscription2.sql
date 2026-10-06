set search_path to planner;

drop table if exists service cascade;
create table service (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    name varchar(150) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

-- if there are two or more subsctiotion in the same weekday and time, then it will be alternating sessions
-- if there are two or more subscription in the same due_day, then it will be in the same invoice
create table subscription (
    id bigserial primary key,
    -- customer data
    customer_id bigint not null references customer(id) on delete cascade,
    -- service data
    service_id bigint not null references service(id) on delete cascade,
    session_minutes int not null, -- minutes per session
    session_weekday int not null, -- 1-7, or null if custom, 0 if custom
    session_time time null, -- 14:00, 14:30, etc, or null if custom or no definition
    session_recurrence int not null, -- 0: single, 1: weekly, 2: bi-weekly, 3: monthly
    session_monthly_limit int not null, -- if null, there is no limit  
    session_order int not null default 0, -- order of the session in the month (if there is more then one sunscription in same weekday and time)
    -- conditions
    price_per_session numeric(15, 2) not null default 0, -- if null, use service price_per_session
    price_per_month numeric(15, 2) not null default 0, -- if null, use service price_per_month
    -- invoicing data
    invoice_type int not null, -- 1: pre-paid (just for recurrence, not for single), 2: post-paid, 3: per session
    due_day int not null, -- 1-30, 0 if per_session
    -- dates
    start_at date not null, 
    end_at date null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);




-- service_invoice 
create table service_invoice (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    invoice_date date not null, -- date the invoice is issued
    due_date date not null, -- date the invoice is due
    total_amount numeric(15, 2) not null default 0
);

create table service_invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references service_invoice(id) on delete cascade,
    subscription_id bigint not null references subscription(id) on delete cascade,
    reference_month date not null, -- month this invoice refers to
    sessions_amount int not null,
    amount numeric(15, 2) not null default 0,
);


-- service_invoice_reconciliation
create table service_invoice_reconciliation (
    id bigserial primary key,
    invoice_id bigint not null references service_invoice(id) on delete cascade,
    subscription_id bigint not null references subscription(id) on delete cascade,
    reference_month date not null, -- month this invoice refers to
    sessions_amount int not null,
    amount numeric(15, 2) not null default 0,
);