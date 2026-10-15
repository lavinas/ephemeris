-- Active: 1790015123887@@127.0.0.1@5433@planner@planner
create schema if not exists planner;

set search_path to planner;


drop table invoice;
# plan_invoicing
# plan_invoicing é uma tabela de controle de faturas dos planos
create table invoice (
    id bigserial primary key,
    plan_id bigint not null references plan(id) on delete cascade,
    invoicing_date date not null,
    due_date date not null,
    period_start date not null,
    period_end date not null,
    amount numeric(15, 2) not null,
    reference_invoice_id bigint null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

drop table invoice_item;

# plan_invoicing_item
# plan_invoicing_item é uma tabela de controle de itens de faturas dos planos
create table invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references invoice(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    price numeric(15, 2) not null,
    description varchar(255) null,
    invoiced_quantity int not null,
    realized_quantity int not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

drop table invoice_item_session;

create table invoice_item_session (
    id bigserial primary key,
    invoice_item_id bigint not null references invoice_item(id) on delete cascade,
    session_id bigint not null references session(id) on delete cascade,
    created_at timestamp not null
);