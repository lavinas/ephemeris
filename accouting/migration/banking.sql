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

create table account (
    id bigserial primary key,
    vendor_id bigint not null,
    bank_name varchar(255) not null,
    bank_code varchar(10) not null,
    account_number varchar(50) not null,
    account_agency varchar(20) not null,
    pix_key varchar(50),
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    CONSTRAINT fk_account_vendor foreign key (vendor_id) references vendor(id),
    CONSTRAINT uk_account_vendor_agency_number unique(vendor_id, account_agency, account_number)
);

create index idx_account_vendor_id on account(vendor_id);


create table statement (
    id bigserial primary key,
    account_id bigint not null,
    
    




    
)