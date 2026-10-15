## Planejamento 

### 0. Manutenções
* Revisar log billing e planner



Invoices errors/Improvements
- No invoicing, estratégia agenda, utilizar como data da invoice (invoice_date) o primeiro dia da data do mes que está se gerando ao inves do primeiro dia do mes da data atual.
- Ao gerar a invoice no html, o botão ver pré-faturas não esta redirecionando para pré-fatura


No subprojeto planner:
    - No htmx de pré-faturas, na tabela, acrescentar o campo tipo plano, após o cliente e antes do vencimento.
    - Utilizar a mesma formatação de tipo de plano que no htmx de gestão de planos


- Fazer um serviço de conciliação das sessões criadas com as pré faturas, prevendo as seguintes regras:
    - Deve-se encontrar ao item da pré-fatura correspondente a esta sessão, ou seja, do mesmo customer, do mesmo serviço, e que a data da sessão esteja dentro do período de competência da pré-fatura (data inicial e data final)
    - Deve-se verificar que a sessão já não está associada a uma pré-fatura, ou seja, que não existe uma entry em invoice_item_session com o session_id e o invoice_item_id correspondentes
    - Caso a pré-fatura seja do tipo agenda ou do tipo pacote ou tipo notebook, deve-se somar 1 na coluna de quantidade realizado do item da fatura e registrar a associação na tabale invoice_item_session
     
     


Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    