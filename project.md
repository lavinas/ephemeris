## Planejamento 

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
    * Deve-se verificar se o comptime foi criado para aquele customer e para aquele service (se não deve-se criar com valores invoiced_quantity, consumed_minutes, balance = 0)
    * Também deve-se vericar se já não existe o id da sessão na comptime_session com deleted_at = nil (para evitar duplicidade)
    * comptime.customer_id e comptime.service_id deverão ser iguais ao customer e ao service da sessão
    * comptime.consumed_minutes deve ser incrementado em session.session_minutes
    * Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
    * Deve-se criar uma linha na tabela comptime_session para aquele comptime e para aquela sessão com:
        - comptime_id = comptime.id
        - session_id = session.id
        - month = ano-mes-1 do session_date
        - consumed_minutes = session.session_minutes
    * Todas as sessões devem ser contabilizadas, porém aquelas com o status cancelada_nao_cobrar deve deve ser contabilizada com valor 0 em balance_changed (e assim com efeito zero na comptime.balance)
- Para cada sessão deletada para o mesmo customer + service, deverá ser computada a sessão no banco de horas:
    * Deve-se localizar comptime_session correspondente a aquela sessão
    * Deve-se subtrair o comptime.consumed_minutes pelo comptime_session.consumed_minutes
    * Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
    * Deve-se apagar a linha na tabela comptime_session (soft delete - deleted_at)
- Para cada sessão atualizada para o mesmo customer + service, deverá ser computada a sessão no banco de horas:
    * Deve-se localizar comptime_session correspondente a aquela sessão
    * Deve-se calcular o valor a ser atualizada no comptime.consumed_minutes = session.session_minutes - comptime_session.consumed_minutes (se session.status == "cancelada_nao_cobrar" deve-se considerar session.session_minutes = 0)
    * Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
    * Deve-se atualizar o valor do session_minutes na colune comptime_session.consumed_minutes (se session.status == "cancelada_nao_cobrar" deve-se considerar session.session_minutes = 0)
    * Não esquecer de atualizar o updated_at de todas as tabelas
- Para cada invoice_item criada para um determinado customer + service, deverá ser computada a fatura no banco de horas:
    * Deve-se verificar se o comptime foi criado para aquele customer e para aquele service (se não deve-se criar com valores invoiced_quantity, consumed_minutes, balance = 0)
    * Também deve-se verificar se a invoice_item_id já não está na tabela comptime_invoice_item e com deleted_at = nil (para se evitar duplicidade)
    * comptime.customer_id e comptime.service_id deverão ser iguais ao customer_id da invoice pai e service_id da invoice_item
    * comptime.invoiced_minutes deverá ser somado de invoice_item.quantity * service.session_minutes
    * Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
    * Deve-se criar uma linha na tabela comptime_invoice_item para aquele comptime e para aquela sessão com:
        - comptime_id = comptime.id
        - invoice_item_id = invoice_item.id
        - month = ano-mes-1 do invoice.period_start
        - invoiced_minutes = invoice_item.quantity * service.session_minutes
 - Para cada invoice_item ou invoice pai deletadas para o mesmo customer + service, deverá ser computada a fatura no banco de horas:
    * Deve-se localizar comptime_invoice_item correspondente a aquela invoice_item ou invoice pai
    * Deve-se subtrair o comptime.invoiced_minutes pelo comptime_invoice_item.invoiced_minutes
    * Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
    * Deve-se apagar a linha na tabela comptime_invoice_item (soft delete - deleted_at)
 - Para cada invoice_item ou invoice pai atualizada para o mesmo customer + service, deverá ser computada a fatura no banco de horas:
    * Deve-se localizar comptime_invoice_item correspondente a aquela invoice_item
    * Deve-se calcular o valor a ser atualizada no comptime.invoiced_minutes = invoice_item.quantity * service.session_minutes - comptime_invoice_item.invoiced_minutes 
    * Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
    * Deve-se atualizar o valor da coluna comptime_invoice_item.invoiced_minutes com o valor de invoice_item.quantity * service.session_minutes
    * Não esquecer de atualizar o updated_at de todas as tabelas
- Para cada comptime extra criada:
    - Deve-se somar o valor de comptime_extra.invoiced_minutes ao comptime.invoiced_minutes na tabela comptime
    - Deve-se calcular e ajustar o novo valor de comptime.balance = invoiced_minutes - consumed_minutes 
- Vale a mesma regra do comptime extra em deleção e update
- Ainda para as as atualizações acima:
    - Deve-se atualizar o balance de todos os registros da tabela comptime de meses subsequentes do mesmo customer_id e service_id com balance do registro atual + invoiced_minutes subsequente - consumed_minutes subsequente (isto para que o saldo seja propagado corretamente e não se precise fazer somas de registros quando se consultar o saldo de um cliente e um serviço em um determinado mês)
- Dado que a atualização da banco de horas envolve mais de uma linha e com atualização mais complexa por conta da atualização de saldo e a concorrência pode gerar problemas no calculo do saldo, seria interessante que esta atualização fosse feita através de um processo único e a entrada dos dados fosse feita através de uma fila ou alhgo assim e o envio dados ao se criar/deletar/atualizar sessões e invoices(invoice_item) seja feita de forma assincrona atraves de uma fila. Prever o melhar arquitetura para esta sincronização de dados com a garantia do processamento.
- Deve-se fazer um serviço e em forma de job que passa-se o vendor_id, data inicial e data final, com possibilidade de se passar o tipo de dados (session, invoice, todos) e o serviço ressincronisa sessões e invoices com o comptime;
- Deve-se fazer um documento descrevendo esta arquitetura em planner/docs/arquitetura_banco_horas.md
- Deve-se criar o domínio, dtos, eliminando o que antes era reconciliation.go e criar o comptime.go
- Deve-se eliminar o serviço anterior reconciliation.go e criar o novo serviço como comptime.go;
- No htmx de sessoes, retirar as ações referentes a concilição
- No htmx de invoices retirar as ações referentes a conciliação e ajustar para haja apenas o valor quantity
- Fazer htmx no modelo dos demais dominios para apresentar o banco de horas baseado no htmx de sessoes e pre-faturas com os seguintes requisitos:
    - Prever a listagem de banco de horas com os seguintes dados: mês, cliente, serviço, minutos faturados, minutos consumidos e saldo
    - Prever o seguintes filtros: mes inicial, mes final, cliente, serviço
    - Ordenar pelo mes (decrescente) e pelo cliente (crescente)
    - Prever o vendor default = 1
    - Prever um botão (modal, etc) para resoncronizar os dados chamando o serviço descrito acima com parâmetros data inicial e data final (vendor_id = 1)
- Fazer testes automatizados para validar as regras de negócio acima

    

     
    








                    