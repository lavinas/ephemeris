create schema if not exists planner;

set search_path to planner;

create table if not exists session_status (
    id int primary key,
    name varchar(50) not null,
    description varchar(255),
    created_at timestamp not null,
    updated_at timestamp not null
);

insert into session_status (id, name, description, created_at, updated_at) values
(1, 'agendada', 'Sessão agendada e aguardando início', now(), now()),
(2, 'realizada', 'Sessão realizada com sucesso', now(), now()),
(3, 'reagendar', 'Sessão não realizada e aguardando novo agendamento', now(), now()),
(4, 'reagendada', 'Sessão reagendada e uma nova sessão agendada', now(), now()),
(5, 'cancelada_cobrar', 'Sessão não realizada e deve ser cobrada', now(), now()),
(6, 'cancelada_nao_cobrar', 'Sessão não realizada e não deve ser cobrada', now(), now());

create table if not exists plan (
    id int primary key,
    name varchar(50) not null,
    description varchar(255),
    periodicity int, -- 1 - Única, 2 - Semanal, 3 - Quinzenal, 4 - Mensal
    duration int not null, -- minutes
    created_at timestamp not null,
    updated_at timestamp not null
);

insert into plan (id, name, description, periodicity, duration, created_at, updated_at) values
(1, 'inaugural_30', 'Sessão única avulsa', 1, 30, now(), now()),
(2, 'inaugural_45', 'Sessão única avulsa', 1, 45, now(), now()),
(3, 'inaugural_60', 'Sessão única avulsa', 1, 60, now(), now()),
(4, 'avulsa_30', 'Sessão avulsa de 30 minutos', 1, 30, now(), now()),
(5, 'avulsa_60', 'Sessão avulsa de 60 minutos', 1, 60, now(), now()),

(6, 'semanal_30', 'Plano com sessões semanais de 30 minutos', 2, 30, now(), now()),
(7, 'semanal_60', 'Plano com sessões semanais de 60 minutos', 2, 60, now(), now()),
(8, 'quinzenal_30', 'Plano com sessões quinzenais de 30 minutos', 3, 30, now(), now()),
(9, 'quinzenal_60', 'Plano com sessões quinzenais de 60 minutos', 3, 60, now(), now());
(2, 'inaugural_60', 'Sessão única avulsa', 1, 60, now(), now()),
(3, 'avulsa_30', 'Sessão avulsa de 30 minutos', 1, 30, now(), now()),
(4, 'avulsa_60', 'Sessão avulsa de 60 minutos', 1, 60, now(), now()),

(2, 'semanal_30', 'Plano com sessões semanais de 30 minutos', 2, 30, now(), now()),
(3, 'semanal_60', 'Plano com sessões semanais de 60 minutos', 2, 60, now(), now()),
(4, 'quinzenal_30', 'Plano com sessões quinzenais de 30 minutos', 3, 30, now(), now()),
(5, 'quinzenal_60', 'Plano com sessões quinzenais de 60 minutos', 3, 60, now(), now()),


create table if not exists customer_plan (
    id bigserial primary key,
    customer_id bigint not null references billing.customer(id) on delete cascade,
    plan_id bigint not null references plan(id) on delete cascade,
    start_date date not null,
    end_date date,
    created_at timestamp not null,
    updated_at timestamp not null
);

create table if not exists session (
    id bigserial primary key,
    customer_plan_id bigint not null references customer_plan(id) on delete cascade,
    start_date date not null,
    start_time time not null,
    session_status_id bigint not null references session_status(id),
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null,
    constraint unique_session_customer_start_date_start_time unique(customer_plan_id, start_date, start_time)
);
