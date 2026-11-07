-- Active: 1790470840125@@192.168.1.138@5433@planner@planner
create table tmp_service as
select * from service;

select count(1) from tmp_service;

drop table service cascade;

create table service (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    name varchar(150) not null,
    description text,
    session_minutes int null, -- if there is a predicted duration for the service
    nbs_code numeric(10,0),
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

insert into service
select id, vendor_id, name, description, session_minutes, 122051100, created_at, updated_at from tmp_service;

select count(1) from service;


select setval(pg_get_serial_sequence('service', 'id'), coalesce(max(id), 0) + 1, false) from service;

alter table session add constraint fk_session_service foreign key (service_id) references service(id);

alter table invoice_item add constraint fk_invoice_item_service foreign key (service_id) references service(id);

drop table tmp_service;