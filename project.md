## Planejamento 

### 0. Manutenções
* Revisar log billing e planner



Invoices errors/Improvements
- No invoicing, estratégia agenda, utilizar como data da invoice (invoice_date) o primeiro dia da data do mes que está se gerando ao inves do primeiro dia do mes da data atual.
- Ao gerar a invoice no html, o botão ver pré-faturas não esta redirecionando para pré-fatura


No subprojeto planner:
- Existe um erro conceitual na geração das invoices da estrategia de agenda:
    - A dada da invoice e data de vencimento são sempre de acordo com o mês corrente do calculo tanto pra pré-pago como pós-pago
    - A diferença entre pré-pago e pós-pago é somente que no pós-pago é calculado o número de sessões previstas do mês anterior e no pré-pago do mes atual
    - A validação se a invoice já existe portanto é com base na data da invoice e não na data de vencimento 


Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    