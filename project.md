## Planejamento 

### 0. Manutenções
* Ajustar horário do servidor - OK - 07/09/2026
* Colocar backend billing em produção - OK - 07/09/2026
* Fazer bakup automatico de banco para billing e planner - OK - 07/09/2026
* Classificar sessoes por dia desc e id desc - OK - 09/09/2026
* Integrar o antigravitty no vscode de forma otima - OK - 09/09/2026
* Fazer log por dia - OK - 09/09/2026
* Revisar log billing e planner
* Script para fazer backup automático de log pata billing e para o planner 
* Criar past web no billing e colocar as pastas images e templates (como no planner)

### 0.5 - Restruturar front-end para colocar um frame em conjunto com todos os serviços:
* Colocar um frame teste no planner que pega o subframe sessoes - OK - 13/09/2026
* Trocar a porta do billing para 8081 - OK - 13/09/2026
* Colocar o frame teste e o billing 80-81 em produção
* Criar um projeto front-service (pensar no nome)
* Colocar o frame no projeto externo apontando para planner e billing
* Ajustar a pasta de templates/static do billing para fiocar igual o planner

### 1. Incluir Criar cliente e serviço no billing e planner

#### Neste momento o Planner irá buscar e cadastrar usuários e serviços por um adaptador que acesserá api de serviços externos (neste caso apis do bliling)...em um segundo momento pode-se pensar e criar uma base local e fazer o sincronismo de bases via fila:

1. Sincronizar usuários billing e planner
* Ajustar estrutura do billing para ficar igual ao planner
* Criar o frame de cadastro de gestão de usuários (customer) no billing
* Criar e expor html no no billing integrando aos serviços de cadastro
* Colocar a opção usuários no frame externo 

2. Criar e sincronizar serviços entre billing e planner
* Criar o domain de serviços
* Criar os serviços de serviços
* Criar a Api de serviços
* Aproveitar para ajustar o adapter http do billing para ficar proximo ao Planner
* Criar o adapter de busca de serviços externos no planner com acesso as apis billing
* Criar os serviços de crude de serviços
* Criar adapter no Planner para crude de usuários
* Criar serviço no Planner para crude de usuários
* Incluir a edição de usuários no adapter api
* Incluir a edição de usuários no adapter htmx
* Alterar crude de sessões para listar apenas usuários cadastrados vindos do adapter

### 2.Criar a geração automática de faturas 

* Criar base de dados 
* Criar os crudes
* Criar o agente automático


### 3. Conciliação Planner x Billing
* Criar base de dados
* Criar os crudes
* Criar os serviços api e htmx
