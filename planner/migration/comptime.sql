
set search_path to planner;


drop table if exists comptime cascade;
drop table if exists comptime_session cascade;
drop table if exists comptime_invoice_item cascade;
drop table if exists comptime_extra cascade;


create table comptime (
    id serial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    balance interval not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);
create unique index idx_comptime_customer_id_service_id on comptime(customer_id, service_id);



create table comptime_session (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    session_id bigint not null references session(id) on delete cascade,
    ref_date date not null,
    ref_month date not null,
    balance_changed interval not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create unique index idx_comptime_session_session_id on comptime_session(session_id);
create index idx_comptime_session_comptime_id on comptime_session(comptime_id);
create index idx_comptime_session_ref_date on comptime_session(ref_date);
create index idx_comptime_session_ref_month on comptime_session(ref_month);


create table comptime_invoice_item (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    invoice_item_id bigint not null references invoice_item(id) on delete cascade,
    ref_date date not null,
    ref_month date not null,
    balance_changed interval not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create unique index idx_comptime_invoice_item_invoice_item_id on comptime_invoice_item(invoice_item_id);
create index idx_comptime_invoice_item_comptime_id on comptime_invoice_item(comptime_id);
create index idx_comptime_invoice_item_ref_date on comptime_invoice_item(ref_date);
create index idx_comptime_invoice_item_ref_month on comptime_invoice_item(ref_month);


create table comptime_extra (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    ref_date date not null,
    ref_month date not null,
    balance_changed interval not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_comptime_extra_comptime_id on comptime_extra(comptime_id);
create index idx_comptime_extra_ref_date on comptime_extra(ref_date);
create index idx_comptime_extra_ref_month on comptime_extra(ref_month);