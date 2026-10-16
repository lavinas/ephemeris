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
- No htmx de gestão de planos e no handler_htmx, colocar um botão 'Gerar pré-faturas' entre 'adicionar' e 'limpar filtros' que chama o serviço de 
invoicing, de acordo com  os filtros ativos do momento.
- Ainda no htmx de gestão de planos, no pop-up de confirmação de geração de pré-faturas individual (Processamento Concluído), o bostão "Ver Pré faturas" deveria direcionar para a tela de pré-faturas, porém apenas está fechando o pop-up
     


Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    