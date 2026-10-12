## Planejamento 

### 0. Manutenções
* Revisar log billing e planner

No subsistema planner:
- Criar um serviço que gera as faturas com base nos planos cadastrados de acordo com a arquitetura vigente:
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
        - para plano pacote:
            - o faturamento desta modalidade deve ocorrer uma única vez por plano, desta forma, verificar se não existe fatura com o mesmo plano id e que não esteja deletada (deleted_at) e caso tenha logar, não criar a fatura.
            - 
