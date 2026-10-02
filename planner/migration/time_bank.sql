-------------------------------------------------------
-- 4. Banco de Horas / Créditos de Sessões
-------------------------------------------------------
drop table if exists time_bank cascade;
create table time_bank (
    id bigserial primary key,
    customer_id bigint not null references customer(id) on delete cascade,
    offer_id bigint null references offer(id) on delete set null,
    reference_month date not null,      -- Ex: '2026-10-01' (mês de competência da divergência)
    minutes_balance int not null,       -- Positivo: crédito (+60 min), Negativo: débito/consumo (-60 min)
    sessions_balance int not null,      -- Positivo: +1 aula, Negativo: -1 aula
    entry_type varchar(50) not null,    -- 'confronto_mensal', 'ajuste_manual', 'consumo_credito'
    notes text,                         -- Motivo do ajuste (ex: "Aluno faltou aula dia 15/10 com aviso prévio")
    created_at timestamp not null default now()
);

