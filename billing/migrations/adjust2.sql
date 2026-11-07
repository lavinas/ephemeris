-- Active: 1790470738918@@192.168.1.138@5432@billing@billing
create table tmp_invoice_item as
select * from invoice_item;


select count(1) from tmp_invoice_item;

drop table if exists invoice_item cascade;

create table invoice_item (
    id bigserial primary key,
    invoice_id bigint not null references invoice(id) on delete cascade,
    price numeric(15, 2) not null,
    quantity int not null,
    description varchar(255) not null,
    nbs_code numeric(10,0) null,
    created_at timestamp not null,
    updated_at timestamp not null
);

insert into invoice_item
select id, invoice_id, price, quantity, description, 122051100, created_at, updated_at from tmp_invoice_item;


select count(1) from invoice_item;

create index idx_invoice_item_invoice_id on invoice_item(invoice_id);

select setval(pg_get_serial_sequence('invoice_item', 'id'), coalesce(max(id), 0) + 1, false) from invoice_item;

select * from invoice_item;

drop table tmp_invoice_item;