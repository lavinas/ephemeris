create schema if not exists planner;

set search_path to planner;

# plan_invoicing
# plan_invoicing é uma tabela de controle de faturas dos planos
create table invoice (
    id bigserial primary key,
    plan_id bigint not null references plan(id) on delete cascade,
    invoicing_date date not null,
    due_date date not null,
    amount numeric(15, 2) not null,
    description text not null,
    reference_invoice_id bigint null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

# plan_invoicing_item
# plan_invoicing_item é uma tabela de controle de itens de faturas dos planos
create table invoice_item (
    id bigserial primary key,
    plan_invoicing_id bigint not null references invoice(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    price numeric(15, 2) not null,
    description varchar(255) null,
    amount numeric(15, 2) not null,
    invoiced_quantity int not null,
    realized_quantity int not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create table invoice_item_session (
    id bigserial primary key,
    invoice_item_id bigint not null references invoice_item(id) on delete cascade,
    session_id bigint not null references session(id) on delete cascade,
    created_at timestamp not null
);