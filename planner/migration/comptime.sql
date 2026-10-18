
create table comptime (
    id serial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade
    session_time time not null,
    invoiced_time time not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create table comptime_session (
    comptime_id bigint not null references comptime(id) on delete cascade,
    session_id bigint not null references session(id) on delete cascade,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create table comptime_session_invoice (
    comptime_id bigint not null references comptime(id) on delete cascade,
    invoice_id bigint not null references invoice(id) on delete cascade,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);