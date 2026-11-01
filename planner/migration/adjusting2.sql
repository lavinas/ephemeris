set search_path to planner;

create table tmp_invoice as select * from invoice;
create table tmp_invoice_item as select * from invoice_item;

select count(1) from tmp_invoice;
select count(1) from tmp_invoice_item;


drop table invoice_item cascade;

drop table invoice cascade;


create table invoice (
    id bigserial primary key,
    plan_id bigint null references plan(id) on delete cascade,
    invoicing_date date not null,
    due_date date not null,
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
    period_start date not null,
    period_end date not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_invoice_item_invoice_id on invoice_item(invoice_id);

insert into invoice
select id, plan_id, invoicing_date, due_date, amount, reference_invoice_id, created_at, updated_at, deleted_at 
from tmp_invoice;

insert into invoice_item
select a.id, a.invoice_id, a.service_id, a.price, a.description, a.quantity, b.period_start, b.period_end, a.created_at, a.updated_at, a.deleted_at 
  from tmp_invoice_item a
inner join tmp_invoice b on b.id = a.invoice_id

select setval(pg_get_serial_sequence('invoice', 'id'), coalesce(max(id), 0) + 1, false) from invoice;

select setval(pg_get_serial_sequence('invoice_item', 'id'), coalesce(max(id), 0) + 1, false) from invoice_item;

alter table comptime_invoice_item add constraint fk_comptime_invoice_item_invoice_id foreign key (invoice_item_id) references invoice_item(id);

drop table tmp_invoice;
drop table tmp_invoice_item;