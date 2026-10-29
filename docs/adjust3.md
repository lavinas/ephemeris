No subprojeto Planner (gestão de planos):
- Para os planos, criar a estratégia "Manual" a qual não terá especialização e nem itens, apenas o cliente, data inicio e data fim  e pode ter o valor do pacote;
- Esta estratégia será utilizada para a criação de pré-invoices avulsas e para clientes via criação de pré-fatura diretamente (pela funcionalidade de criação e pela importação do csv).
- Desta forma, todas as pré-faturas que são criadas diretamente na funcionalidade de criação ou importação de csv devem estar dentro desta estratégia. 
- Esta estratégia não toma nenhuma ação na funcionalidade gerar dos planos, é passiva.
- Na criação da pré-fatura diretamente, caso o cliente já tenha um um plano manual cujo período envolve a data da fatura, utilizar este plano, caso contrário, criar um plano desta modalidade para o cliente....no caso da criação de uma pré-fatura deve-se utilizar a data inicio e fim do plano a data da fatura, caso seja a importação do csv, deve-se criar um plano desta estratégia para o cliente cuja da datainicio é a data da primeira data  da fatura do csv e a data fim a data da ultima data da fatura do csv.
