## Planejamento 

### 0. Manutenções
* Revisar log billing e planner


No subsistema planner:
- Criar serviço de consulta das invoices (plan/invoicing/list), conforme arquitetura e criterios dos demais serviços de consulta:
    + Os parametros de entrada (dto) são:
        * page: int (não obrigatório, padrão 1)
        * page_size: int (não obrigatório, padrão 10)
        * start_date: date (não obrigatório)
        * end_date: date (não obrigatório)
        * invoice_id: int64 (não obrigatório)
        * nickname: customer nickname (não obrigatório)
        * plan_type: int (não obrigatório)
        * service_id: int (não obrigatório)
        * plan_id: int (não obrigatório)
    + A ordem deve ser por invoice_date descendente e invoice_id descendente
    + O dto de saída é uma lista paginada de invoices (não deletadas deleted_at != null), com os seguintes campos:
        * id: int64
        * invoice_date: date
        * nickname (customer nickname)
        * due_date
        * amount
        * items (caixa descrevendo os itens da invoice):
            * id: int64
            * service_id: int64
            * description: string
            * price: float64
            * invoiced_quantity: int
            * realized_quantity: int
- Colocar a api no handler de invoices (plan/invoicing/invoices)
- Fazer o htmx e adicionar ao handler_html no mesmo modelo e critérios que as demais htmx de consulta como plans.html e sessions.html (criar o invoicing.hmtl)....colocar no menu lateral como invoicing (como Pré-Faturas)

No subprojeto vendor:
- adicionar o icone invoicing no sidebar como Pré-Faturas, apontando para o sub-projeto correspondente




Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    