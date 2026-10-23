-- Active: 1790015123887@@127.0.0.1@5433@planner@planner
create schema if not exists planner;

set search_path to planner;

drop table if exists invoice cascade;
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

create index idx_invoice_plan_id on invoice(plan_id);



drop table if exists invoice_item cascade;

# plan_invoicing_item
# plan_invoicing_item é uma tabela de controle de itens de faturas dos planos
create table invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references invoice(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    price numeric(15, 2) not null,
    description varchar(255) null,
    quantity int not null, -- new
    -- invoiced_quantity int not null, -- old
    -- realized_quantity int not null, -- old
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_invoice_item_invoice_id on invoice_item(invoice_id);
