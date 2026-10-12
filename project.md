## Planejamento 

### 0. Manutenções
* Revisar log billing e planner


No subsistema planner, nas funcionalidades de gestão de planos (criar, alterar, listar), existe uma variável que determina a estratégia de armazenamento chamada tipo de plano (1: plan_agenda, 2: plan_package, 3: plan_notebook). No plano plan_notebook existe duas variáveis que são exclusivas (payment_day e payment_term) que estão armazenadas no campo plan_notebook.payment_day e plan_notebook.payment_term. Preciso reestruturar de forma que exista um quarto tipo de plano chamado pay-per-view (4), que desta forma tem uma tabela exlusiva plan_payperview e assume o papel das variáveis payment_term, enquanto o tipo 3 (notebook - caderneta) fica com a tabela especialista plan_notebook com o campo payment_day. O objetivo é que a lógica de faturamento seja a mesma (criacao de invoice_item e invoice), mas que as estratégias sejam diferentes para cada tipo de plano. Preciso aljustar todo o fluxo dos serviços (htmx, handlers, service, dto e adapters) para contemplar esta mudança. Na base de dev já apliquei os ajustes no banco de dados para conter o tipo 4 e a tabela plan_payperview...um detalhe é que pode manter o nome pay-per-view o mesmo que em ingles.



No subsistema planner:
- Criar um serviço que gera as faturas (invoicing) com base nos planos cadastrados de acordo com a arquitetura vigente:
    * O serviço recebe via dto: 
        - vendor_id (opcional) - se não informado pegar todos
        - data de inicio (opcional) - caso não informado, condiderar n meses antes do inicio do mes atual 
        - data fim (opcional) - caso não informado, considerar o fim do mes atual
        - tipo do plano (opcional) - se não informado considerar todos
        - tipo de serviço (opcional) - se não informado considerar todos
        - cliente_id (opcional) - se não informado considerar todos os clientes do vendor
        - reprocess (opcional) - se true, reprocesar faturas existentes - default false
    * O serviço deve retornar apenas o dto padrão de retorno (base), informando o resultado (ok ou erro), de acordo com os demais serviços. 
    * O serviço deve executar todos os planos encontrados que estavam vigentes dentro o periodo informado e também considerar os demais filtros que não foram enviados nulos no dto: vendor_id, tipo_plano, tipo_servico, cliente_id  
    * Caso a opção reprocess for verdadeira e a fatura existir, ele deve ser deletada (soft delete - valor em deleted_at) antes de ser gerada novamente.
    * para cada plano encontrado deve-se analisar o tipo de plano e estabelecer estratégias diferentes para cada tipo (agenda, pacote e caderneta):
        - para tipo plano pacote:
            - o faturamento desta modalidade deve ocorrer uma única vez por plano, desta forma, verificar se não existe fatura com o mesmo plano id e que não esteja deletada (deleted_at) e caso tenha logar e não criar a fatura.
            - Se a fatura ainda não for criada, criar a fatura:
                + Criar invoice_item: Para cadas plan_item criar um invoice_item correspondente, com a seguinte correspondência de valores:
                    invoice_item.invoice_id = invoice.id
                    invoice_item.service_id = plan_item.service_id
                    invoice_item.price = plan_item.price caso plan_item é não nulo caso contrário 0
                    invoice_item.description = plan_item.service.description
                    invoice_item.invoiced_quantity = 'pegar plan.plan_package.quantity e distribuir proporcionalmente entre os plan_items pelo order_index'
                    invoice_item.realized_quantity = 0
                    invoice_item.created_at = now()
                    invoice_item.updated_at = now()
                    invoice_item.deleted_at = null
                + Criar invoice com a seguinte correspondência:
                    invoice.plan_id = plan.id
                    invoice.invoice_date = plan.plan_start
                    invoice.due_date = max(plan.plan_package.payment_date, current_date)
                    invoice.amount = sum of plan_item.price * plan_item.invoiced_quantity se plan.price é nulo senão plan.price
                    invoice.reference_invoice_id = null
                    invoice.created_at = now()
                    invoice.updated_at = now()
                    invoice.deleted_at = null
        - plano caderneta:
            - O faturamento do plano caderneta é feito com base nas sessões do customer no mês anterior e pode ter duas estratégias distintas com base no plan_notebook.payment_day:
                - Caso o plan_notebook.payment_day é diferente de nulo, então deve-se gerar uma única fatura por mês, onde o faturamento ocorre uma vez por mês somando todas as sessões do mês anterior, dia de faturamento primeiro dia do mês corrente e com vencimento no dia do pagamento. 
                - Caso o plan_notebook.payment_day é nulo, então o faturamento deve ocorrer  
                


            - o faturamento deste plano pode ter duas estratégias dependendo do campo payment_day: 
                - se o campo payment_day não é nulo, então o faturamento deve ocorrer uma vez a cada mês dentro do período recebido como parâmetro;
                - se o campo payment_day é nulo, então o faturamento deve ocorrer
                


            -  O faturamento deste plano deve ocorrer uma vez por mês, desta forma, a cada mês dentro do período recebido como parâmetro, verificar que já não existe uma fatura para aquele mês, e caso tenha, logar e não criar a fatura.
            - O plano caderneta pode ter duas estratégias, dependendo do campo payment_day: 
                

            - Se a fatura ainda não for criada, criar a fatura:
                + Criar invoice_item: Para cadas plan_item criar um invoice_item correspondente, com a seguinte correspondência de valores:
                    invoice_item.invoice_id = invoice.id
                    invoice_item.service_id = plan_item.service_id
                    invoice_item.price = plan_item.price se plan_item price é não nulo, caso contrário 0
                    invoice_item.description = plan_item.service.description
                    invoice_item.invoiced_quantity = buscar em sessoes a quantidade de sessoes do mesmo serviço no mês anterior
                    invoice_item.realized_quantity = 0
                    invoice_item.created_at = now()
                    invoice_item.updated_at = now()
                    invoice_item.deleted_at = null
                + Criar invoice com a seguinte correspondência:
                    invoice.plan_id = plan.id
                    invoice.invoice_date = plan.plan_start
                    invoice.due_date = max(plan.plan_package.payment_date, current_date)
                    invoice.amount = sum of plan_item.price * plan_item.invoiced_quantity se plan.price é nulo senão plan.price
                    invoice.reference_invoice_id = null
                    invoice.created_at = now()
                    invoice.updated_at = now()
                    invoice.deleted_at = null




Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    