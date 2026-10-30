-- Active: 1790015123887@@127.0.0.1@5433@planner@planner

set search_path to planner;

create table tmp_comptime as
select * from comptime;

drop table comptime cascade;

create table comptime (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    month date not null,
    invoiced_minutes int not null default 0,
    consumed_minutes int not null default 0,
    adjusted_minutes int not null default 0,
    balance int not null default 0,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);
create index idx_comptime_vendor_id on comptime(vendor_id);
create index idx_comptime_customer_id on comptime(customer_id);
create index idx_comptime_service_id on comptime(service_id);
create index idx_comptime_customer_id_service_id on comptime(customer_id, service_id);
create index idx_comptime_month on comptime(month);
create unique index idx_comptime_unique_month on comptime(vendor_id, customer_id, service_id, month) where deleted_at is null;

insert into comptime
select id, vendor_id, customer_id, service_id, month, invoiced_minutes, consumed_minutes, 0, balance, created_at, updated_at, deleted_at
  from tmp_comptime;

select setval(pg_get_serial_sequence('comptime', 'id'), coalesce(max(id), 0) + 1, false) from comptime;

drop table comptime_extra;

create table comptime_adjust (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    month date not null,
    adjusted_minutes int not null,
    description varchar(250) not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_comptime_adjust_comptime_id on comptime_adjust(comptime_id);
create index idx_comptime_adjust_month on comptime_adjust(month);

alter table comptime_session 
  add constraint fk_comptime_session_comptime
foreign key (comptime_id)
references comptime(id);

alter table comptime_invoice_item 
  add constraint fk_comptime_invoice_item_comptime
foreign key (comptime_id)
references comptime(id);
