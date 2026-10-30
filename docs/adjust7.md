No subprojeto planner, nos serviços de hora extra (comptime):
- Alterei a tabela comptime para conter um valor de ajuste (adjusted_minutes) e também criei a tabela especializada para adjustes manuais de horas (comptime_adjust)...ver em /planner/migration/comtime.sql e no banco
- Ajustar as funcionalidades do sistema planner para comportar esta nova funcionalidade da seguinte forma:
    * Criar um serviço para inclusão ou alteração de um ajuste manual de horas para um determinado mês/customer/serviço
        ** o valor ajustado pode ser positivo ou negativo (adjusted minutes)
        ** deve-se ter uma justificativa para tal ajuste (description)
- Deve-se ter apenas uma linha de ajuste por customer/serviço/mês....pode-se inserir, alterar ou deletar tal ajuste
- Quando o ajuste é incluido/alterado/deletado, deve-se caalcular o ajuste no saldo do mês, isto é, balance = invoiced_minutes - consumed_minutes + adjusted_minutes e deve-se propagar o saldo ajustando todas outras linhas de meses posteriores ao mês atual do mesmo customer e serviço, da mesma forma que é feito para as alterações no campos invoiced_minutes e consumed_minutes.
- Fazer as apis conforme os serviços
- No htmx de comptime:
    * Colocar a coluna de Ajustados ao lado do realizados em cor específica.
    * Colocar um mouse over sobre os dados dos Ajustados, onde é apresentada a descrição, caso tenha
    * Colocar uma ação de Ajustar, onde através de um modal será incluido, ajustado ou deletado/zerado horas e espcificando uma justificativa/descrição
    * Retirar a ação de deletar
- Fazer as alterações seguindo os padrões do projeto, arquitetura hexagonal e bosas praticas de codigo e arquitetura limpas
