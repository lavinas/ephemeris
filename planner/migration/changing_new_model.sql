-- Active: 1790470840125@@192.168.1.138@5433@planner@planner
select count(1) from customer;


select count(1)
  from session a;

select * from session;

select count(1), min(a.session_date), max(a.session_date)
  from session a
left join customer b
    on b.nickname = a.customer_nickname
where b.nickname is null;


select * from customer;

drop table tmp_customer;

create table tmp_customer as
select distinct a.customer_nickname
  from session a
left join customer b
    on b.nickname = a.customer_nickname
where b.nickname is null;

select count(1) from tmp_customer;

insert into customer (name, vendor_id, nickname, document, email, whatsapp, created_at, updated_at, status)
select a.customer_nickname, 1, a.customer_nickname, null, null, null, now(), now(), 1
  from tmp_customer a;

select count(1) from customer;

alter table session rename to session_old;

create table session (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    session_date date not null,
    session_minutes int not null,
    session_service varchar(100) not null,
    session_status varchar(50) not null, -- realizada, cancelada_cobrar, cancelada_nao_cobrar 
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    constraint fk_customer_id foreign key (customer_id) references customer(id) on delete cascade
);

insert into session (customer_id, session_date, session_minutes, session_service, session_status, comments, created_at, updated_at)
select b.id, session_date, session_minutes, session_service, session_status, comments, a.created_at, a.updated_at
  from session_old a
  join customer b on b.nickname = a.customer_nickname;

select count(1) from session;
select count(1) from session_old;