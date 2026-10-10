## Planejamento 

### 0. Manutenções
* Revisar log billing e planner

### 1. Criação da invoice 
* Decisão da arquitetura (colocar em billing ou novo serviço)
* Criação do modelo de planos



- No subsistema planner, criar a estrutura de criação, listagem, atualização e deleção de planos, baseado na arquitetura de outras funcionalidades existentes.
- Fazer baseado na estrutura de tabelas que estão em migration/plan.sql.
- Criar os domímions, serviços, dtos, handlers e atualizar os adapters de forma coerente com a arquitetura existente.
- As seguintes premissas foram pensadas na geração da estrutura:
    * Esta estrutura será utilizada para posterioemente ser a base para os algortmos de geração de faturas
    * O plano deve estar vinculado a um customer (utilizar nickname para vincular) e uma lista de serviços (items)
    * Alem disto o plano terá uma data inicio (Obrigatório) e data final (não obrigatório, ou seja, o plano está em andamento sem um fim previsto) e também deve ser de algum tipo (plano de agenda, plano de pacote ou plano caderneta).
    * Para cada tipo de plano existirão atributos extras (vide migration/plan.sql)
    * Algumas validações necessárias para inclusão, atualização, etc:
        1. O cliente somente poderá ter um tipo de serviço ativo de acordo com a coinsidencia de períodos de inicio e fim, ou seja, dentro do período de uma plano de um cliente para um determinado serviço, não se pode cadastrar ou atualizar outro plano com mesmo cliente, serviço...é possível cadastrar mais de um período/serviço caso os períodos não coincidam....considerar data fim infinito caso seja nulo.
        2. Todos os planos devem ter um tipo e um registro na tabela especializada correspondente.
        3. O tipo caderneta (plan_notebook) deve ter uma e apenas uma das duas variáveis com valor (a outra deve ser nula) e não deve ter as duas nulas.
- Fazer também o htmx e os handler_html, baseado na estrutura de sessoes, porém considerando as regras da base de planos 
- Expor também as apis de cadastro, atualização, listagem e deleção baseado no que já existe em sessões 