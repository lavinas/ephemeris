set search_path to planner;

drop table if exists comptime cascade;
create table comptime (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    customer_id bigint not null references customer(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    month date not null,
    invoiced_minutes int not null default 0,
    consumed_minutes int not null default 0,
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


drop table if exists comptime_session cascade;
create table comptime_session (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    session_id bigint not null references session(id) on delete cascade,
    month date not null,
    consumed_minutes int not null default 0,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_comptime_session_session_id on comptime_session(session_id);
create index idx_comptime_session_comptime_id on comptime_session(comptime_id);
create index idx_comptime_session_month on comptime_session(month);
create unique index idx_comptime_session_unique_active on comptime_session(session_id) where deleted_at is null;


drop table if exists comptime_invoice_item cascade;
create table comptime_invoice_item (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    invoice_item_id bigint not null references invoice_item(id) on delete cascade,
    month date not null,
    invoiced_minutes int not null default 0,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_comptime_invoice_item_invoice_item_id on comptime_invoice_item(invoice_item_id);
create index idx_comptime_invoice_item_comptime_id on comptime_invoice_item(comptime_id);
create index idx_comptime_invoice_item_month on comptime_invoice_item(month);
create unique index idx_comptime_invoice_item_unique_active on comptime_invoice_item(invoice_item_id) where deleted_at is null;

drop table if exists comptime_extra cascade;

create table comptime_extra (
    id bigserial primary key,
    comptime_id bigint not null references comptime(id) on delete cascade,
    month date not null,
    invoiced_minutes int not null default 0,
    consumed_minutes int not null default 0,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index idx_comptime_extra_comptime_id on comptime_extra(comptime_id);
create index idx_comptime_extra_month on comptime_extra(month);