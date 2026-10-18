## Planejamento 

### 0. Manutenções
* Revisar log billing e planner
* Verificar que no caso do plano pacote o número de items seja menor que a quantidade de aulas


No subprojeto planner:
- Irei trocar o conceito de conciliação (reconciliate) para o conceito de banco de horas (comptime);
- Na base de dados atual de invoiced do planner (vide /planner/migration/invoiced.sql):
    * Remover a coluna realized_quantity
    * alterar a coluna invoiced_quantity para apenas quantity
    * desta forma não haverá mais conciliação de quantidades no domínio de faturamento, será criado um novo domínio chamado comptime para controlar o banco de horas
- Haverá a criação das tabelas comptime (vide /planner/migration/comptime.sql):
    * A tabela comptime será responsável por controlar o banco de horas.
    * A tabela comptime_session será responsável por controlar as sessões que foram utilizadas para somar ou abater o banco de horas.
    * A tabela comptime_session_invoice será responsável por controlar as faturas que foram utilizadas para somar ou abater o banco de horas.
- Para cada sessão criada para o mesmo customer + service, deverá ser computada a sessão no banco de horas:
    * Deve-se verificar se o comptime foi criado para aquele customer e para aquele service (se não deve-se criar com balance 0)
    * comptime.customer_id e comptime.service_id deverão ser iguais ao customer e ao service da sessão
    * comptime.balance deverá ser decrementado de acordo com a conversão do numero de minutos da sessão (session.session_minutes) para duration
    * Deve-se criar uma linha na tabela comptime_session para aquele comptime e para aquela sessão com balance_changed o valor decrementado no balance (o valor aqui pode ser sempre positivo porem decrementando comptime.balance), o ref_date é a data da sessão e o ref_month é o primeiro dia do mês da sessão
    * Todas as sessões devem ser contabilizadas, porém aquelas com o status cancelada_nao_cobrar deve deve ser contabilizada com valor 0 em balance_changed (e assim com efeito zero na comptime.balance)
    * Fazer as alterações na base de dados de forma atômica (com transação)
- Para cada sessão deletada para o mesmo customer + service, deverá ser computada a sessão no banco de horas:
    * Deve-se localizar comptime_session correspondente a aquela sessão
    * Deve-se somar do balance da comptime o valor de balance_changed
    * Deve-se apagar a linha na tabela comptime_session
    * Fazer as alterações na base de dados de forma atômica (com transação)
- Para cada sessão atualizada para o mesmo customer + service, deverá ser computada a sessão no banco de horas:
    * Deve-se localizar comptime_session correspondente a aquela sessão
    * Deve-se somar do balance da comptime o valor de balance_changed
    * Deve-se atualizar o valor do session_minutes na coluna balance_changed
    * Deve-se subtrair do balance o novo valor de balance_changed
    * Todas as sessões devem ser contabilizadas, porém aquelas com o status cancelada_nao_cobrar deve deve ser contabilizada com valor 0 em balance_changed (e assim com efeito zero na comptime.balance)
    * Fazer as alterações na base de dados de forma atômica (com transação)
- Para cada invoice_item criada para um determinado customer + service, deverá ser computada a fatura no banco de horas:
    * Deve-se verificar se o comptime foi criado para aquele customer e para aquele service (se não deve-se criar com balance 0)
    * comptime.customer_id e comptime.service_id deverão ser iguais ao customer_id da invoice pai e service_id da invoice_item
    * comptime.balance deverá ser somado de acordo com a conversão do numero de minutos da sessão (invoice_item.quantity * service.session_minutes) para duration
    * Deve-se criar uma linha na tabela comptime_invoice_item para aquele comptime e para aquela sessão com balance_changed o valor incrementado no balance (também sempre positivo) e o ref_date é invoice.period_start e o ref_month é o primeiro dia 
 - Para cada invoice_item ou invoice pai deletadas para o mesmo customer + service, deverá ser computada a fatura no banco de horas:
    * Deve-se localizar comptime_invoice_item correspondente a aquela invoice_item ou invoice pai
    * Deve-se subtrair do balance da comptime o valor de balance_changed
    * Deve-se apagar a linha na tabela comptime_invoice_item
    * Fazer as alterações na base de dados de forma atômica (com transação)
 - Para cada invoice_item ou invoice pai atualizada para o mesmo customer + service, deverá ser computada a fatura no banco de horas:
    * Deve-se localizar comptime_invoice_item correspondente a aquela invoice_item ou invoice pai
    * Deve-se subtrair do balance da comptime o valor de balance_changed
    * Deve-se atualizar o valor do session_minutes na coluna balance_changed
    * Deve-se somar do balance o novo valor (invoice_item.quantity * service.session_minutes) de balance_changed
    * Fazer as alterações na base de dados de forma atômica (com transação)
- Deve-se criar e alterar adequadamente os domain correspondentes;
- Deve-se eliminar o serviço anterior reconciliation.go e criar o novo serviço como comptime.go;
- Deve-se criar um htmx e hadler_html (bem como serviços, dtos e apis correspondentes) para gestão do comptime:
    * todo
    
    
    
    

    
    








                    