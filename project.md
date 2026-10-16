## Planejamento 

### 0. Manutenções
* Revisar log billing e planner



Invoices errors/Improvements
- No invoicing, estratégia agenda, utilizar como data da invoice (invoice_date) o primeiro dia da data do mes que está se gerando ao inves do primeiro dia do mes da data atual.
- Ao gerar a invoice no html, o botão ver pré-faturas não esta redirecionando para pré-fatura


No subprojeto planner:
- Na tabela invoice, criei os campos period_start date e period_end date para armazenar o periodo considerado no faturamento, ou seja, o periodo de calculo das sessoes consideradas em seus itens;
- Da mesma forma que foram calculados as quantidades de itens;
    - na estrategia agenda - recorrencia pre-paga - o periodo é o primeiro e ultimo mes em processamento corrente
    - na estrategia agenda - recorrencia pos-paga - o periodo é o primeiro e ultimo mes anterior do mes em processamento corrente
    - na estrategia notebook (caderneta) - o periodo é o primeiro e ultimo mes anterior do mes em processamento corrente
    - na estrategia pay-per-view - o periodo é o dia da sessão no inicio e fim do período
    - na estrategia pacote - o período é o período é o período do plan_package - ou seja, o período é do dia do início do plano até o dia do fim do plano
- Adicionalmente no caso de cadastro de planos pacote (e somente do pacote), deixar obrigatório e fazer a validação de que a data inicio e data fim são obrigatórias



No subprojeto planner:
- Fazer um serviço de conciliação das sessões criadas com as pré faturas, prevendo as seguintes regras:
    - Deve-se encontrar ao item da pré-fatura correspondente a esta sessão, ou seja, do mesmo customer, do mesmo serviço, e que a data da sessão esteja dentro do período de competência da pré-fatura (data inicial e data final)
    - Deve-se verificar que a sessão já não está associada a uma pré-fatura, ou seja, que não existe uma entry em invoice_item_session com o session_id e o invoice_item_id correspondentes
    - Caso a pré-fatura seja do tipo agenda ou do tipo pacote ou tipo notebook, deve-se somar 1 na coluna de quantidade realizado do item da fatura e registrar a associação na tabale invoice_item_session
    - No caso de pay-per-view, a associação já deve criar a pré-fatura, de novo, caso já não exista o plano e o registro na tabela de associação.
    - A conciliação deve ser disparada a cada criação de sessão ou a cada criação de uma nova pré-fatura, ou seja, no caso da criação da sessão, deve-se buscar a pré-fatura e seus itens e no caso da criação da pré-fatura deve-se buscar as sessões de mesmo customer e serviço dentro do período de inicio e fim de cada item de pré-fatura
    - Caso a sessão seja deletada, deve-se fazer o mesmo, porém subtrindo 1 na coluna de quantidade realizado do item da fatura e deletando o registro na tabela de associação.
Utilizar a melhor arquitetura baseada no hexagonal e nos bons principios de DDD
Apresentar a arquitetura a ser utilizada antes de aplica-la por favor

     


Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    