No subprojeto planner:
- Na funcionalidade de listagem do Banco de Horas (comptime), gostaria de trocar os conceitos de Minutos Faturados (InvoicedMinutes) e Minutos Consumidos (ConsumedMinutes) por Minutos Previstos (PredictMinutes) e Minutos Realizados (DoneMinutes):
    * Para os meses que anteriores ao mês corrente, os cálculos continham o mesmo, isto é o que foi faturado x o que foi consumidos por sessões.
    * Para o mes corrente, o calculo deve ser baseado no plano ativo para o cliente e para o serviço (ativo é cuja data atual é maior que a data de inicio e menor que a data de fim do plano ou nula);
    * Caso nenhum plano seja encontrado para o cliente e para o serviço, o previsto deve ser 0.
    * Para o caso do plano encontrado seja das estratégias de caderneta e payperview, então o valor previsto é o mesmo que o realizado (ou seja o consumido);
    * Para o caso do plano encontrado seja de pacote, então o previsto deve ser proporcional a quantidade de dias desde o início até hoje, ou seja: (data atual - data inicio) / (data fim - data inicio) * (quantidade de sessões do plano * minutos do serviço), arredondado pra inteiro
    * Para o caso do plano encontrado seja do tipo assinatura:
        ** Calcular o numero de sessões previstas para o mes corrente baseado nos atributos do plano (periodicidade, inicio do plano, quantidade limite mensal, dia da semana).....vide ou utilizar planner/domain/invoicing/agenda/CalculateSessionsCount
        ** Calcular a quantidade de minutos proporcionais utilizando a mesma formula do plano pacote, porém agora com o numero de sessões resgatadas do subintem anterior * minutos do serviço.
- Ainda para esta funcionalidade trazer também o números de sessões correspondentes aos mintudos retornados em previstos, realizados, saldo mês e saldo total, dividindo-se os minutos calculados  em cada linha pelo número de minutos do serviço correspondente.
- Ajustar o htmx para mostrar Previsto e Realizado ao invés de Fatirado e Consumido, e também embaixo dos minutos previstos, realizados, saldo mes e saldo total apresentar o número de sessões fracionada (9.99) em caractere um pouco menor.
- Manter as premissas de arquitetura e padrões de projeto existentes no subprojeto, bem como bosa praticas de codigo e arquitura hexagonal limpa, tentando ao maximo reaproveitar o conceito de dominio/invoicing para os calculos


Para esta ultima atualização, existem alguns erros conceituais que foram colocados e precisam ser corrigidos: 
- Para calculo dos valores previstos do mês corrente tanto da estratégia de pacotes e de agenda, calcular primeiro o número de sessões estimadas, arredondando para inteiro e depois calculando o número de minutos atraves da multiplicação do numero de sessões estimadas vezes o número de minutos