# 1. Verificar tabelas com muitas tuplas mortas e eficiência do HOT
# Esta query lista quais tabelas estão acumulando tuplas pendentes de limpeza (dead tuples) e qual porcentagem dos UPDATEs conseguiram usar HOT (quanto mais perto de 100%, melhor):
# O que observar:
#   Se pct_mortas for maior que 15% a 20%, o autovacuum pode estar atrasado nessa tabela.
#   Se pct_hot_updates for muito baixo em tabelas que sofrem muitos updates, os updates estão tendo que escrever em índices e alocar novos blocos.

SELECT 
    schemaname,
    relname AS tabela,
    n_live_tup AS tuplas_vivas,
    n_dead_tup AS tuplas_mortas,
    ROUND(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS pct_mortas,
    n_tup_upd AS total_updates,
    n_tup_hot_upd AS hot_updates,
    CASE 
        WHEN n_tup_upd > 0 THEN ROUND(n_tup_hot_upd * 100.0 / n_tup_upd, 2)
        ELSE 0 
    END AS pct_hot_updates,
    last_vacuum,
    last_autovacuum
FROM pg_stat_user_tables
WHERE schemaname = 'billing'
ORDER BY n_dead_tup DESC;


# 2. Verificar transações "presas" que impedem o HOT de ser limpo
# Cadeias HOT só ficam acumuladas e não são limpas pelo PostgreSQL se houver transações antigas abertas (ex: conexões em estado idle in transaction). Enquanto elas existirem, o PostgreSQL é proibido de apagar versões antigas:
# O que observar: 
#   Se você vir conexões com duracao_transacao de várias horas ou dias com estado idle in transaction, elas estão congelando a limpeza de HOT chains no banco inteiro.

SELECT 
    pid,
    usename,
    client_addr,
    state,
    now() - xact_start AS duracao_transacao,
    query
FROM pg_stat_activity
WHERE (now() - xact_start) > INTERVAL '5 minutes'
  AND state != 'idle'
ORDER BY duracao_transacao DESC;


# 3. Inspecionar o tamanho das cadeias HOT em uma página (pageinspect)
# Se você desconfiar que uma tabela específica está com cadeias HOT muito longas dentro de uma página (bloco), podemos usar a extensão pageinspect para auditar:
# Se você vir muitos lp_flags = 2 (redirect) seguidos de vários itens com is_heap_only = true, significa que essa página passou por uma sequência densa de atualizações na mesma linha.

CREATE EXTENSION IF NOT EXISTS pageinspect;
-- Inspeciona o bloco 1 da tabela invoice procurando redirecionamentos e tuplas HOT
SELECT 
    lp, 
    lp_flags,
    CASE lp_flags
        WHEN 0 THEN 'unused'
        WHEN 1 THEN 'normal'
        WHEN 2 THEN 'redirect (HOT root)'
        WHEN 3 THEN 'dead'
    END AS tipo_ponteiro,
    t_ctid,
    (t_infomask2 & 8192) != 0 AS is_heap_only
FROM heap_page_items(get_raw_page('billing.invoice', 1));

# 4. Verificar se há corrupções silenciosas nos índices (amcheck)
# Para garantir preventivamente que nenhum outro índice do banco esteja com ponteiros corrompidos ou quebrados, o PostgreSQL tem uma extensão oficial chamada amcheck:
# (Se qualquer índice tiver problemas de consistência interna, o comando vai acusar com erro; se rodar e não retornar nada além de linhas normais, todos os índices estão íntegros).

CREATE EXTENSION IF NOT EXISTS amcheck;
-- Faz uma checagem de integridade B-Tree em todos os índices do esquema billing
SELECT 
    c.relname AS indice,
    bt_index_check(c.oid, true) -- se retornar vazio (null), o índice está 100% íntegro!
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'billing'
  AND c.relam = 403; -- 403 = índices do tipo btree