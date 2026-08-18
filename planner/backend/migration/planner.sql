create schema if not exists planner;

set search_path to planner;


create table if not exists session (
    id bigserial primary key,
    customer_nickname varchar(150) not null,
    session_date date not null,
    session_minutes int not null,
    session_status varchar(50) not null, -- realizada, cancelada_cobrar, cancelada_nao_cobrar 
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null 
);


select * from session;