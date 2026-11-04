Integração com o Santander API no subprojeto accouning:
- Ajustar a arquitetura para uma estrutura hexagonal baseada nos demais subprojetos (planner, etc):
    * Logs, config, banco de dados etc localizados no internal/adapter
    * Entidades e Regras de negócio em internal/domain
    * Lógica de negócio em internal/service
    * Interfaces do serviço em internal/dto
    * Interfaces das aplicações externas em internal/port
- As Entidades previstas basear-se na base de dados que está em migration
- Em migration, temos as seguintes tabelas:
    * vendor: Empresas que utilizarão o sistema
    * account: Contas bancárias dos vendors
    * account_santander_api: Tabela especializada de dados de acesso de api santander no caso da conta ser do santander api
    * statement: dados base de movimentação bacária
    * statement_santander_api: Movimentações bancárias obtidas via Santander API
- Prever que esta forma de obter os dados não é a única, por exemplo para algumas outras contas não teremos acesso a api, então teremos outras formas de obter esses dados, como por exemplo subir um arqyuivo ofx via backend.
- Mas caso se tenha os dados na tabela account_santander_api, então esta será a forma de busca dos dados.
- Deixar de forma ortogonal e separados as classes de códigos para buscar os dados.
- As classes além de buscar os dados, devem persistir na base os dados correspondente (no caso do santander api: statement_santander_api) e depois fazer um de-para para statement.
- Na statement o valor pode ser positivo e negativo, de acordo com o que veio nos dados do santander (debito/credito, valor)
- Um pdf com a descrição da api de consulta de extrato do santander api está em /docs/extrato-santander.pdf, caso não tenha o algortmo já mapeada no agente.
- As apis podem ser em ambiente de sandbox ou produção no Santander api, desta forma, deixar em arquivo de configuração as urls e parametros necessários de acesso.
- A ideia é que se busque os dados das apis de tempos em tempos, configuração esta em arquivo de configuração no modelo cron.
- Para controle de duplicidade dos dados via multiplos extratos, criei a coluna checksum na tabela statement. Este checksum é feito com base em algum hash dos campos da tabela (account_id, ddate, amount, memo, partie).
- Também para o caso específico do Santander api, deve-se verificar a duplicidade através do id de transação único que vem no extrato e está mapeado no campo transaction_id.
- No caso de duplicidade, ignorar o registro duplicado no processamento do extrato.
- A importação dos dados do Santander via api deve retroagir a n dias da data atual (com n configuravel em arquivo de configuração).
- O job de importação dos dados dpo santander api deve buscar todas as contas com status ativo e não deletadas e fazer o processamento a princípio de forma sequencial.
- Criar serviço de consulta do extrato via api e htmx para que o usuário possa consultar o extrato de uma conta.
- Basear-se na estrutura de handlers de api e htmx do subprojeto planner
- No htmx colocar os filtros pela data e por conta (deixar como default a primeira conta encontrada)
- É necessáio sempre ter uma conta para consulta
- As informações retornadas na tabela são da tabela statement
    * data (sdate), descrição (memo), valor (amount), contraparte (partie)
    * ordenar pela data decrencente, e depois pelo id decrescente também
- Fazer a paginação conforme os demais serviços de outros subprojetos
- Propor alguma melhoria ou alteração caso seja necessário, sempre utilizando as boas práticas solid e clean code, etc.