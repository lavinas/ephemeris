## Planejamento 

### 0. Manutenções
* Revisar log billing e planner


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
            - O faturamento do plano caderneta é feito para um determinado mês com base nas sessões registradas do customer no mês anterior para os serviços deste plano, então deve-se buscar no dominio sessoes por plano id e customer id as sessões não canceladas do mês anterior ao mês que está sendo faturado.
            - Para o período passado como parâmetros de entrada (dto), deve-se considerar todos os meses dentro do período cujo plano esteja ativo
            - Verificar se já não há uma fatura gerada para este período para este plano e se já gerada e o reprocess é true, deve-se deletar a fatura existente. Caso contrário, logar e pular para o próximo mês.
            - Para cada mês dentro do período, deve-se buscar as sessões não canceladas do customer para os serviços deste plano no mês anterior ao mês que está sendo faturado.
            - Se a fatura ainda não for criada, criar a fatura:
                + Criar invoice_item: Para cadas plan_item criar um invoice_item correspondente, com a seguinte correspondência de valores:
                    invoice_item.invoice_id = invoice.id
                    invoice_item.service_id = plan_item.service_id
                    invoice_item.price = plan_item.price se plan_item price é não nulo, caso contrário 0
                    invoice_item.description = plan_item.service.description
                    invoice_item.invoiced_quantity = buscar em sessoes a quantidade de sessoes não canceladas do mesmo serviço e customer no mês anterior
                    invoice_item.realized_quantity = o mesmo que invoice_item.invociced_quantity
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
            - Adicionalmente deve-se inserir na tabela invoice_item_session um registro de ligação entre o item e a sessão
        - plano pay-per-view:
            - O faturamento deste plano deve ocorrer para cadas sessão do customer não deletada e do serviço correspondente dentro do período informado
            - Verificar se já não há uma fatura gerada para esta sessão e se já gerada e o reprocess é true, deve-se deletar a fatura existente. Caso contrário, logar e pular para o próximo mês.
            - Se a fatura ainda não for criada, criar a fatura:
                + Criar invoice_item: Para cadas plan_item criar um invoice_item correspondente, com a seguinte correspondência de valores:
                    invoice_item.invoice_id = invoice.id
                    invoice_item.service_id = plan_item.service_id
                    invoice_item.price = plan_item.price se plan_item price é não nulo, caso contrário 0
                    invoice_item.description = plan_item.service.description
                    invoice_item.invoiced_quantity = 1
                    invoice_item.realized_quantity = 1
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
                + Adicionalmente deve-se inserir na tabela invoice_item_session um registro de ligação entre o item e a sessão      
        - plano agenda
            - O faturamento deste plano deve ser feito uma vez por mês nos meses dentro do período informado e dentro das dastas de validade do plano.
            - Verificar se já não há uma fatura gerada para este mês para este plano e se já gerada e o reprocess é true, deve-se deletar a fatura existente. Caso contrário, logar e pular para o próximo mês.
            - A quantidade de sessões calculadas dependem das variaveis plan_agenda.recurrence, plan_agenda.week_day, plan_agenda.term_type, plan_agenda.monthly_service_limit, plan.plan_start, plan.plan_end da seguinte forma:
                + De acordo com plan_agenda.term_type: se for 1 (pre-paid) deve-se calcular com base no mes corrente ao processamento, senão 2 (post-paid) deve-se calcular com base no mes anterior ao processamento
                + Para o período especificado no sub-item anterior, calcular quantas sessoes haveriam com base nas variaveis plan_agenda.recurrence, plan_agenda.week_day: calcular quantos dias da semana tem no periodo especificado de acordo com o plan_agenda.recurrence: se for 1 (weekly) deve-se calcular quantos dias da semana existem no intervalor de 7 dias dentro do período especificado a partir da data de inicio e até a data fim....no caso de bi-weekly o mesmo porém no intervalo de 14 dias e mensal no intervalor de 1 mes. Ajustar para manter o dia da semana especificado.
                + Se o monthly_service_limit for diferente de null, deve-se limitar a quantidade de sessões calculadas pela expressão min(quantidade_calculada, plan_agenda.monthly_service_limit)
            - Se a fatura ainda não for criada, criar a fatura:
                + Criar invoice_item: Para cadas plan_item criar um invoice_item correspondente, com a seguinte correspondência de valores:
                    invoice_item.invoice_id = invoice.id
                    invoice_item.service_id = plan_item.service_id
                    invoice_item.price = plan_item.price se plan_item price é não nulo, caso contrário 0
                    invoice_item.description = plan_item.service.description
                    invoice_item.invoiced_quantity = calculo de quantidade especificado acima
                    invoice_item.realized_quantity = 0
                    invoice_item.created_at = now()
                    invoice_item.updated_at = now()
                    invoice_item.deleted_at = null
                + Criar invoice com a seguinte correspondência:
                    invoice.plan_id = plan.id
                    invoice.invoice_date = o primeiro dia do mes do periodo especificado ou a data de inicio, o que for menor
                    invoice.due_date = mes e ano do periodo de processamento porem o dia deve-se tirar de plan_agenda.payment_day
                    invoice.amount = sum of plan_item.price * plan_item.invoiced_quantity se plan.price é nulo senão plan.price
                    invoice.reference_invoice_id = null
                    invoice.created_at = now()
                    invoice.updated_at = now()
                    invoice.deleted_at = null
    * Modelar a arquitetura das diversas estratégias descritas acima dentro de adapters (cadas estrategia um adapter) dentro da pasta adapters/invoicing que já foi criada.
    



Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    