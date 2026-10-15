## Planejamento 

### 0. Manutenções
* Revisar log billing e planner



Invoices errors/Improvements
- No invoicing, estratégia agenda, utilizar como data da invoice (invoice_date) o primeiro dia da data do mes que está se gerando ao inves do primeiro dia do mes da data atual.
- Ao gerar a invoice no html, o botão ver pré-faturas não esta redirecionando para pré-fatura


No subprojeto planner:
- No serviço de invoicing, na estratégia de pacote, na descrição do serviço, colocar service.description + ' - ' data_inicial + ' a ' + data_final...se data_inicial = data_final ou data_final = null, colocar somente data_inicial...datas no formato dd/mm/aaaa
- No serviço de invoicing, na estratégia payperview, na descrição do serviço, colocar service.description + ' - ' + data da sessão (dd/mm/aaaa)

Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    