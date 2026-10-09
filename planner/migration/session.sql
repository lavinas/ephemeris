-- Active: 1790470840125@@192.168.1.138@5433@planner@planner
create schema if not exists planner;

set search_path to planner;

# vendor table
drop table if exists vendor cascade;
create table vendor (
    id bigserial primary key,
    nickname varchar(150),
    legal_name varchar(150) not null,
    trading_name varchar(150),
    document varchar(50) not null,
    email VARCHAR(150) not null,
    whatsapp varchar(20) not null,
    tax_document varchar(50) not null,
    account_bank varchar(100) not null,
    account_agency varchar(20) not null,
    account_number varchar(50) not null,
    pix_token varchar(255)not null,
    pix_name varchar(255) not null,
    pix_city varchar(255) not null,
    logo_name varchar(255),
    created_at timestamp not null,
    updated_at timestamp not null,
    last_rps bigint not null,
    smtp_host varchar(255),
    smtp_port int,
    smtp_user varchar(255),
    smtp_password varchar(255),
    constraint unique_vendor_document unique(document),
    constraint unique_vendor_nickname unique(nickname)
);

# main vendor
insert into vendor (nickname, legal_name, trading_name, document, tax_document, account_bank, account_agency, account_number, pix_token, pix_name, pix_city, logo_name, email, whatsapp, last_rps, smtp_host, smtp_port, smtp_user, smtp_password, created_at, updated_at) values
('estudio_amelia', 'Cardoso e Barbosa Serviços Musicais e Tecnologia LTDA', 'Estúdio Amélia Cardoso', '27.928.875/0001-04', '5.727.888-1', '033 - Santander', '0985', '13001001-4', '27.928.875/0001-04', 'Estúdio Vocal Amélia Cardoso', 'São Paulo', 'logo_amelia.png', 'financeiro@ameliacardoso.com.br', '(11) 98088-8399', 2435, 'smtp.zoho.com', 465, 'financeiro@ameliacardoso.com.br', 'pwd22Adm**', now(), now());

# customer table
drop table if exists customer cascade;
create table customer (
    id bigserial primary key,
    name varchar(150) not null,
    vendor_id bigint not null references vendor(id) on delete cascade,
    nickname varchar(150) not null,
    document varchar(50),
    email varchar(150),
    whatsapp varchar(20),
    created_at timestamp not null,
    updated_at timestamp not null,
    status int not null default 1,
    constraint unique_customer_document unique(vendor_id, document),
    constraint unique_customer_nickname unique(vendor_id, nickname)
);

drop table service;

# service
create table service (
    id bigserial primary key,
    vendor_id bigint not null references vendor(id) on delete cascade,
    name varchar(150) not null,
    description text,
    session_minutes int null, -- if there is a predicted duration for the service
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

# insert default services for vendor_id 1

select * from service;

insert into service (vendor_id, name, description, session_minutes, created_at, updated_at) values 
(1, 'online/canto/30', 'Aula de canto online de 30 minutos', 30, now(), now()),
(1, 'online/canto/45', 'Aula de canto online de 45 minutos', 45, now(), now()),
(1, 'online/canto/60', 'Aula de canto online de 60 minutos', 60, now(), now()),
(1, 'online/piano/30', 'Aula de piano online de 30 minutos', 30, now(), now()),
(1, 'online/piano/45', 'Aula de piano online de 45 minutos', 45, now(), now()),
(1, 'online/piano/60', 'Aula de piano online de 60 minutos', 60, now(), now());

create table tmp_session as
select * from session;

select count(1) from tmp_session

select * from tmp_session;

# session new
drop table if exists session;
create table session (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    session_date date not null,
    session_service varchar(100) not null,
    session_minutes int not null,
    -- session_service varchar(100) not null,
    session_status varchar(50) not null, -- realizada, cancelada_cobrar, cancelada_nao_cobrar 
    comments text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    constraint fk_customer_id foreign key (customer_id) references customer(id) on delete cascade
);

select * from tmp_session;

insert into session
select a.id, a.customer_id, b.id, a.session_date, a.session_service, a.session_minutes, a.session_status, a.comments, a.created_at, a.updated_at, a.deleted_at 
  from tmp_session a
    inner join service b on b.name = concat(replace(a.session_service, 'aula', 'online'), '/', a.session_minutes)

select count(1)
  from session;

select * from session;

commit;

select count(1)
  from tmp_session a
  left join service b on b.name = concat(replace(a.session_service, 'aula', 'online'), '/', a.session_minutes)
  where b.id is null;

# contract
create table contract (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    contract_start date not null,
    contract_end date null,
    contract_type int not null check(contract_type in (1,2,3)), -- 1: monthly agenda (agenda), 2: package (pacote), 3: notebook (caderneta)
    price numeric(15, 2) null, -- if not null, then override the item prices
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

# contract_item 
create table contract_item (
    id bigserial primary key,
    contract_id bigint not null references contract(id) on delete cascade,
    service_id bigint not null references service(id) on delete cascade,
    price numeric(15, 2) null, -- if null, use service price
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

# agenda contract
create table contract_agenda (
    id bigserial primary key references contract(id) on delete cascade,
    recurrence int not null check (recurrence in (1, 2, 3)), -- 1: weekly, 2: bi-weekly, 3: monthly
    week_day int not null check (week_day in (1, 2, 3, 4, 5, 6, 7)), -- 1: sunday, 2: monday, 3: tuesday, 4: wednesday, 5: thursday, 6: friday, 7: saturday
    term_type int not null check (term_type in (1, 2)), -- 1 - pre-paid, 2 - post-paid
    payment_day int not null check (payment_day between 1 and 31), -- day of the month to charge (1-31)
    monthly_service_limit int null check (monthly_service_limit > 0), -- limit number of sessions per month, if null, no limit,
);

# contract_package
create table contract_package (
    id bigserial primary key references contract(id) on delete cascade,
    service_quantity int not null check (service_quantity > 0), -- quantity of sessions
    payment_date date not null 
);

# notebook contract
create table contract_notebook (
    id bigserial primary key references contract(id) on delete cascade,
    session_payment_day int null check (session_payment_day between 1 and 31),
    session_payment_term int null check (session_payment_term > 0),
    constraint check_payment_day_or_term_exclusive 
        check (num_nonnulls(session_payment_day, session_payment_term) = 1)
);



    
