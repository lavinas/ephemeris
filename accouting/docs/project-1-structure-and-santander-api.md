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


Alguns ajustes devem ser feitos após as implementações:
- Retirar do código, as implementações que se remetem ao open finance brasil que foi feito o ultimo pedido, vou utilizar apenas o santander api por agora.
- Alterei a tabela statement_santander_api para comportar os dados que realmente retornam da chamada Open API - Contas Correntes, desta forma, deve-se alterar o dominio e serviços para comportar os dados de acordo com o retorno.
- Foi criado também um campo chamado position, onde deve ser armazenada a posição do registro na lista de registro de um determinado dia, assumindo o risco que o santander apresenta os registros em uma ordem específica, e que novos campos sempre são apresentados aos registros que já estavam na lista. Este campos será utilizado para a garantia de unicidade, conforme descrito abaixo.
- Alterei tambpem a tabela statament para que tenha os campos baseados no do santander. A proximas implementações de extratos irão se adapatar para este nomenclatura também.
- Foi criado também o campo key na tabela statement_santander_api, onde será colocado um hash para garantir a unicidade e deve ser:
    * key = sha256(credit_debit_type, transaction_name, historic_complement, amount, transaction_date, document_type)
    * Este campo pode se repetir e será usado para se calcular o hash principal da tabela statamente como descrito abaixo.
- Desta forma o mapeamento dos dados do sanatander para a tabela de statement é:
    * transaction_name -> transaction
    * historic_complement -> complement 
    * amount -> amount se debito_credit = CEDITO e amount * -1 se debito_credit = DEBITO 
    * transaction_date -> sdate
    * document_number -> document
- O campo checksum não existe mais, porem foi criado o campo key, onde deve ser calculado uma chave de unicidade para o registro para uma determinada conta (account)....cadas especialização da statement terá sua estratégia de calculo de unicidade....para o santander api teremos: 
    * A composição da chave deve ser calculada com base no campo key da tabela statement_santander_api, e mais a identificação da conta (account_id) e também da ocorrência da chave para um determinado dia.
    * Desta forma, o key deve ser calculado:
        statement.key = hash256(account_id, ocorrencia do statament_santander_api.key no dia que está sendo processado, statament_santander_api.key)
    * Por exemplo, se no dia que o extrato está sendo processado é a primeira ocorrência do statament_santander_api.key, então a ocorrência será 1. Se for a segunda ocorrência, então a ocorrência será 2, e assim por diante. 
- Registros que estão importados cujo key1 já está na tabela, devem ser ignorados na importação, garantindo assim a não duplicidade na importação
- Além disto, na tabela statament foi criado o campo balance, que é o saldo acumulado da conta e deve ser calculado da seguinte forma:
    - Pelo account_id, data da transação deve-se encontrar o ultimo registro (com id maior)
    - O valor do saldo é o valor do registro anterior mais o amount
    - Caso seja o primeiro registro deve-se pegar como saldo anterior, o saldo que está na tabela (initial_balance)
    - Caso tenha registros posteriores, deve-se atualizar o saldo dos registros
- Na tabela account também existe o campo initial_date, deve-se considerar apenas registros com data maior ou igual esta data
- Mudar o htmx e também a correspondente api (dto, seriços, etc) para conter os seguintes campos na tabela:
    - Data (sdate)
    - Transacão (transaction)
    - Complemento (complement)
    - Documento (document)
    - Valor (amount)
    - Saldo (balance)
- Continuar ordenando pela data decrescente e pelo id decrescente
- Os filtros continuam os mesmos, porém adicionar o filtro de todos/crédito/débito (dependendo do valor da transação)


Ações para as contas e regras de ação:
- Foram criadas as tabelas bank_account_action e bank_account_action_rule que armazenam ações que devem ser feitas ao processar os extratos das contas.
- A tabela bank_account_action contém a conta a qual a ação será aplicada e uma especificação de ação que é um inteiro:
    * Para cada ação deve-se ter um código que implementa a regra a ação. 
    * A primeira regra é filtrar transações (1), ou seja, desconsiderar a linha do extrato na importação cuja linha cai nas regras especificadas em bank_account_action_rule.
    * As demais regras serão implementadas futuramente, porém deve-se prever uma arquitetura que se deixe claro e isolado o código de cada regra.
- A tabela bank_account_action_rule contém as regras que serão aplicadas às transações.
    * A tabela contém os seguintes campos:
        - action_id: foreign key para bank_account_action. A relação deve ser de 1 para muitos (1 ação pode ter muitas regras)
        - field: número que identifica o campo da tabela statement que deverá ser analisado pela regra e podem ser:
             1 - date, 2 - transaction, 3 - complement, 4 - document, 5 - amount
        - operator - número que identifica o operador que deverá ser usado para analisar o campo. Podem ser:
             1 - equals, 2 - contains, 3 - starts_with, 4 - ends_with, 5 - regex, 6 - greater_than, 7 - less_than
        - value - texto que deverá ser a variável utilizado pelo operador para a análise
        - case_sensitive - booleano que indica se a comparação deve ser case sensitive
    * Pode-se ter um ou mais regras por ação, caso tenha mais que um, as regras combinam-se como 'and', ou seja, elas devem ser todas verdadeiras para que a ação seja aplicada.
- Devem-se ter os serviço de inclusão, edição e exclusão destas entidades, que devem ser implementados via api e via htmx
- No htmx deve-se ter um botão aplicado a conta selecionada abrindo assim um modal para edição de ações.
- Pode-se ter mais de uma bank_account_action para uma mesma conta, inclusive para a mesma ação, e neste, caso (mesma conta e mesma ação), a combinação das duas sobrepostas tem o mesmo efeito que um 'or' ou seja, deve ser ignorado o registro se qualquer uma das regras for verdadeira. 
- Deve-se implementar no serviço de sincronismo, tanto via job como via sincronimo no htmx o filtro dos campos referente as ações e regras cadastradas. 

Botão excluir no modal de sincronismo:
- Deve-se ter no modal de sincronimos dos extratos um botão para que, se clicado, ele faz com que o sincrimos exclua as linhas do extrato para os dias sincronizados antes de fazer o processamento.
- O default deve ser sempre false, isto é, deve-se clicar para deletar.
- O sincronismo via job é sempre desligado
- A opção para exclusão também deve constar na api rest da chamada da funcionalidade