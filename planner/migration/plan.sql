-- Active: 1790015123887@@127.0.0.1@5433@planner@planner
create schema if not exists planner;

set search_path to planner;

# plan
create table plan (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    plan_start date not null,
    plan_end date null,
    plan_type int not null check(plan_type in (1,2,3)), -- 1: monthly agenda (agenda), 2: package (pacote), 3: notebook (caderneta)
    price numeric(15, 2) null, -- if not null, then override the item prices
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

# plan_item 
create table plan_item (
    id bigserial primary key,
    plan_id bigint not null references plan(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    price numeric(15, 2) null, -- if null, use service price
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

# agenda plan
create table plan_agenda (
    id bigserial primary key references plan(id) on delete cascade,
    recurrence int not null check (recurrence in (1, 2, 3)), -- 1: weekly, 2: bi-weekly, 3: monthly
    week_day int not null check (week_day in (1, 2, 3, 4, 5, 6, 7)), -- 1: sunday, 2: monday, 3: tuesday, 4: wednesday, 5: thursday, 6: friday, 7: saturday
    term_type int not null check (term_type in (1, 2)), -- 1 - pre-paid, 2 - post-paid
    payment_day int not null check (payment_day between 1 and 31), -- day of the month to charge (1-31)
    monthly_service_limit int null check (monthly_service_limit > 0) -- limit number of sessions per month, if null, no limit,
);

# plan_package
create table plan_package (
    id bigserial primary key references plan(id) on delete cascade,
    service_quantity int not null check (service_quantity > 0), -- quantity of sessions
    payment_date date not null 
);

# notebook plan
create table plan_notebook (
    id bigserial primary key references plan(id) on delete cascade,
    session_payment_day int null check (session_payment_day between 1 and 31),
    session_payment_term int null check (session_payment_term > 0),
    constraint check_payment_day_or_term_exclusive 
        check (num_nonnulls(session_payment_day, session_payment_term) = 1)
);


select count(1) from customer;


select * from service;

    select * from customer where id = 268;