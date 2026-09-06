## Planejamento 

### 0. Manutenções
* Ajustar horário do servidor
* Colocar backend billing em produção
* Fazer log por dia
* Script para fazer backup automático do banco e do log pata billing e para o planner 

### 1. Incluir Criar cliente e serviço no billing e planner

#### Neste momento o Planner irá buscar e cadastrar usuários e serviços por um adaptador que acesserá api de serviços externos (neste caso apis do bliling)...em um segundo momento pode-se pensar e criar uma base local e fazer o sincronismo de bases via fila:

1. Sincronizar usuários billing e planner
* Ajustar htmx de sessões do planner para conter o botão e o serviço de inclusção e alteração de usuários
* Criar adapter no Planner para crude de usuários
* Criar serviço no Planner para crude de usuários
* Incluir a edição de usuários no adapter api
* Incluir a edição de usuários no adapter htmx
* Alterar crude de sessões para listar apenas usuários cadastrados vindos do adapter 

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
