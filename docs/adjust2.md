No suprojeto planner:
- Ao se criar uma sessão para um cliente em um determinado serviço, caso exista para este cliente um plano na estratégia pay-per-view cuja data da sessão esteja dentro do periodo do plano e o serviço seja o mesmo do plano, deve ser criada uma pré-fatura de acordo com os parametros:
    * cliente: o mesmo da sessão
    * data da invoice: data da sessão
    * data de vencimento: data da sessão mais prazo em dia configurado em plan_payperview
    * data_inicio: data da sessão
    * data_fim: data da sessão
    * amount: valor do item do plano ou valor do pacote
    * service_id: o mesmo do plano
    * description: a descrição que está em service + '- dd/mm/aaaa' (data do plano)
    * quantity: 1
- Verificar se é melhor fazer de forma assincrona e caso seja fazendo garantindo o registro adequado.
- Ao se criar a pré-fatura, deve-se acionar o registro do banco de horas como é feito normalmente agora
'