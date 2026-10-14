## Planejamento 

### 0. Manutenções
* Revisar log billing e planner


No subsistema planner:
- Para o htmx de planos, colocar na tabela de planos mais um action para a geração da invoice correspondente


No subprojeto vendor:
- na api e serviço invoicing, receber no dto como parametro o plan_id e considerar este parametro na execução do serviço.
- no htmx de planos, prever mais uma ação na listagem de planos que permita a geração da invoice correspondente ao plano, abrindo um modal para que seja informada as datas de inicio e fim.
- fazer um serviço completo (dto, service, etc) para deletar (soft-delete) uma invoice, recebendo o id da invoice e atualizando a data de delete da base.
- no htmx de invoicing (pré fatura), colocar um action para apagar a fatura perguntando antes se deseja apagar




Outros:
    * Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas
                    
                    

                    